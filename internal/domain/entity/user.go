package entity

import "time"

type User struct {
	ID        string
	AccountID string
	Name      string
	WeightKG  *float64
	HeightCM  *float64
	AgeYears  *int16
	CreatedAt time.Time
	UpdatedAt time.Time
}
