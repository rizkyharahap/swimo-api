package auth

import (
	"context"
	"errors"
	"haphap/swimo-api/internal/app/auth/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAccountExists = errors.New("account already exists")
)

type AuthRepository interface {
	GetAuthByEmail(ctx context.Context, email string) (*entity.Auth, error)
	CreateAccount(ctx context.Context, tx pgx.Tx, email, passwordHash string) (id string, err error)
	CreateUser(ctx context.Context, tx pgx.Tx, user *entity.User) (id string, err error)
	CreateUserSession(ctx context.Context, session *entity.Session) (id string, err error)
	CreateGuestSession(ctx context.Context, session *entity.Session) (id string, err error)
	CountRecentGuestByUA(ctx context.Context, ua *string, since *time.Time) (count int, err error)
}

type authRepository struct{ db *pgxpool.Pool }

func NewAuthRepository(db *pgxpool.Pool) AuthRepository { return &authRepository{db: db} }

func (r *authRepository) GetAuthByEmail(ctx context.Context, email string) (*entity.Auth, error) {
	const sql = `
		SELECT
		    a.id, a.email, a.password_hash, a.is_locked, 
			u.name, u.weight_kg, u.height_cm, u.age_years
		FROM accounts AS a
		JOIN users AS u ON a.id = u.account_id
		WHERE a.email = $1`

	var auth entity.Auth
	if err := r.db.QueryRow(ctx, sql, email).Scan(
		&auth.AccountID,
		&auth.Email,
		&auth.PasswordHash,
		&auth.IsLocked,
		&auth.Name,
		&auth.WeightKG,
		&auth.HeightCM,
		&auth.AgeYears,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrInvalidCreds
		}

		return nil, err
	}

	return &auth, nil
}

func (r *authRepository) CreateAccount(ctx context.Context, tx pgx.Tx, email, passwordHash string) (id string, err error) {
	const sql = `INSERT INTO accounts (email, password_hash) VALUES ($1, $2) RETURNING id`

	if err = tx.QueryRow(ctx, sql, email, passwordHash).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return "", ErrAccountExists
		}

		return "", err
	}

	return id, nil
}

func (r *authRepository) CreateUser(ctx context.Context, tx pgx.Tx, user *entity.User) (id string, err error) {
	const sql = `
		INSERT INTO users (account_id, name, weight_kg, height_cm, age_years)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id`

	if err = tx.QueryRow(ctx, sql, &user.AccountID, &user.Name, &user.WeightKG, &user.HeightCM, &user.AgeYears).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *authRepository) CreateUserSession(ctx context.Context, session *entity.Session) (id string, err error) {
	const sql = `
		INSERT INTO sessions (account_id, kind, user_agent, expires_at, refresh_token_hash, refresh_expires_at)
		VALUES ($1, 'user', $2, $3, $4, $5)
		RETURNING id`

	if err = r.db.QueryRow(ctx, sql, &session.AccountID, &session.UserAgent, &session.ExpiresAt, &session.RefreshTokenHash, &session.RefreshExpiresAt).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (r *authRepository) CreateGuestSession(ctx context.Context, session *entity.Session) (id string, err error) {
	const sql = `
		INSERT INTO SESSIONS (account_id, kind, user_agent, expires_at, refresh_token_hash, refresh_expires_at)
		VALUES (NULL, 'guest', $1, $2, $3, $4)
		RETURNING id`

	if err = r.db.QueryRow(ctx, sql, &session.UserAgent, &session.ExpiresAt, &session.RefreshTokenHash, &session.RefreshExpiresAt).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func (r *authRepository) CountRecentGuestByUA(ctx context.Context, ua *string, since *time.Time) (count int, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE kind='guest' AND user_agent = $1 AND created_at >= $2`, &ua, &since).Scan(&count)

	return count, err
}
