package handlers

import (
	"fmt"
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
	tenants   *service.TenantService
}

func NewDocumentTemplateHandler(templates *service.DocumentTemplateService, patients *service.PatientService, users *service.UserService, tenants *service.TenantService) *DocumentTemplateHandler {
	return &DocumentTemplateHandler{templates: templates, patients: patients, users: users, tenants: tenants}
}

type documentTemplateRequest struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Content          string `json:"content"`
	IsActive         bool   `json:"is_active"`
	IncludeHeader    bool   `json:"include_header"`
	IncludeFooter    bool   `json:"include_footer"`
	IncludeSignature bool   `json:"include_signature"`
	IncludeStamp     bool   `json:"include_stamp"`
}

func (r documentTemplateRequest) toModel() *models.DocumentTemplate {
	return &models.DocumentTemplate{
		Name:             r.Name,
		Type:             models.DocumentTemplateType(r.Type),
		Content:          r.Content,
		IsActive:         r.IsActive,
		IncludeHeader:    r.IncludeHeader,
		IncludeFooter:    r.IncludeFooter,
		IncludeSignature: r.IncludeSignature,
		IncludeStamp:     r.IncludeStamp,
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
// days_off, ...) comes from the caller's vars. Clinic/professional
// letterhead data is resolved here too (branding), separately from vars —
// see DocumentTemplateService.Generate for why layout flags themselves
// live on the template, not in this handler.
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

	var professional *models.User
	if req.PatientID != nil {
		patient, err := h.patients.Get(c.Context(), tenantID, *req.PatientID)
		if err != nil {
			return respondErr(c, err)
		}
		vars["patient_name"] = patient.Name
	}
	if req.ProfessionalID != nil {
		professional, err = h.users.Get(c.Context(), tenantID, *req.ProfessionalID)
		if err != nil {
			return respondErr(c, err)
		}
		vars["professional_name"] = professional.Name
	}
	now := time.Now()
	vars["date"] = now.Format("02/01/2006")

	tenant, err := h.tenants.Get(c.Context(), tenantID)
	if err != nil {
		return respondErr(c, err)
	}
	branding := service.DocumentBranding{
		ClinicName:     tenant.Name,
		ClinicDocument: tenant.Document,
		ClinicAddress:  tenantAddressLine(tenant),
		ClinicPhone:    tenant.Phone,
		ClinicEmail:    tenant.Email,
		SignatureCity:  derefStr(tenant.AddressCity),
		SignatureDate:  formatDatePtBR(now),
	}
	if professional != nil {
		branding.ProfessionalName = professional.Name
		branding.ProfessionalRegistration = councilRegistration(professional)
	}
	// Best-effort: a missing/unreachable logo must never block generating
	// the document itself — the header just renders without one.
	if logoBytes, err := h.tenants.GetLogoBytes(c.Context(), tenantID); err == nil && len(logoBytes) > 0 {
		branding.LogoBytes = logoBytes
		branding.LogoMimeType = mimeTypeFromFileKey(derefStr(tenant.LogoKey))
	}

	pdfBytes, tmpl, err := h.templates.Generate(c.Context(), tenantID, id, vars, branding)
	if err != nil {
		return respondErr(c, err)
	}

	filename := strings.ReplaceAll(tmpl.Name, `"`, "") + ".pdf"
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(pdfBytes)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// tenantAddressLine joins whatever address parts a tenant has filled in
// into a single letterhead line — no part is required, so a clinic that
// hasn't set an address yet just gets a shorter (or empty) line rather than
// a broken-looking gap.
func tenantAddressLine(t *models.Tenant) string {
	cityStateParts := make([]string, 0, 2)
	for _, p := range []string{derefStr(t.AddressCity), derefStr(t.AddressState)} {
		if p != "" {
			cityStateParts = append(cityStateParts, p)
		}
	}
	cityState := strings.Join(cityStateParts, "/")

	zip := derefStr(t.AddressZip)
	if zip != "" {
		zip = "CEP " + zip
	}

	lineParts := make([]string, 0, 3)
	for _, p := range []string{derefStr(t.AddressStreet), cityState, zip} {
		if p != "" {
			lineParts = append(lineParts, p)
		}
	}
	return strings.Join(lineParts, " - ")
}

// councilRegistration prints "CRM/SP 123456" style text — empty when the
// professional has no council fields on file (e.g. front-desk staff, or a
// health professional who hasn't completed their profile yet).
func councilRegistration(u *models.User) string {
	if u.CouncilType == nil || *u.CouncilType == "" || u.CouncilNumber == nil || *u.CouncilNumber == "" {
		return ""
	}
	reg := *u.CouncilType
	if u.CouncilState != nil && *u.CouncilState != "" {
		reg += "/" + *u.CouncilState
	}
	return reg + " " + *u.CouncilNumber
}

func mimeTypeFromFileKey(key string) string {
	lower := strings.ToLower(key)
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	default:
		return "image/png"
	}
}

var portugueseMonths = [...]string{
	"janeiro", "fevereiro", "março", "abril", "maio", "junho",
	"julho", "agosto", "setembro", "outubro", "novembro", "dezembro",
}

// formatDatePtBR renders "4 de agosto de 2026" for the signature block —
// Go's time package has no built-in Portuguese locale, and pulling in a
// full i18n dependency for one date format isn't worth it.
func formatDatePtBR(t time.Time) string {
	return fmt.Sprintf("%d de %s de %d", t.Day(), portugueseMonths[t.Month()-1], t.Year())
}
