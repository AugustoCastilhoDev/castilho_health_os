package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/castilho/health-os/internal/api/handlers"
	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/config"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/infra/db"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
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

	issuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.JWTTTL)

	tenantRepo := repository.NewTenantRepository(gdb)
	userRepo := repository.NewUserRepository(gdb)

	authService := service.NewAuthService(tenantRepo, userRepo, issuer)
	authHandler := handlers.NewAuthHandler(authService)

	app := fiber.New()
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, err := gdb.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "down"})
		}
		return c.JSON(fiber.Map{"status": "up"})
	})

	app.Post("/auth/login", authHandler.Login)

	api := app.Group("/api", appmiddleware.RequireAuth(issuer))

	api.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":   appmiddleware.UserID(c),
			"tenant_id": appmiddleware.TenantID(c),
			"role":      appmiddleware.Role(c),
		})
	})

	api.Get("/admin/ping", appmiddleware.RequireRole(models.RoleTenantAdmin), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
