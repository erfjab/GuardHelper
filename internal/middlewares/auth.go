package middlewares

import (
	"guardhelper/internal/config"
	"github.com/gofiber/fiber/v2"
)

func AuthApiKey(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-KEY")
	if apiKey != config.Cfg.ApiKey {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	return c.Next()
}