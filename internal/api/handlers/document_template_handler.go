package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type DocumentTemplateHandler struct {
	templates *service.DocumentTemplateService
	patients  *service.PatientService
	users     *service.UserService
}

func NewDocumentTemplateHandler(templates *service.DocumentTemplateService, patients *service.PatientService, users *service.UserService) *DocumentTemplateHandler {
	return &DocumentTemplateHandler{templates: templates, patients: patients, users: users}
}

type documentTemplateRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	IsActive bool   `json:"is_active"`
}

func (r documentTemplateRequest) toModel() *models.DocumentTemplate {
	return &models.DocumentTemplate{
		Name:     r.Name,
		Type:     models.DocumentTemplateType(r.Type),
		Content:  r.Content,
		IsActive: r.IsActive,
	}
}

func (h *DocumentTemplateHandler) Create(c *fiber.Ctx) error {
	var req documentTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tmpl := req.toModel()
	if err := h.templates.Create(c.Context(), appmiddleware.TenantID(c), tmpl); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(tmpl)
}

func (h *DocumentTemplateHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	tmpl, err := h.templates.Get(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tmpl)
}

func (h *DocumentTemplateHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req documentTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tmpl := req.toModel()
	tmpl.ID = id
	if err := h.templates.Update(c.Context(), appmiddleware.TenantID(c), tmpl); err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tmpl)
}

func (h *DocumentTemplateHandler) List(c *fiber.Ctx) error {
	templates, err := h.templates.ListByTenant(c.Context(), appmiddleware.TenantID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(templates)
}

type generateDocumentRequest struct {
	PatientID      *uuid.UUID        `json:"patient_id"`
	ProfessionalID *uuid.UUID        `json:"professional_id"`
	Vars           map[string]string `json:"vars"`
}

// Generate fills a template's {{tag}} placeholders and streams back the
// rendered PDF. patient_name/professional_name/date are always resolved
// server-side (never trusted from the request body) so a generated document
// can't be tricked into naming the wrong patient; any other tag (cid,
// days_off, ...) comes from the caller's vars.
func (h *DocumentTemplateHandler) Generate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req generateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tenantID := appmiddleware.TenantID(c)
	vars := map[string]string{}
	for k, v := range req.Vars {
		vars[k] = v
	}

	if req.PatientID != nil {
		patient, err := h.patients.Get(c.Context(), tenantID, *req.PatientID)
		if err != nil {
			return respondErr(c, err)
		}
		vars["patient_name"] = patient.Name
	}
	if req.ProfessionalID != nil {
		professional, err := h.users.Get(c.Context(), tenantID, *req.ProfessionalID)
		if err != nil {
			return respondErr(c, err)
		}
		vars["professional_name"] = professional.Name
	}
	vars["date"] = time.Now().Format("02/01/2006")

	pdfBytes, tmpl, err := h.templates.Generate(c.Context(), tenantID, id, vars)
	if err != nil {
		return respondErr(c, err)
	}

	filename := strings.ReplaceAll(tmpl.Name, `"`, "") + ".pdf"
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(pdfBytes)
}
