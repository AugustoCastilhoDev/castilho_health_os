package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/castilho/health-os/internal/config"
	"github.com/castilho/health-os/internal/infra/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gdb, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, err := gdb.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "down",
			})
		}
		return c.JSON(fiber.Map{"status": "up"})
	})

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
