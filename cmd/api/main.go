package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/castilho/health-os/internal/api"
	"github.com/castilho/health-os/internal/api/handlers"
	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/config"
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
	patientRepo := repository.NewPatientRepository(gdb)
	appointmentRepo := repository.NewAppointmentRepository(gdb)
	ruleRepo := repository.NewFinancialRuleRepository(gdb)
	txRepo := repository.NewFinancialTransactionRepository(gdb)

	settlementService := service.NewSettlementService(appointmentRepo, ruleRepo, txRepo)

	h := &api.Handlers{
		Auth:        handlers.NewAuthHandler(service.NewAuthService(tenantRepo, userRepo, issuer)),
		Tenant:      handlers.NewTenantHandler(service.NewTenantService(gdb, tenantRepo)),
		User:        handlers.NewUserHandler(service.NewUserService(userRepo)),
		Patient:     handlers.NewPatientHandler(service.NewPatientService(patientRepo)),
		Appointment: handlers.NewAppointmentHandler(service.NewAppointmentService(appointmentRepo), settlementService),
		Financial:   handlers.NewFinancialHandler(service.NewFinancialService(ruleRepo, txRepo), settlementService),
	}

	app := fiber.New()
	app.Use(recover.New())

	healthCheck := func(c *fiber.Ctx) error {
		sqlDB, err := gdb.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "down"})
		}
		return c.JSON(fiber.Map{"status": "up"})
	}

	api.RegisterRoutes(app, h, issuer, healthCheck)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
