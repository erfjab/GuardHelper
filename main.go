package main

import (
	"guardhelper/internal/config"
	usersRoutes "guardhelper/internal/routes/users"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	log.Printf("Starting GuardHelper...")
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded successfully.")
	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	fiberApp.Use(logger.New())
	usersRoutes.RegisterUserRoutes(fiberApp)
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "GuardHelper is running"})
	})
	fiberApp.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "not found",
		})
	})
	log.Printf("Listening on http://0.0.0.0:99")
	log.Fatal(fiberApp.Listen(":99"))
}
