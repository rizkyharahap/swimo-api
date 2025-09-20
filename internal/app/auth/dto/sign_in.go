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
	validationErrors := make(map[string]string)

	sanitizedEmail := strings.TrimSpace(strings.ToLower(r.Email))
	if !validator.EmailPattern.MatchString(sanitizedEmail) {
		validationErrors["email"] = "email is not a valid format"
	}

	if len(r.Password) < 8 {
		validationErrors["password"] = "password must be at least 8 characters"
	}

	if len(validationErrors) > 0 {
		return &validator.ValidationError{Errors: validationErrors}
	}

	return nil
}
