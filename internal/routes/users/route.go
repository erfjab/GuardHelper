package users

import (
	"guardhelper/internal/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	userGroup := app.Group("/api/users") 

	userGroup.Get("/", middlewares.AuthApiKey, GetAllUsers)
}
