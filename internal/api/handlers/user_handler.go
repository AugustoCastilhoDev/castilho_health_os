package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{users: s}
}

type createUserRequest struct {
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	Role          string  `json:"role"`
	CouncilType   *string `json:"council_type"`
	CouncilNumber *string `json:"council_number"`
	CouncilState  *string `json:"council_state"`
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	user, err := h.users.Create(c.Context(), appmiddleware.TenantID(c), service.CreateUserInput{
		Name:          req.Name,
		Email:         req.Email,
		Password:      req.Password,
		Role:          models.UserRole(req.Role),
		CouncilType:   req.CouncilType,
		CouncilNumber: req.CouncilNumber,
		CouncilState:  req.CouncilState,
	})
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	user, err := h.users.Get(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(user)
}

// List backs GET /api/users?role=DOCTOR — role is optional; omit it to
// list every user in the tenant. This is what populates "pick a
// professional" dropdowns (e.g. when booking an appointment), so it's open
// to any authenticated role rather than admin-only.
func (h *UserHandler) List(c *fiber.Ctx) error {
	var role *models.UserRole
	if q := c.Query("role"); q != "" {
		r := models.UserRole(q)
		role = &r
	}
	users, err := h.users.List(c.Context(), appmiddleware.TenantID(c), role)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(users)
}

type updateUserRequest struct {
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	IsActive      bool    `json:"is_active"`
	CouncilType   *string `json:"council_type"`
	CouncilNumber *string `json:"council_number"`
	CouncilState  *string `json:"council_state"`
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req updateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	user, err := h.users.Update(c.Context(), appmiddleware.TenantID(c), id, service.UpdateUserInput{
		Name:          req.Name,
		Email:         req.Email,
		Role:          models.UserRole(req.Role),
		IsActive:      req.IsActive,
		CouncilType:   req.CouncilType,
		CouncilNumber: req.CouncilNumber,
		CouncilState:  req.CouncilState,
	})
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(user)
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) ChangeOwnPassword(c *fiber.Ctx) error {
	var req changePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	err := h.users.ChangeOwnPassword(c.Context(), appmiddleware.TenantID(c), appmiddleware.UserID(c), req.OldPassword, req.NewPassword)
	if err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req resetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.users.ResetPassword(c.Context(), appmiddleware.TenantID(c), id, req.NewPassword); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
