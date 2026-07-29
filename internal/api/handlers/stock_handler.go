package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type StockHandler struct {
	stock *service.StockService
}

func NewStockHandler(s *service.StockService) *StockHandler {
	return &StockHandler{stock: s}
}

type stockItemRequest struct {
	Name            string `json:"name"`
	Unit            string `json:"unit"`
	MinQuantity     *int   `json:"min_quantity"`
	IsActive        bool   `json:"is_active"`
	InitialQuantity int    `json:"initial_quantity"`
}

func (h *StockHandler) CreateItem(c *fiber.Ctx) error {
	var req stockItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	item := &models.StockItem{Name: req.Name, Unit: req.Unit, MinQuantity: req.MinQuantity}
	saved, err := h.stock.CreateItem(c.Context(), appmiddleware.TenantID(c), item, req.InitialQuantity, appmiddleware.UserID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(saved)
}

func (h *StockHandler) GetItem(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	item, err := h.stock.GetItem(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(item)
}

func (h *StockHandler) ListItems(c *fiber.Ctx) error {
	items, err := h.stock.ListItems(c.Context(), appmiddleware.TenantID(c))
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(items)
}

func (h *StockHandler) UpdateItem(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req stockItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	item := &models.StockItem{Name: req.Name, Unit: req.Unit, MinQuantity: req.MinQuantity, IsActive: req.IsActive}
	item.ID = id
	saved, err := h.stock.UpdateItem(c.Context(), appmiddleware.TenantID(c), item)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(saved)
}

type stockMovementRequest struct {
	Type     string  `json:"type"`
	Quantity int     `json:"quantity"`
	Note     *string `json:"note"`
}

func (h *StockHandler) RecordMovement(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item id"})
	}
	var req stockMovementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	movement := &models.StockMovement{
		ItemID:      itemID,
		Type:        models.StockMovementType(req.Type),
		Quantity:    req.Quantity,
		Note:        req.Note,
		CreatedByID: appmiddleware.UserID(c),
	}
	item, err := h.stock.RecordMovement(c.Context(), appmiddleware.TenantID(c), movement)
	if err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"item": item, "movement": movement})
}

func (h *StockHandler) ListMovements(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item id"})
	}
	movements, err := h.stock.ListMovementsByItem(c.Context(), appmiddleware.TenantID(c), itemID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(movements)
}
