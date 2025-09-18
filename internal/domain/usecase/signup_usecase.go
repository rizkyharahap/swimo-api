package usecase

import (
	"context"
	"errors"
	"haphap/swimo-api/internal/domain/repository"
	"log/slog"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail     = errors.New("invalid email")
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordNotMatch = errors.New("password confirm mismatch")
)

type SignUpInput struct {
	Email           string
	Password        string
	ConfirmPassword string
	Name            string
	WeightKG        *float64
	HeightCM        *float64
	AgeYears        *int16
}

type SignUpUsecase interface {
	Execute(ctx context.Context, in SignUpInput) error
}

type signUpUC struct {
	pool     *pgxpool.Pool
	accounts repository.AccountRepository
	users    repository.UserRepository
}

func NewSignUpUsecase(pool *pgxpool.Pool, a repository.AccountRepository, u repository.UserRepository) SignUpUsecase {
	return &signUpUC{pool: pool, accounts: a, users: u}
}

var emailRx = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func (uc *signUpUC) Execute(ctx context.Context, in SignUpInput) error {
	// validate
	if !emailRx.MatchString(in.Email) {
		return ErrInvalidEmail
	}

	if len(in.Password) < 8 {
		return ErrPasswordTooShort
	}

	if in.Password != in.ConfirmPassword {
		return ErrPasswordNotMatch
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Transaction Start
	tx, err := uc.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// rollback safety
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Create account
	accID, err := uc.accounts.Create(ctx, tx, in.Email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrAccountExists) {
			return repository.ErrAccountExists
		}
		return err
	}

	// Create user profile
	_, err = uc.users.Create(ctx, tx, accID, in.Name, in.WeightKG, in.HeightCM, in.AgeYears)
	if err != nil {
		slog.Warn("signup: create user failed, rolling back", slog.String("account_id", accID), slog.String("err", err.Error()))
		return err // tx rollback by defer
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil // avoid defered rollback

	slog.Info("signup success", slog.String("email", in.Email))
	return nil
}
