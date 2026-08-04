package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

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

// ImportItems takes a multipart-uploaded CSV (field name "file") and
// creates one StockItem per valid row — see parseStockImportCSV for the
// expected columns and service.StockService.ImportItems for how
// duplicates/row-level failures are handled. Malformed-file parse issues
// (bad quantity, unreadable line) are merged into the same "failed" list
// the service reports, so the caller sees one combined summary either way.
func (h *StockHandler) ImportItems(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "arquivo CSV é obrigatório"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "não foi possível abrir o arquivo"})
	}
	defer file.Close()

	rows, parseIssues, err := parseStockImportCSV(file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	result := h.stock.ImportItems(c.Context(), appmiddleware.TenantID(c), rows, appmiddleware.UserID(c))
	// Built via make(..., 0, ...) rather than append(parseIssues, result.Failed...)
	// on purpose: appending zero elements onto a nil parseIssues returns nil,
	// not an empty slice, which serializes to JSON `null` instead of `[]` —
	// found live when a CSV with zero failures crashed the frontend on
	// `result.failed.length` (Cannot read properties of null).
	failed := make([]service.StockImportIssue, 0, len(parseIssues)+len(result.Failed))
	failed = append(failed, parseIssues...)
	failed = append(failed, result.Failed...)
	result.Failed = failed
	return c.JSON(result)
}

// parseStockImportCSV reads a CSV with a required header row. Recognized
// headers (case-insensitive, order doesn't matter): "nome"/"name" and
// "unidade"/"unit" (required), "quantidade_minima"/"min_quantity" and
// "quantidade_inicial"/"initial_quantity" (optional). A row with an
// unparseable quantity is reported as an issue and excluded from rows —
// everything else (missing name/unit, negative quantity) is left to
// StockService.ImportItems, which already validates those via CreateItem.
func parseStockImportCSV(r io.Reader) ([]service.StockImportRow, []service.StockImportIssue, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("não foi possível ler o cabeçalho do CSV")
	}
	colIndex := make(map[string]int, len(header))
	for i, h := range header {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}
	nameIdx, hasName := firstCSVColumn(colIndex, "nome", "name")
	unitIdx, hasUnit := firstCSVColumn(colIndex, "unidade", "unit")
	if !hasName || !hasUnit {
		return nil, nil, fmt.Errorf(`o CSV precisa ter as colunas "nome" e "unidade"`)
	}
	minIdx, hasMin := firstCSVColumn(colIndex, "quantidade_minima", "min_quantity")
	initialIdx, hasInitial := firstCSVColumn(colIndex, "quantidade_inicial", "initial_quantity")

	field := func(record []string, idx int, ok bool) string {
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	var rows []service.StockImportRow
	var issues []service.StockImportIssue
	sourceRow := 1
	for {
		sourceRow++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			issues = append(issues, service.StockImportIssue{Row: sourceRow, Reason: "linha malformada"})
			continue
		}

		row := service.StockImportRow{
			SourceRow: sourceRow,
			Name:      field(record, nameIdx, hasName),
			Unit:      field(record, unitIdx, hasUnit),
		}

		invalid := false
		if raw := field(record, minIdx, hasMin); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil {
				issues = append(issues, service.StockImportIssue{Row: sourceRow, Name: row.Name, Reason: `"quantidade_minima" precisa ser um número inteiro`})
				invalid = true
			} else {
				row.MinQuantity = &n
			}
		}
		if raw := field(record, initialIdx, hasInitial); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil {
				issues = append(issues, service.StockImportIssue{Row: sourceRow, Name: row.Name, Reason: `"quantidade_inicial" precisa ser um número inteiro`})
				invalid = true
			} else {
				row.InitialQuantity = n
			}
		}
		if invalid {
			continue
		}
		rows = append(rows, row)
	}
	return rows, issues, nil
}

func firstCSVColumn(colIndex map[string]int, names ...string) (int, bool) {
	for _, n := range names {
		if idx, ok := colIndex[n]; ok {
			return idx, true
		}
	}
	return 0, false
}
