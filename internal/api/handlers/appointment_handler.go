package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type AppointmentHandler struct {
	appointments *service.AppointmentService
}

func NewAppointmentHandler(s *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointments: s}
}

type createAppointmentRequest struct {
	PatientID      uuid.UUID `json:"patient_id"`
	ProfessionalID uuid.UUID `json:"professional_id"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	DurationMin    int       `json:"duration_min"`
}

func (h *AppointmentHandler) Create(c *fiber.Ctx) error {
	var req createAppointmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	appt := &models.Appointment{
		PatientID:      req.PatientID,
		ProfessionalID: req.ProfessionalID,
		ScheduledAt:    req.ScheduledAt,
		DurationMin:    req.DurationMin,
	}
	if err := h.appointments.Create(c.Context(), appmiddleware.TenantID(c), appt); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(appt)
}

func (h *AppointmentHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	appt, err := h.appointments.Get(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(appt)
}

// ListByProfessional backs GET /api/appointments?professional_id=&from=&to=
// — the agenda view for a given professional's day/week.
func (h *AppointmentHandler) ListByProfessional(c *fiber.Ctx) error {
	professionalID, err := uuid.Parse(c.Query("professional_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "professional_id is required and must be a valid UUID"})
	}
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "from must be an RFC3339 timestamp"})
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to must be an RFC3339 timestamp"})
	}
	appts, err := h.appointments.ListByProfessional(c.Context(), appmiddleware.TenantID(c), professionalID, from, to)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(appts)
}

func (h *AppointmentHandler) ListByPatient(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	appts, err := h.appointments.ListByPatient(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(appts)
}

type transitionRequest struct {
	To     string `json:"to"`
	Reason string `json:"reason"`
}

func (h *AppointmentHandler) Transition(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req transitionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	appt, err := h.appointments.Transition(
		c.Context(),
		appmiddleware.TenantID(c),
		id,
		models.AppointmentStatus(req.To),
		appmiddleware.UserID(c),
		req.Reason,
	)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(appt)
}
