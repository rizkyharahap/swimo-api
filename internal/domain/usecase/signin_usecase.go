package usecase

import (
	"context"
	"errors"
	"haphap/swimo-api/internal/config"
	"haphap/swimo-api/internal/domain/repository"
	"haphap/swimo-api/internal/security"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound     = errors.New("user not found")
	ErrLocked       = errors.New("account locked")
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type SignInInput struct {
	Email     string
	Password  string
	UserAgent *string
}

type SignInOutput struct {
	Name         string   `json:"name"`
	Weight       *float64 `json:"weight"`
	Height       *float64 `json:"height"`
	Age          *int16   `json:"age"`
	Email        string   `json:"email"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresInMs  int64    `json:"expiresIn"`
}

type SignInUseCase interface {
	Execute(ctx context.Context, in SignInInput) (*SignInOutput, error)
}

type signInUc struct {
	cfg      *config.Config
	accounts repository.AccountRepository
	users    repository.UserRepository
	sessions repository.SessionRepository
}

func NewSignInUseCase(cfg *config.Config, accountRepo repository.AccountRepository, userRepo repository.UserRepository, sessionRepo repository.SessionRepository) SignInUseCase {
	return &signInUc{cfg: cfg, accounts: accountRepo, users: userRepo, sessions: sessionRepo}
}

func (uc *signInUc) Execute(ctx context.Context, in SignInInput) (*SignInOutput, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	accID, hash, locked, err := uc.accounts.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if accID == "" {
		return nil, ErrNotFound
	}
	if locked {
		return nil, ErrLocked
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		return nil, ErrInvalidCreds
	}

	// create session with refresh
	refresh, err := security.NewOpaqueRefreshToken(32) // 64 hex
	if err != nil {
		return nil, err
	}
	refreshHash := security.SHA256Hex(refresh)

	now := time.Now()
	expiresAt := now.Add(uc.cfg.JWTAccessTTL)
	refreshExp := now.Add(uc.cfg.JWTRefreshTTL)

	sessID, err := uc.sessions.CreateUserSession(ctx, accID, &refreshHash, &expiresAt, &refreshExp, in.UserAgent)
	if err != nil {
		return nil, err
	}

	// access token
	access, exp, err := security.NewAccessToken(uc.cfg.JWTSecret, "user", accID, sessID, uc.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	// profile
	up, err := uc.users.GetByAccountID(ctx, accID)
	if err != nil {
		return nil, err
	}

	return &SignInOutput{
		Name:         up.Name,
		Weight:       up.WeightKG,
		Height:       up.HeightCM,
		Age:          up.AgeYears,
		Email:        up.Email,
		Token:        access,
		RefreshToken: refresh,
		ExpiresInMs:  exp.Sub(now).Milliseconds(),
	}, nil
}
