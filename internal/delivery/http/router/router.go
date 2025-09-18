package router

import (
	"haphap/swimo-api/internal/delivery/http/handler"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, auth *handler.AuthHandler) {
	apiV1 := app.Group("/api/v1")
	apiV1.Post("/sign-in", auth.SignIn)
	apiV1.Post("/sign-up", auth.SignUp)
}
