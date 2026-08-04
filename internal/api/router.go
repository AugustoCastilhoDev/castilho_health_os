// Package api wires HTTP routes to handlers and middleware. Handlers and
// middleware live in their own subpackages (internal/api/handlers,
// internal/api/middleware) so this file stays a pure routing table.
package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/castilho/health-os/internal/api/handlers"
	"github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
)

type Handlers struct {
	Auth             *handlers.AuthHandler
	Tenant           *handlers.TenantHandler
	User             *handlers.UserHandler
	Patient          *handlers.PatientHandler
	Appointment      *handlers.AppointmentHandler
	Financial        *handlers.FinancialHandler
	MedicalRecord    *handlers.MedicalRecordHandler
	DocumentTemplate *handlers.DocumentTemplateHandler
	PatientDocument  *handlers.PatientDocumentHandler
	Memed            *handlers.MemedHandler
	Stock            *handlers.StockHandler
	Odontograma      *handlers.OdontogramaHandler
}

func RegisterRoutes(app *fiber.App, h *Handlers, issuer *auth.JWTIssuer, healthCheck fiber.Handler) {
	app.Get("/health", healthCheck)

	// Public: no tenant/user exists yet to hold a token.
	app.Post("/auth/register", h.Tenant.Register)
	app.Post("/auth/login", h.Auth.Login)

	protected := app.Group("/api", middleware.RequireAuth(issuer))

	protected.Get("/tenant", h.Tenant.GetCurrent)

	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":   middleware.UserID(c),
			"tenant_id": middleware.TenantID(c),
			"role":      middleware.Role(c),
		})
	})

	// Role groups, named for what they gate rather than who's in them, so
	// the intent reads at the call site.
	frontDesk := middleware.RequireRole(models.RoleTenantAdmin, models.RoleReceptionist)
	admin := middleware.RequireRole(models.RoleTenantAdmin)
	finance := middleware.RequireRole(models.RoleTenantAdmin, models.RoleFinance)
	health := middleware.RequireRole(models.RoleDoctor, models.RoleDentist)
	dentist := middleware.RequireRole(models.RoleDentist)

	// Get/List are open to any authenticated role (they back "pick a
	// professional" pickers); mutating a user's account is admin-only.
	// Self password change lives under /users/me/password rather than
	// gating it with `admin` — it's a different concern (auth to your own
	// account, not staff management).
	users := protected.Group("/users")
	users.Post("/", admin, h.User.Create)
	users.Get("/", h.User.List)
	users.Get("/:id", h.User.Get)
	users.Put("/:id", admin, h.User.Update)
	users.Post("/:id/reset-password", admin, h.User.ResetPassword)
	users.Put("/me/password", h.User.ChangeOwnPassword)

	patients := protected.Group("/patients")
	patients.Post("/", frontDesk, h.Patient.Create)
	patients.Get("/", h.Patient.Search)
	patients.Get("/:id", h.Patient.Get)
	patients.Put("/:id", frontDesk, h.Patient.Update)
	patients.Delete("/:id", admin, h.Patient.Delete)
	patients.Get("/:patientID/appointments", h.Appointment.ListByPatient)
	patients.Get("/:patientID/medical-records", h.MedicalRecord.ListByPatient)
	// Two-step upload (see PatientDocumentHandler): step 1 mints a presigned
	// PUT URL, the browser PUTs the bytes straight to R2, step 2 persists
	// the metadata row. Open to any authenticated role, same as front-desk
	// registering a payment.
	patients.Post("/:patientID/documents/upload-url", h.PatientDocument.CreateUploadURL)
	patients.Post("/:patientID/documents", h.PatientDocument.Create)
	patients.Get("/:patientID/documents", h.PatientDocument.ListByPatient)
	// Issuing a prescription is health-professional-only, same as writing a
	// medical record; reading the audit trail is open like the rest of the
	// patient record.
	patients.Post("/:patientID/memed-prescriptions", health, h.Memed.CreatePrescriptionLog)
	patients.Get("/:patientID/memed-prescriptions", h.Memed.ListByPatient)
	// Odontograma: reading the chart/history stays open like the rest of the
	// patient record; recording a finding/procedure is dentist-only rather
	// than the broader `health` group, since this module is odonto-exclusive.
	patients.Get("/:patientID/odontograma", h.Odontograma.GetChart)
	patients.Get("/:patientID/odontograma-entries", h.Odontograma.ListByPatient)

	// Who may trigger which specific status transition (e.g. only the
	// assigned professional starts IN_PROGRESS) isn't enforced yet — the
	// state machine itself (CanTransitionTo) is the real guard here, role
	// policy on individual transitions is a later refinement.
	appointments := protected.Group("/appointments")
	appointments.Post("/", frontDesk, h.Appointment.Create)
	appointments.Get("/", h.Appointment.ListByProfessional)
	appointments.Get("/:id", h.Appointment.Get)
	appointments.Post("/:id/transition", h.Appointment.Transition)
	// Manual retry for automatic settlement (see SettlementService) — for
	// when the auto-trigger on COMPLETED/mark-paid couldn't run yet, e.g. no
	// financial rule was configured for the professional at that time.
	appointments.Post("/:id/settle", finance, h.Appointment.Settle)

	rules := protected.Group("/financial-rules")
	rules.Post("/", finance, h.Financial.CreateRule)
	rules.Get("/", h.Financial.ListRulesByProfessional)
	rules.Get("/:id", h.Financial.GetRule)
	rules.Put("/:id", finance, h.Financial.UpdateRule)

	// CreateTransaction itself is open to any authenticated role because a
	// PATIENT_PAYMENT is normal front-desk work; the PROFESSIONAL_PAYOUT
	// restriction lives in FinancialService since only the request body
	// (its Type field) reveals which case applies.
	transactions := protected.Group("/financial-transactions")
	transactions.Post("/", h.Financial.CreateTransaction)
	transactions.Get("/", h.Financial.ListTransactions)
	transactions.Get("/:id", h.Financial.GetTransaction)
	transactions.Post("/:id/mark-paid", finance, h.Financial.MarkPaid)
	transactions.Get("/appointment/:appointmentID", h.Financial.ListTransactionsByAppointment)

	// Create/Update/Lock are restricted to health professionals (writing a
	// clinical note is their call, not front-desk/admin); Get/ListByPatient
	// stay open to any authenticated role like the rest of the patient
	// record.
	medicalRecords := protected.Group("/medical-records")
	medicalRecords.Post("/", health, h.MedicalRecord.Create)
	medicalRecords.Get("/:id", h.MedicalRecord.Get)
	medicalRecords.Put("/:id", health, h.MedicalRecord.Update)
	medicalRecords.Post("/:id/lock", health, h.MedicalRecord.Lock)

	// Managing the reusable template body is admin-only; Get/List/Generate
	// are open to any authenticated role since generating an atestado is
	// normal day-to-day use, not clinic configuration.
	documentTemplates := protected.Group("/document-templates")
	documentTemplates.Post("/", admin, h.DocumentTemplate.Create)
	documentTemplates.Get("/", h.DocumentTemplate.List)
	documentTemplates.Get("/:id", h.DocumentTemplate.Get)
	documentTemplates.Put("/:id", admin, h.DocumentTemplate.Update)
	documentTemplates.Post("/:id/generate", h.DocumentTemplate.Generate)

	// Download is open to any authenticated role (viewing an exam is normal
	// clinical work); deleting a patient file outright is admin-only, same
	// sensitivity level as Patient.Delete.
	patientDocuments := protected.Group("/patient-documents")
	patientDocuments.Get("/:id/download-url", h.PatientDocument.DownloadURL)
	patientDocuments.Delete("/:id", admin, h.PatientDocument.Delete)

	// GetPrescriberToken is what the frontend calls to load Memed's own
	// widget for the current professional — health-professional-only since
	// only DOCTOR/DENTIST issue prescriptions.
	protected.Get("/memed/token", health, h.Memed.GetPrescriberToken)
	protected.Post("/memed-prescriptions/:memedPrescriptionID/cancel", health, h.Memed.Cancel)

	// Estoque: managing items/quantities is front-desk work (same group
	// that registers a patient payment), same reasoning as FinancialRule
	// being gated by `finance` — List/Get stay open to any authenticated
	// role since checking stock is normal day-to-day use.
	stockItems := protected.Group("/stock-items")
	stockItems.Post("/", frontDesk, h.Stock.CreateItem)
	stockItems.Get("/", h.Stock.ListItems)
	stockItems.Get("/:id", h.Stock.GetItem)
	stockItems.Put("/:id", frontDesk, h.Stock.UpdateItem)
	stockItems.Post("/:itemID/movements", frontDesk, h.Stock.RecordMovement)
	stockItems.Get("/:itemID/movements", h.Stock.ListMovements)

	odontogramaEntries := protected.Group("/odontograma-entries")
	odontogramaEntries.Post("/", dentist, h.Odontograma.CreateEntry)

	protected.Get("/admin/ping", admin, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})
}
