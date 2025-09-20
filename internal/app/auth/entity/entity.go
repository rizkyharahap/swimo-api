package entity

import (
	"errors"
	"haphap/swimo-api/config"
	"haphap/swimo-api/pkg/security"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type (
	User struct {
		ID        string
		AccountID string
		Name      string
		WeightKG  *float64
		HeightCM  *float64
		AgeYears  *int16
	}

	Auth struct {
		AccountID    string
		Email        string
		PasswordHash string
		IsLocked     bool
		Name         string
		WeightKG     *float64
		HeightCM     *float64
		AgeYears     *int16
	}

	Session struct {
		AccountID        *string
		RefreshTokenHash *string
		ExpiresAt        *time.Time
		RefreshExpiresAt *time.Time
		UserAgent        *string
	}
)

func (u *Auth) ComparePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCreds
	}

	return nil
}

func NewSession(cfg *config.Config, userAgent *string, accountId *string) (*Session, error) {

	refreshToken, err := security.NewOpaqueRefreshToken(32)
	if err != nil {
		return nil, err
	}
	refreshTokenHash := security.SHA256Hex(refreshToken)

	now := time.Now()
	expiresAt := now.Add(cfg.Auth.JWTAccessTTL)
	refreshExp := now.Add(cfg.Auth.JWTRefreshTTL)

	return &Session{
		AccountID:        accountId,
		UserAgent:        userAgent,
		ExpiresAt:        &expiresAt,
		RefreshTokenHash: &refreshTokenHash,
		RefreshExpiresAt: &refreshExp,
	}, nil
}
