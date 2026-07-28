package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type PatientDocumentHandler struct {
	documents *service.PatientDocumentService
}

func NewPatientDocumentHandler(s *service.PatientDocumentService) *PatientDocumentHandler {
	return &PatientDocumentHandler{documents: s}
}

type uploadURLRequest struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
}

// CreateUploadURL is step 1 of 2 for an upload: it hands back a presigned
// PUT URL the browser sends the file bytes to directly. Nothing is
// persisted here — the metadata row (step 2, Create) only exists once the
// upload itself has actually succeeded.
func (h *PatientDocumentHandler) CreateUploadURL(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	var req uploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	uploadURL, fileKey, err := h.documents.CreateUploadURL(c.Context(), appmiddleware.TenantID(c), patientID, req.FileName, req.MimeType)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"upload_url": uploadURL, "file_key": fileKey})
}

type patientDocumentRequest struct {
	FileKey     string `json:"file_key"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	MimeType    string `json:"mime_type"`
	Description string `json:"description"`
}

// Create is step 2 of 2: persisting the metadata row once the browser's
// direct PUT to R2 has finished. uploaded_by_id always comes from the
// authenticated session, never the request body.
func (h *PatientDocumentHandler) Create(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	var req patientDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	doc := &models.PatientDocument{
		PatientID:    patientID,
		UploadedByID: appmiddleware.UserID(c),
		FileKey:      req.FileKey,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		Description:  req.Description,
	}
	if err := h.documents.Create(c.Context(), appmiddleware.TenantID(c), doc); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(doc)
}

func (h *PatientDocumentHandler) ListByPatient(c *fiber.Ctx) error {
	patientID, err := uuid.Parse(c.Params("patientID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid patient id"})
	}
	docs, err := h.documents.ListByPatient(c.Context(), appmiddleware.TenantID(c), patientID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(docs)
}

func (h *PatientDocumentHandler) DownloadURL(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	url, err := h.documents.DownloadURL(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"url": url})
}

func (h *PatientDocumentHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.documents.Delete(c.Context(), appmiddleware.TenantID(c), id); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
