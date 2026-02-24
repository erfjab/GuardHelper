package main

import (
	"guardhelper/internal/config"
	"guardhelper/internal/database"
	usersRoutes "guardhelper/internal/routes/users"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/cast"
)

func main() {
	log.Printf("Starting GuardHelper [v0.1.1]")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded successfully.")

	database.ConnectDB()
	log.Printf("Database connected successfully.")

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
	address := cfg.ApiHost + ":" + cast.ToString(cfg.ApiPort)
	if cfg.ApiSslCertFile != "" && cfg.ApiSslKeyFile != "" {
		log.Printf("Listening on https://%s:%d", cfg.ApiHost, cfg.ApiPort)
		log.Fatal(fiberApp.ListenTLS(address, cfg.ApiSslCertFile, cfg.ApiSslKeyFile))
	} else {
		log.Printf("Listening on http://%s:%d", cfg.ApiHost, cfg.ApiPort)
		log.Fatal(fiberApp.Listen(address))
	}
}
