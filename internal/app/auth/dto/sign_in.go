package dto

import (
	"haphap/swimo-api/pkg/validator"
	"strings"
)

type (
	SignInRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		UserAgent *string
	}

	SignInResponse struct {
		Name         string   `json:"name"`
		Weight       *float64 `json:"weight"`
		Height       *float64 `json:"height"`
		Age          *int16   `json:"age"`
		Email        string   `json:"email"`
		Token        string   `json:"token"`
		RefreshToken string   `json:"refreshToken"`
		ExpiresInMs  int64    `json:"expiresIn"`
	}

	SignInGuestResponse struct {
		Weight       *float64 `json:"weight"`
		Height       *float64 `json:"height"`
		Age          *int16   `json:"age"`
		Token        string   `json:"token"`
		RefreshToken string   `json:"refreshToken"`
		ExpiresInMs  int64    `json:"expiresIn"`
	}
)

func (r *SignInRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	sanitizedEmail := strings.TrimSpace(strings.ToLower(r.Email))
	if sanitizedEmail == "" {
		errors["email"] = "Email is required"
	} else if !validator.EmailPattern.MatchString(sanitizedEmail) {
		errors["email"] = "Email is not a valid format"
	}

	if r.Password == "" {
		errors["password"] = "Password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
