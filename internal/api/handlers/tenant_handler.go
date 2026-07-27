package handlers

import (
	"github.com/gofiber/fiber/v2"

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
