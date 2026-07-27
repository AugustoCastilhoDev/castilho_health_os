package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/castilho/health-os/internal/domain/models"
)

// RequireRole must be chained after RequireAuth. It 403s any request whose
// token role isn't in the allowed set — e.g. keeping FinancialRule writes
// to TENANT_ADMIN/FINANCE only.
func RequireRole(roles ...models.UserRole) fiber.Handler {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *fiber.Ctx) error {
		if !allowed[Role(c)] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
		}
		return c.Next()
	}
}
