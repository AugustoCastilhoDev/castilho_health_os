package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/castilho/health-os/internal/api"
	"github.com/castilho/health-os/internal/api/handlers"
	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/config"
	"github.com/castilho/health-os/internal/infra/db"
	"github.com/castilho/health-os/internal/memed"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
	"github.com/castilho/health-os/internal/storage"
)

// memedClientAdapter bridges *memed.Client to service.MemedClient — the two
// packages deliberately declare their own Prescriber struct (see
// MemedClient's doc comment in internal/service/memed_service.go), so this
// small conversion is the only place that needs to know both shapes.
type memedClientAdapter struct {
	c *memed.Client
}

func (a memedClientAdapter) FetchOrRegisterToken(ctx context.Context, p service.MemedPrescriber) (string, error) {
	return a.c.FetchOrRegisterToken(ctx, memed.Prescriber{
		ExternalID:  p.ExternalID,
		Name:        p.Name,
		Surname:     p.Surname,
		CPF:         p.CPF,
		BoardCode:   p.BoardCode,
		BoardNumber: p.BoardNumber,
		BoardState:  p.BoardState,
		BirthDate:   p.BirthDate,
		Email:       p.Email,
		Phone:       p.Phone,
		Sex:         p.Sex,
	})
}

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
	medicalRecordRepo := repository.NewMedicalRecordRepository(gdb)
	documentTemplateRepo := repository.NewDocumentTemplateRepository(gdb)
	patientDocumentRepo := repository.NewPatientDocumentRepository(gdb)
	memedLogRepo := repository.NewMemedPrescriptionLogRepository(gdb)
	stockItemRepo := repository.NewStockItemRepository(gdb)
	stockMovementRepo := repository.NewStockMovementRepository(gdb)
	odontogramaRepo := repository.NewOdontogramaRepository(gdb)

	settlementService := service.NewSettlementService(appointmentRepo, ruleRepo, txRepo)
	patientService := service.NewPatientService(patientRepo)
	userService := service.NewUserService(userRepo)

	// R2 is optional (see Config.IsR2Configured) — leaving objectStorage as
	// a true nil interface (never assigned a typed-nil *storage.R2Client)
	// when unconfigured is what lets PatientDocumentService's `== nil`
	// check work correctly.
	var objectStorage service.ObjectStorage
	if cfg.IsR2Configured() {
		r2Client, err := storage.NewR2Client(context.Background(), storage.R2Config{
			AccountID:       cfg.R2AccountID,
			AccessKeyID:     cfg.R2AccessKeyID,
			SecretAccessKey: cfg.R2SecretAccessKey,
			Bucket:          cfg.R2BucketName,
		})
		if err != nil {
			log.Fatalf("storage: %v", err)
		}
		objectStorage = r2Client
		log.Println("R2 object storage configured — document upload/download enabled")
	} else {
		log.Println("R2 not configured (R2_* env vars unset) — document upload/download disabled")
	}

	// Same optional-config pattern as R2 above: a true nil interface until
	// MEMED_API_KEY/MEMED_SECRET_KEY are both set, so MemedService's
	// `== nil` check works and the app still starts without them.
	var memedClient service.MemedClient
	if cfg.IsMemedConfigured() {
		memedClient = memedClientAdapter{memed.NewClient(memed.Config{
			APIKey:    cfg.MemedAPIKey,
			SecretKey: cfg.MemedSecretKey,
			BaseURL:   cfg.MemedAPIURL,
		})}
		log.Println("Memed configured — prescription issuance enabled")
	} else {
		log.Println("Memed not configured (MEMED_API_KEY/MEMED_SECRET_KEY unset) — prescription issuance disabled")
	}

	h := &api.Handlers{
		Auth:             handlers.NewAuthHandler(service.NewAuthService(tenantRepo, userRepo, issuer)),
		Tenant:           handlers.NewTenantHandler(service.NewTenantService(gdb, tenantRepo)),
		User:             handlers.NewUserHandler(userService),
		Patient:          handlers.NewPatientHandler(patientService),
		Appointment:      handlers.NewAppointmentHandler(service.NewAppointmentService(appointmentRepo), settlementService),
		Financial:        handlers.NewFinancialHandler(service.NewFinancialService(ruleRepo, txRepo), settlementService),
		MedicalRecord:    handlers.NewMedicalRecordHandler(service.NewMedicalRecordService(medicalRecordRepo)),
		DocumentTemplate: handlers.NewDocumentTemplateHandler(service.NewDocumentTemplateService(documentTemplateRepo), patientService, userService),
		PatientDocument:  handlers.NewPatientDocumentHandler(service.NewPatientDocumentService(patientDocumentRepo, objectStorage)),
		Memed:            handlers.NewMemedHandler(service.NewMemedService(memedLogRepo, userRepo, memedClient), cfg.MemedFrontendScriptURL),
		Stock:            handlers.NewStockHandler(service.NewStockService(stockItemRepo, stockMovementRepo)),
		Odontograma:      handlers.NewOdontogramaHandler(service.NewOdontogramaService(odontogramaRepo)),
	}

	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		// Content-Disposition carries the generated document's filename
		// (see DocumentTemplateHandler.Generate) — browsers hide response
		// headers from JS on cross-origin requests unless explicitly
		// exposed, so without this the frontend can't read it.
		ExposeHeaders: "Content-Disposition",
	}))

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
