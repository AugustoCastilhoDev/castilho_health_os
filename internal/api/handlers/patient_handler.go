package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type PatientHandler struct {
	patients *service.PatientService
}

func NewPatientHandler(s *service.PatientService) *PatientHandler {
	return &PatientHandler{patients: s}
}

type patientRequest struct {
	Name          string  `json:"name"`
	Document      string  `json:"document"`
	BirthDate     *string `json:"birth_date"` // "YYYY-MM-DD"
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	AddressZip    string  `json:"address_zip"`
	AddressStreet string  `json:"address_street"`
	AddressCity   string  `json:"address_city"`
	AddressState  string  `json:"address_state"`
}

func (r patientRequest) toModel() (*models.Patient, error) {
	p := &models.Patient{
		Name:          r.Name,
		Document:      r.Document,
		Phone:         r.Phone,
		Email:         r.Email,
		AddressZip:    r.AddressZip,
		AddressStreet: r.AddressStreet,
		AddressCity:   r.AddressCity,
		AddressState:  r.AddressState,
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *r.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("%w: birth_date must be formatted as YYYY-MM-DD", service.ErrValidation)
		}
		p.BirthDate = &t
	}
	return p, nil
}

func (h *PatientHandler) Create(c *fiber.Ctx) error {
	var req patientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	patient, err := req.toModel()
	if err != nil {
		return respondErr(c, err)
	}
	if err := h.patients.Create(c.Context(), appmiddleware.TenantID(c), patient); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(patient)
}

func (h *PatientHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	patient, err := h.patients.Get(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(patient)
}

func (h *PatientHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req patientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	patient, err := req.toModel()
	if err != nil {
		return respondErr(c, err)
	}
	patient.ID = id
	if err := h.patients.Update(c.Context(), appmiddleware.TenantID(c), patient); err != nil {
		return respondErr(c, err)
	}
	return c.JSON(patient)
}

func (h *PatientHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.patients.Delete(c.Context(), appmiddleware.TenantID(c), id); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PatientHandler) Search(c *fiber.Ctx) error {
	patients, err := h.patients.Search(
		c.Context(),
		appmiddleware.TenantID(c),
		c.Query("q"),
		c.QueryInt("limit", 20),
		c.QueryInt("offset", 0),
	)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(patients)
}
