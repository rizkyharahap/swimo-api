package http

import (
	"errors"
	"haphap/swimo-api/internal/app/auth"
	"haphap/swimo-api/internal/app/auth/dto"
	"haphap/swimo-api/internal/app/auth/entity"
	"haphap/swimo-api/pkg/response"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUsecase auth.AuthUseCase
}

func NewAuthHandler(authUsecase auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUsecase}
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var req dto.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("signup parse error", slog.String("err", err.Error()))
		return c.Status(http.StatusBadRequest).JSON(response.BaseResponse{Message: "Invalid JSON body."})
	}

	// validate required fields
	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(err)
	}

	err := h.authUsecase.SignUp(c.Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrAccountExists) {
			return c.Status(http.StatusConflict).JSON(response.BaseResponse{Message: "Email already exists."})
		}

		return err
	}

	return c.Status(http.StatusCreated).JSON(response.BaseResponse{Message: "User registered successfully."})
}

func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
	var req dto.SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.BaseResponse{Message: "Invalid JSON body."})
	}

	// validate required fields
	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(err)
	}

	ua := string(c.Request().Header.UserAgent())
	req.UserAgent = &ua

	out, err := h.authUsecase.SignIn(c.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCreds):
			return c.Status(http.StatusUnprocessableEntity).JSON(response.BaseResponse{Message: "Invalid Email or Passwords."})
		case errors.Is(err, auth.ErrLocked):
			return c.Status(http.StatusForbidden).JSON(response.BaseResponse{Message: "Your account has been locked."})
		default:
			return err
		}
	}

	return c.Status(http.StatusOK).JSON(response.BaseResponse{
		Data:    out,
		Message: "Sign-in successfull.",
	})
}

func (h AuthHandler) SignInGuest(c *fiber.Ctx) error {
	userAgent := string(c.Request().Header.UserAgent())

	out, err := h.authUsecase.SignInGuest(c.Context(), dto.SignInRequest{UserAgent: &userAgent})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrGuestDisabled):
			return c.Status(http.StatusForbidden).JSON(response.BaseResponse{Message: "Guest sign-in is currently disabled. Please create an account."})
		case errors.Is(err, auth.ErrGuestLimited):
			return c.Status(http.StatusTooManyRequests).JSON(response.BaseResponse{Message: "Guest session limit reached. Please try again later."})
		default:
			return err
		}
	}

	return c.Status(http.StatusOK).JSON(response.BaseResponse{
		Data:    out,
		Message: "Guest sign-in successful.",
	})
}
