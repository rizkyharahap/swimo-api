package http

import (
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, authHandler *AuthHandler) {
	apiV1 := app.Group("/api/v1")
	apiV1.Post("/sign-in", authHandler.SignIn)
	apiV1.Post("/sign-in-guest", authHandler.SignInGuest)
	apiV1.Post("/sign-up", authHandler.SignUp)
}
