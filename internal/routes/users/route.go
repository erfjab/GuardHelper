package users

import (
	"guardhelper/internal/middlewares"
	"github.com/gofiber/fiber/v2"
)

func RegisterServiceRoutes(app *fiber.App) {
	serviceGroup := app.Group("/api/users") 

	serviceGroup.Get("/", middlewares.AuthApiKey, GetAllUsers)
}
