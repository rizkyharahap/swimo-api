package auth

import (
	"context"
	"errors"
	"haphap/swimo-api/config"
	"haphap/swimo-api/internal/app/auth/dto"
	"haphap/swimo-api/internal/app/auth/entity"
	"haphap/swimo-api/pkg/security"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrGuestDisabled = errors.New("guest sign in disabled")
	ErrGuestLimited  = errors.New("guest sign in rate limited")
	ErrLocked        = errors.New("account locked")
)

type AuthUseCase interface {
	SignUp(ctx context.Context, req dto.SignUpRequest) error
	SignIn(ctx context.Context, req dto.SignInRequest) (*dto.SignInResponse, error)
	SignInGuest(ctx context.Context, req dto.SignInRequest) (*dto.SignInGuestResponse, error)
}

type authUseCase struct {
	cfg      *config.Config
	pool     *pgxpool.Pool
	authRepo AuthRepository
}

func NewAuthUseCase(cfg *config.Config, pool *pgxpool.Pool, authRepo AuthRepository) AuthUseCase {
	return &authUseCase{cfg, pool, authRepo}
}

func (uc *authUseCase) SignUp(ctx context.Context, req dto.SignUpRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Transaction Start
	tx, err := uc.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create account
	email := strings.TrimSpace(strings.ToLower(req.Email))

	accountID, err := uc.authRepo.CreateAccount(ctx, tx, email, string(hash))
	if err != nil {
		slog.Warn("signup: create account failed, rolling back", slog.String("email", email), slog.String("err", err.Error()))
		return err
	}

	// Create user profile
	user := req.ToUserEntity(accountID)

	_, err = uc.authRepo.CreateUser(ctx, tx, user)
	if err != nil {
		slog.Warn("signup: create user failed, rolling back", slog.String("account_id", accountID), slog.String("err", err.Error()))
		return err // tx rollback by defer
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		slog.Error("signup: commit transaction failed", slog.String("email", email), slog.String("err", err.Error()))
		return err
	}

	slog.Info("signup success", slog.String("email", email))
	return nil
}

func (uc *authUseCase) SignIn(ctx context.Context, req dto.SignInRequest) (*dto.SignInResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	auth, err := uc.authRepo.GetAuthByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if auth.IsLocked {
		return nil, ErrLocked
	}

	if err = auth.ComparePassword(req.Password); err != nil {
		return nil, err
	}

	// Create session with refresh token
	session, err := entity.NewSession(uc.cfg, req.UserAgent, &auth.AccountID)
	if err != nil {
		return nil, err
	}

	sessionId, err := uc.authRepo.CreateUserSession(ctx, session)
	if err != nil {
		return nil, err
	}

	accessToken, exp, err := security.NewAccessToken(uc.cfg.Auth.JWTSecret, "user", auth.AccountID, sessionId, uc.cfg.Auth.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &dto.SignInResponse{
		Name:         auth.Name,
		Weight:       auth.WeightKG,
		Height:       auth.HeightCM,
		Age:          auth.AgeYears,
		Email:        auth.Email,
		Token:        accessToken,
		RefreshToken: *session.RefreshTokenHash,
		ExpiresInMs:  time.Until(exp).Milliseconds(),
	}, nil
}

func (uc *authUseCase) SignInGuest(ctx context.Context, req dto.SignInRequest) (*dto.SignInGuestResponse, error) {
	if !uc.cfg.Auth.GuestEnabled {
		return nil, ErrGuestDisabled
	}

	if uc.cfg.Auth.GuestRatePerMinute > 0 {
		since := time.Now().UTC().Add(-1 * time.Minute)
		if cnt, err := uc.authRepo.CountRecentGuestByUA(ctx, req.UserAgent, &since); err == nil && cnt >= uc.cfg.Auth.GuestRatePerMinute {
			return nil, ErrGuestLimited
		}
	}

	// Create session with refresh token
	session, err := entity.NewSession(uc.cfg, req.UserAgent, nil)
	if err != nil {
		return nil, err
	}

	sessionId, err := uc.authRepo.CreateGuestSession(ctx, session)
	if err != nil {
		return nil, err
	}

	access, exp, err := security.NewAccessToken(uc.cfg.Auth.JWTSecret, "guest", "", sessionId, uc.cfg.Auth.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &dto.SignInGuestResponse{
		Weight:       nil,
		Height:       nil,
		Age:          nil,
		Token:        access,
		RefreshToken: *session.RefreshTokenHash,
		ExpiresInMs:  time.Until(exp).Milliseconds(),
	}, nil
}
