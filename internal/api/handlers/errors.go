package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

// respondErr maps a service/repository error to the right HTTP status.
// Every handler routes its service-layer error through here instead of
// hand-rolling status codes, so the mapping stays in one place.
func respondErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	case errors.Is(err, repository.ErrInvalidTransition):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repository.ErrLocked):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrSettlementNotReady):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrConflict):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrValidation):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
}
