package handlers

import (
	"github.com/gofiber/fiber/v2"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type TenantHandler struct {
	tenants *service.TenantService
}

func NewTenantHandler(s *service.TenantService) *TenantHandler {
	return &TenantHandler{tenants: s}
}

type registerTenantRequest struct {
	TenantName     string `json:"tenant_name"`
	TenantSlug     string `json:"tenant_slug"`
	TenantType     string `json:"tenant_type"`
	TenantDocument string `json:"tenant_document"`
	TenantEmail    string `json:"tenant_email"`
	TenantPhone    string `json:"tenant_phone"`
	AdminName      string `json:"admin_name"`
	AdminEmail     string `json:"admin_email"`
	AdminPassword  string `json:"admin_password"`
}

// Register is the tenant onboarding endpoint: public (no auth token yet —
// there's no user to hold one), creates the clinic and its first
// TENANT_ADMIN. The caller still has to hit /auth/login afterwards; this
// deliberately doesn't auto-issue a token so registration and login stay
// two independently testable flows.
func (h *TenantHandler) Register(c *fiber.Ctx) error {
	var req registerTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tenant, admin, err := h.tenants.Register(c.Context(), service.RegisterTenantInput{
		TenantName:     req.TenantName,
		TenantSlug:     req.TenantSlug,
		TenantType:     models.TenantType(req.TenantType),
		TenantDocument: req.TenantDocument,
		TenantEmail:    req.TenantEmail,
		TenantPhone:    req.TenantPhone,
		AdminName:      req.AdminName,
		AdminEmail:     req.AdminEmail,
		AdminPassword:  req.AdminPassword,
	})
	if err != nil {
		return respondErr(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"tenant": fiber.Map{"id": tenant.ID, "slug": tenant.Slug, "name": tenant.Name},
		"admin":  fiber.Map{"id": admin.ID, "email": admin.Email},
	})
}

// GetCurrent returns the tenant for the session's own tenant_id (from JWT
// claims) — "which clinic am I logged into", for the app shell to display.
func (h *TenantHandler) GetCurrent(c *fiber.Ctx) error {
	tenant, err := h.tenants.Get(c.Context(), appmiddleware.TenantID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tenant)
}

type updateTenantProfileRequest struct {
	Name          string  `json:"name"`
	Document      string  `json:"document"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	AddressStreet *string `json:"address_street"`
	AddressCity   *string `json:"address_city"`
	AddressState  *string `json:"address_state"`
	AddressZip    *string `json:"address_zip"`
}

// UpdateProfile is the Configurações screen's save action — admin-only
// (see router.go), scoped to the session's own tenant_id like every other
// tenant-scoped write in this app.
func (h *TenantHandler) UpdateProfile(c *fiber.Ctx) error {
	var req updateTenantProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tenant, err := h.tenants.UpdateProfile(c.Context(), appmiddleware.TenantID(c), service.UpdateTenantProfileInput{
		Name:          req.Name,
		Document:      req.Document,
		Email:         req.Email,
		Phone:         req.Phone,
		AddressStreet: req.AddressStreet,
		AddressCity:   req.AddressCity,
		AddressState:  req.AddressState,
		AddressZip:    req.AddressZip,
	})
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tenant)
}

type logoUploadURLRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

// CreateLogoUploadURL is step 1 of the same two-step upload used for
// patient documents: mint a presigned PUT URL, the browser uploads
// straight to R2, then SetLogo persists the key.
func (h *TenantHandler) CreateLogoUploadURL(c *fiber.Ctx) error {
	var req logoUploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	uploadURL, fileKey, err := h.tenants.CreateLogoUploadURL(c.Context(), appmiddleware.TenantID(c), req.FileName, req.ContentType)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"upload_url": uploadURL, "file_key": fileKey})
}

type setLogoRequest struct {
	FileKey string `json:"file_key"`
}

func (h *TenantHandler) SetLogo(c *fiber.Ctx) error {
	var req setLogoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	tenant, err := h.tenants.SetLogo(c.Context(), appmiddleware.TenantID(c), req.FileKey)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tenant)
}

func (h *TenantHandler) DeleteLogo(c *fiber.Ctx) error {
	if err := h.tenants.DeleteLogo(c.Context(), appmiddleware.TenantID(c)); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// LogoDownloadURL backs the logo preview on the Configurações screen —
// returns an empty url (200, not an error) when no logo is set yet.
func (h *TenantHandler) LogoDownloadURL(c *fiber.Ctx) error {
	url, err := h.tenants.LogoDownloadURL(c.Context(), appmiddleware.TenantID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(fiber.Map{"url": url})
}
