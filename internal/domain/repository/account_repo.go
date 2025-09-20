package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAccountExists = errors.New("account already exists")
)

type AccountRepository interface {
	GetByEmail(ctx context.Context, email string) (string, string, bool, error)
	Create(ctx context.Context, tx pgx.Tx, email, passwordHash string) (string, error) // returns account_id
}

type accountRepo struct{ db *pgxpool.Pool }

func NewAccountRepository(db *pgxpool.Pool) AccountRepository { return &accountRepo{db: db} }

func (r *accountRepo) GetByEmail(ctx context.Context, email string) (string, string, bool, error) {
	const sql = `SELECT id, password_hash, is_locked FROM accounts WHERE email=$1`
	var id, hash string
	var locked bool
	err := r.db.QueryRow(ctx, sql, email).Scan(&id, &hash, &locked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", false, nil
		}
		return "", "", false, err
	}
	return id, hash, locked, nil
}

func (r *accountRepo) Create(ctx context.Context, tx pgx.Tx, email, passwordHash string) (string, error) {
	const sql = `INSERT INTO accounts (email, password_hash) VALUES ($1, $2) RETURNING id`
	var id string
	err := tx.QueryRow(ctx, sql, email, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return "", ErrAccountExists
		}
		return "", err
	}

	return id, nil
}
