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
	Auth        *handlers.AuthHandler
	Tenant      *handlers.TenantHandler
	User        *handlers.UserHandler
	Patient     *handlers.PatientHandler
	Appointment *handlers.AppointmentHandler
	Financial   *handlers.FinancialHandler
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

	protected.Get("/admin/ping", admin, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})
}
