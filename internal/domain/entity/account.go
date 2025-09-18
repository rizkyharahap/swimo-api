package entity

import "time"

type Account struct {
	ID           string
	Email        string
	PasswordHash string
	IsLocked     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
