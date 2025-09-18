package dto

import "fmt"

type SignUpRequest struct {
	Email           string   `json:"email"`
	Password        string   `json:"password"`
	ConfirmPassword string   `json:"confirmPassword"`
	Name            string   `json:"name"`
	Weight          *float64 `json:"weight"`
	Height          *float64 `json:"height"`
	Age             *int16   `json:"age"`
}

func (r *SignUpRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	if r.ConfirmPassword == "" {
		return fmt.Errorf("confirmPassword is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Weight == nil {
		return fmt.Errorf("weight is required")
	}
	if r.Height == nil {
		return fmt.Errorf("height is required")
	}
	if r.Age == nil {
		return fmt.Errorf("age is required")
	}
	return nil
}
