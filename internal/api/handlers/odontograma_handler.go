package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type OdontogramaHandler struct {
	odontograma *service.OdontogramaService
}

func NewOdontogramaHandler(s *service.OdontogramaService) *OdontogramaHandler {
	return &OdontogramaHandler{odontograma: s}
}

type odontogramaEntryRequest struct {
	PatientID   string  `json:"patient_id"`
	ToothNumber string  `json:"tooth_number"`
	Condition   string  `json:"condition"`
	Note        *string `json:"note"`
}

func (h *OdontogramaHandler) CreateEntry(c *fiber.Ctx) error {
	var req odontogramaEntryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient_id"})
	}
	entry := &models.OdontogramaEntry{
		PatientID:    patientID,
		ToothNumber:  req.ToothNumber,
		Condition:    models.ToothCondition(req.Condition),
		Note:         req.Note,
		RecordedByID: appmiddleware.UserID(c),
	}
	if err := h.odontograma.RecordEntry(c.Context(), appmiddleware.TenantID(c), entry); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(entry)
}

func (h *OdontogramaHandler) ListByPatient(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	entries, err := h.odontograma.ListByPatient(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(entries)
}

func (h *OdontogramaHandler) GetChart(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	chart, err := h.odontograma.CurrentChart(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(chart)
}
