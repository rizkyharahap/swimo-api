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
		errors["email"] = "Email is not a valid format"
	}

	if r.Password == "" {
		errors["password"] = "Password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if r.ConfirmPassword == "" {
		errors["confirmPassword"] = "Confirm password is required"
	}
	if r.Password != r.ConfirmPassword {
		errors["confirmPassword"] = "Confirm passwords do not match"
	}

	if strings.TrimSpace(r.Name) == "" {
		errors["name"] = "Name is required"
	}

	if r.Weight == nil {
		errors["weight"] = "Weight is required"
	} else if *r.Weight < 0 {
		errors["weight"] = "Weight cannot be negative"
	}

	if r.Height == nil {
		errors["height"] = "Height is required"
	} else if *r.Height < 0 {
		errors["height"] = "Height cannot be negative"
	}

	if r.Age == nil {
		errors["age"] = "Age is required"
	} else if *r.Age < 0 {
		errors["age"] = "Age cannot be negative"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
