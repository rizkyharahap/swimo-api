package dto

import (
	"haphap/swimo-api/internal/app/auth/entity"
	"haphap/swimo-api/pkg/validator"
	"strings"
)

type (
	SignUpRequest struct {
		Email           string   `json:"email"`
		Password        string   `json:"password"`
		ConfirmPassword string   `json:"confirmPassword"`
		Name            string   `json:"name"`
		Weight          *float64 `json:"weight"`
		Height          *float64 `json:"height"`
		Age             *int16   `json:"age"`
	}
)

func (r *SignUpRequest) ToUserEntity(accountID string) *entity.User {
	return &entity.User{
		AccountID: accountID,
		Name:      strings.TrimSpace(r.Name),
		WeightKG:  r.Weight,
		HeightCM:  r.Height,
		AgeYears:  r.Age,
	}
}

func (r *SignUpRequest) Validate() error {
	errors := make(map[string]string)

	sanitizedEmail := strings.TrimSpace(strings.ToLower(r.Email))
	if !validator.EmailPattern.MatchString(sanitizedEmail) {
		errors["email"] = "email is not a valid format"
	}

	if len(r.Password) < 8 {
		errors["password"] = "password must be at least 8 characters"
	}

	if r.Password != r.ConfirmPassword {
		errors["confirmPassword"] = "passwords do not match"
	}

	if strings.TrimSpace(r.Name) == "" {
		errors["name"] = "name is required"
	}

	if r.Weight == nil {
		errors["weight"] = "weight is required"
	} else if *r.Weight < 0 {
		errors["weight"] = "weight cannot be negative"
	}

	if r.Height == nil {
		errors["height"] = "height is required"
	} else if *r.Height < 0 {
		errors["height"] = "height cannot be negative"
	}

	if r.Age == nil {
		errors["age"] = "age is required"
	} else if *r.Age < 0 {
		errors["age"] = "age cannot be negative"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
