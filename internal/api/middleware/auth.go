// Package middleware wires cross-cutting Fiber concerns: JWT
// authentication and role-based authorization.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
)

const (
	localsUserID   = "user_id"
	localsTenantID = "tenant_id"
	localsRole     = "role"
)

// RequireAuth validates the bearer JWT and stashes its claims in
// c.Locals for downstream handlers (via UserID/TenantID/Role below) and
// for RequireRole. It must run before any route that calls those helpers.
func RequireAuth(issuer *auth.JWTIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}

		claims, err := issuer.Parse(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals(localsUserID, claims.UserID)
		c.Locals(localsTenantID, claims.TenantID)
		c.Locals(localsRole, claims.Role)
		return c.Next()
	}
}

// UserID, TenantID and Role read the claims RequireAuth stashed earlier in
// the chain. They panic if called on a route that skipped RequireAuth —
// that's a routing bug, not a runtime condition to handle gracefully.
func UserID(c *fiber.Ctx) uuid.UUID { return c.Locals(localsUserID).(uuid.UUID) }
func TenantID(c *fiber.Ctx) uuid.UUID { return c.Locals(localsTenantID).(uuid.UUID) }
func Role(c *fiber.Ctx) models.UserRole { return c.Locals(localsRole).(models.UserRole) }
