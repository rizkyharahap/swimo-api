// internal/domain/repository/user_repo.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
	Name     string
	WeightKG *float64
	HeightCM *float64
	AgeYears *int16
	Email    string
}

type UserRepository interface {
	GetByAccountID(ctx context.Context, accountID string) (*UserProfile, error)
	Create(ctx context.Context, q Querier, accountID, name string, weight, height *float64, age *int16) (string, error)
}

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository { return &userRepo{db: db} }

func (r *userRepo) GetByAccountID(ctx context.Context, accountID string) (*UserProfile, error) {
	const sql = `
		SELECT u.name, u.weight_kg, u.height_cm, u.age_years, a.email
		FROM users u JOIN accounts a ON a.id = u.account_id
		WHERE u.account_id = $1
	`
	var userProfile UserProfile
	if err := r.db.QueryRow(ctx, sql, accountID).Scan(
		&userProfile.Name,
		&userProfile.WeightKG,
		&userProfile.HeightCM,
		&userProfile.AgeYears,
		&userProfile.Email,
	); err != nil {
		return nil, err
	}
	return &userProfile, nil
}

func (r *userRepo) Create(ctx context.Context, q Querier, accountID, name string, weight, height *float64, age *int16) (string, error) {
	const sql = `
		INSERT INTO users (account_id, name, weight_kg, height_cm, age_years)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id`
	var id string
	if err := q.QueryRow(ctx, sql, accountID, name, weight, height, age).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}
