package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository interface {
	CreateUserSession(ctx context.Context, accountID string, refreshHas *string, expiresAt, refreshExpiresAt *time.Time, ua *string) (string, error)
}

type sessionRepo struct{ db *pgxpool.Pool }

func NewSessionRepository(db *pgxpool.Pool) SessionRepository { return &sessionRepo{db: db} }

func (r *sessionRepo) CreateUserSession(ctx context.Context, accountID string, refreshHas *string, expiresAt, refreshExpiresAt *time.Time, ua *string) (string, error) {
	const sql = `
		INSERT INTO sessions (account_id, kind, user_agent, expires_at, refresh_token_hash, refresh_expires_at)
		VALUES ($1, 'user', $2, $3, $4, $5)
		RETURNING id`

	var id string
	if err := r.db.QueryRow(ctx, sql, accountID, ua, expiresAt, refreshHas, refreshExpiresAt).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}
