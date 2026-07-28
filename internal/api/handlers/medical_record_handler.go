package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type MedicalRecordHandler struct {
	records *service.MedicalRecordService
}

func NewMedicalRecordHandler(s *service.MedicalRecordService) *MedicalRecordHandler {
	return &MedicalRecordHandler{records: s}
}

type medicalRecordRequest struct {
	PatientID      uuid.UUID  `json:"patient_id"`
	ProfessionalID uuid.UUID  `json:"professional_id"`
	AppointmentID  *uuid.UUID `json:"appointment_id"`
	Type           string     `json:"type"`
	Content        string     `json:"content"`
}

func (r medicalRecordRequest) toModel() *models.MedicalRecord {
	return &models.MedicalRecord{
		PatientID:      r.PatientID,
		ProfessionalID: r.ProfessionalID,
		AppointmentID:  r.AppointmentID,
		Type:           models.MedicalRecordType(r.Type),
		Content:        r.Content,
	}
}

func (h *MedicalRecordHandler) Create(c *fiber.Ctx) error {
	var req medicalRecordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	record := req.toModel()
	if err := h.records.Create(c.Context(), appmiddleware.TenantID(c), record); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(record)
}

func (h *MedicalRecordHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	record, err := h.records.Get(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(record)
}

func (h *MedicalRecordHandler) ListByPatient(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	records, err := h.records.ListByPatient(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(records)
}

// Update only lets the caller change Type/Content — see
// MedicalRecordService.Update — so patient_id/professional_id/appointment_id
// in the body are accepted but silently ignored, same as PUT /patients/:id
// ignoring an id in the body.
func (h *MedicalRecordHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req medicalRecordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	record := req.toModel()
	record.ID = id
	if err := h.records.Update(c.Context(), appmiddleware.TenantID(c), record); err != nil {
		return respondErr(c, err)
	}
	return c.JSON(record)
}

func (h *MedicalRecordHandler) Lock(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	record, err := h.records.Lock(c.Context(), appmiddleware.TenantID(c), id, appmiddleware.UserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(record)
}
