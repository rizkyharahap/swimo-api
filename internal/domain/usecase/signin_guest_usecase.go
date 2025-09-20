package usecase

import (
	"context"
	"errors"
	"haphap/swimo-api/config"
	"haphap/swimo-api/internal/domain/repository"
	"haphap/swimo-api/internal/security"
	"time"
)

var (
	ErrGuestDisabled = errors.New("guest sign in disabled")
	ErrGuestLimited  = errors.New("guest_signin_rate_limited")
)

type SignInGuestInput struct {
	UserAgent *string
}

type SignInGuestOutput struct {
	Weight       *float64 `json:"weight"`
	Height       *float64 `json:"height"`
	Age          *int16   `json:"age"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresInMs  int64    `json:"expiresIn"`
}

type SignInGuestUseCase interface {
	Execute(ctx context.Context, in SignInGuestInput) (*SignInGuestOutput, error)
}

type signInGuestUc struct {
	cfg      *config.Config
	sessions repository.SessionRepository
}

func NewSignInGuestUseCase(cfg *config.Config, sessionRepo repository.SessionRepository) SignInGuestUseCase {
	return &signInGuestUc{cfg: cfg, sessions: sessionRepo}
}

func (uc *signInGuestUc) Execute(ctx context.Context, in SignInGuestInput) (*SignInGuestOutput, error) {
	if !uc.cfg.Auth.GuestEnabled {
		return nil, ErrGuestDisabled
	}

	// reate limit by UA/second
	if uc.cfg.Auth.GuestRatePerMinute > 0 {
		since := time.Now().UTC().Add(-1 * time.Minute)
		if cnt, err := uc.sessions.CountRecentGuestByUA(ctx, in.UserAgent, &since); err == nil && cnt >= uc.cfg.Auth.GuestRatePerMinute {
			return nil, ErrGuestLimited
		}
	}

	// create opaque refresh token (plain -> response, hash -> DB)
	refresh, err := security.NewOpaqueRefreshToken(32) // 64 hex chars
	if err != nil {
		return nil, err
	}
	refreshHash := security.SHA256Hex(refresh)

	now := time.Now()
	expiresAt := now.Add(uc.cfg.Auth.JWTAccessTTL)
	refreshExp := now.Add(uc.cfg.Auth.JWTRefreshTTL)

	// create guest session (account_id = NULL, kind='guest')
	sessID, err := uc.sessions.CreateGuestSession(ctx, &refreshHash, &expiresAt, &refreshExp, in.UserAgent)
	if err != nil {
		return nil, err
	}

	// issue access token (kind='guest', sub/account_id empty, sid=sessID)
	access, exp, err := security.NewAccessToken(uc.cfg.Auth.JWTSecret, "guest", "", sessID, uc.cfg.Auth.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &SignInGuestOutput{
		Weight:       nil,
		Height:       nil,
		Age:          nil,
		Token:        access,
		RefreshToken: refresh,
		ExpiresInMs:  exp.Sub(now).Milliseconds(),
	}, nil
}
