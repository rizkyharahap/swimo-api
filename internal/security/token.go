package security

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Kind      string `json:"kind"`
	SessionID string `json:"sid"`
	Sub       string `json:"sub"` // account_id
	jwt.RegisteredClaims
}

func NewAccessToken(secret string, accountID, sessionID string, ttl time.Duration) (token string, exp time.Time, err error) {
	now := time.Now()
	exp = now.Add(ttl)
	claims := Claims{
		Kind:      "user",
		SessionID: sessionID,
		Sub:       accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(secret))
	return
}

func NewOpaqueRefreshToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
