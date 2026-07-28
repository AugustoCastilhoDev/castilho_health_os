package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/service"
)

type MemedHandler struct {
	memed *service.MemedService
	// scriptURL is not a secret (it's the public URL of Memed's own widget
	// script) — it's returned alongside the token so sandbox vs. production
	// can be swapped via one env var (main.go) without a frontend change.
	scriptURL string
}

func NewMemedHandler(s *service.MemedService, scriptURL string) *MemedHandler {
	return &MemedHandler{memed: s, scriptURL: scriptURL}
}

// GetPrescriberToken backs the frontend loading Memed's own widget: it
// never returns anything but the per-professional token, and always for
// the authenticated caller — a professional can only ever fetch their own
// token, never someone else's by passing an id.
func (h *MemedHandler) GetPrescriberToken(c *fiber.Ctx) error {
	token, err := h.memed.GetPrescriberToken(c.Context(), appmiddleware.TenantID(c), appmiddleware.UserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"token": token, "script_url": h.scriptURL})
}

type createPrescriptionLogRequest struct {
	MemedPrescriptionID string `json:"memed_prescription_id"`
}

// CreatePrescriptionLog is called by the frontend right after Memed's
// widget fires its "prescricaoImpressa" event — professional_id always
// comes from the authenticated session, never the request body, same
// convention as PatientDocumentHandler.Create's uploaded_by_id.
func (h *MemedHandler) CreatePrescriptionLog(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	var req createPrescriptionLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	log, err := h.memed.RecordIssuance(c.Context(), appmiddleware.TenantID(c), patientID, appmiddleware.UserID(c), req.MemedPrescriptionID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(log)
}

func (h *MemedHandler) ListByPatient(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	logs, err := h.memed.ListByPatient(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(logs)
}

// Cancel mirrors Memed's prescricaoExcluida event — the frontend calls this
// when the professional deletes a prescription from within Memed's widget.
func (h *MemedHandler) Cancel(c *fiber.Ctx) error {
	memedPrescriptionID := c.Params("memedPrescriptionID")
	if err := h.memed.Cancel(c.Context(), appmiddleware.TenantID(c), memedPrescriptionID); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
