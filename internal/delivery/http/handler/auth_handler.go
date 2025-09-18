package handler

import (
	"errors"
	"haphap/swimo-api/internal/delivery/http/dto"
	"haphap/swimo-api/internal/domain/repository"
	"haphap/swimo-api/internal/domain/usecase"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	signin usecase.SignInUseCase
	signup usecase.SignUpUsecase
}

func NewAuthHandler(signin usecase.SignInUseCase, signup usecase.SignUpUsecase) *AuthHandler {
	return &AuthHandler{signin: signin, signup: signup}
}

func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
	var req dto.SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.MessageResponse{Message: "Invalid JSON body"})
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.MessageResponse{Message: "email and password are required"})
	}

	ua := string(c.Request().Header.UserAgent())

	out, err := h.signin.Execute(c.Context(), usecase.SignInInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: &ua,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			return c.Status(http.StatusNotFound).JSON(dto.MessageResponse{Message: "User account does not exists."})
		case errors.Is(err, usecase.ErrLocked):
			return c.Status(http.StatusForbidden).JSON(dto.MessageResponse{Message: "Your account has been locked."})
		case errors.Is(err, usecase.ErrInvalidCreds):
			return c.Status(http.StatusUnauthorized).JSON(dto.MessageResponse{Message: "Invalid email or password."})
		default:
			slog.Error("signin failed", slog.String("err", err.Error()))
			return c.Status(http.StatusInternalServerError).JSON(dto.MessageResponse{Message: "Internal Server Error"})
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"data":    out,
		"message": "Sign-in successfull.",
	})
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var req dto.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("signup parse error", slog.String("err", err.Error()))
		return c.Status(http.StatusBadRequest).JSON(dto.MessageResponse{Message: "Invalid JSON body"})
	}

	// validate required fields
	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(dto.MessageResponse{Message: err.Error()})
	}

	err := h.signup.Execute(c.Context(), usecase.SignUpInput{
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
		Name:            req.Name,
		WeightKG:        req.Weight,
		HeightCM:        req.Height,
		AgeYears:        req.Age,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidEmail):
			return c.Status(http.StatusBadRequest).JSON(dto.MessageResponse{Message: "Email format is invalid."})
		case errors.Is(err, usecase.ErrPasswordTooShort):
			return c.Status(http.StatusUnprocessableEntity).JSON(dto.MessageResponse{Message: "Password must be at least 8 characters long."})
		case errors.Is(err, usecase.ErrPasswordNotMatch):
			return c.Status(http.StatusUnprocessableEntity).JSON(dto.MessageResponse{Message: "Password and confirmPassword do not match."})
		case errors.Is(err, repository.ErrAccountExists):
			return c.Status(http.StatusConflict).JSON(dto.MessageResponse{Message: "Email already exists."})
		default:
			slog.Error("signup failed", slog.String("err", err.Error()))
			return c.Status(http.StatusInternalServerError).JSON(dto.MessageResponse{Message: "Internal Server Error"})
		}
	}

	return c.Status(http.StatusCreated).JSON(dto.MessageResponse{Message: "User registered successfully."})
}
