package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appmiddleware "github.com/castilho/health-os/internal/api/middleware"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

type FinancialHandler struct {
	financial *service.FinancialService
}

func NewFinancialHandler(s *service.FinancialService) *FinancialHandler {
	return &FinancialHandler{financial: s}
}

type financialRuleRequest struct {
	ProfessionalID   uuid.UUID `json:"professional_id"`
	Type             string    `json:"type"`
	Percentage       *float64  `json:"percentage"`
	FixedAmountCents *int64    `json:"fixed_amount_cents"`
	ProcedureCode    *string   `json:"procedure_code"`
	InsurancePlan    *string   `json:"insurance_plan"`
	FeeDeduction     string    `json:"fee_deduction"`
	Priority         int       `json:"priority"`
}

func (r financialRuleRequest) toModel() *models.FinancialRule {
	return &models.FinancialRule{
		ProfessionalID:   r.ProfessionalID,
		Type:             models.FinancialRuleType(r.Type),
		Percentage:       r.Percentage,
		FixedAmountCents: r.FixedAmountCents,
		ProcedureCode:    r.ProcedureCode,
		InsurancePlan:    r.InsurancePlan,
		FeeDeduction:     models.FeeDeductionPolicy(r.FeeDeduction),
		Priority:         r.Priority,
	}
}

func (h *FinancialHandler) CreateRule(c *fiber.Ctx) error {
	var req financialRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	rule := req.toModel()
	if err := h.financial.CreateRule(c.Context(), appmiddleware.TenantID(c), rule); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rule)
}

func (h *FinancialHandler) GetRule(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	rule, err := h.financial.GetRule(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(rule)
}

func (h *FinancialHandler) ListRulesByProfessional(c *fiber.Ctx) error {
	professionalID, err := uuid.Parse(c.Query("professional_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "professional_id is required and must be a valid UUID"})
	}
	rules, err := h.financial.ListRulesByProfessional(c.Context(), appmiddleware.TenantID(c), professionalID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(rules)
}

type financialTransactionRequest struct {
	AppointmentID    *uuid.UUID `json:"appointment_id"`
	PatientID        *uuid.UUID `json:"patient_id"`
	ProfessionalID   *uuid.UUID `json:"professional_id"`
	Type             string     `json:"type"`
	GrossAmountCents int64      `json:"gross_amount_cents"`
	FeeAmountCents   int64      `json:"fee_amount_cents"`
	PaymentMethod    *string    `json:"payment_method"`
	Notes            string     `json:"notes"`
}

func (h *FinancialHandler) CreateTransaction(c *fiber.Ctx) error {
	var req financialTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var pm *models.PaymentMethod
	if req.PaymentMethod != nil {
		v := models.PaymentMethod(*req.PaymentMethod)
		pm = &v
	}

	tx := &models.FinancialTransaction{
		AppointmentID:    req.AppointmentID,
		PatientID:        req.PatientID,
		ProfessionalID:   req.ProfessionalID,
		Type:             models.TransactionType(req.Type),
		GrossAmountCents: req.GrossAmountCents,
		FeeAmountCents:   req.FeeAmountCents,
		PaymentMethod:    pm,
		Notes:            req.Notes,
	}

	if err := h.financial.CreateTransaction(c.Context(), appmiddleware.TenantID(c), appmiddleware.Role(c), tx); err != nil {
		return respondErr(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(tx)
}

func (h *FinancialHandler) GetTransaction(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	tx, err := h.financial.GetTransaction(c.Context(), appmiddleware.TenantID(c), id)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(tx)
}

func (h *FinancialHandler) MarkPaid(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.financial.MarkPaid(c.Context(), appmiddleware.TenantID(c), id); err != nil {
		return respondErr(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *FinancialHandler) ListTransactionsByAppointment(c *fiber.Ctx) error {
	appointmentID, err := uuid.Parse(c.Params("appointmentID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid appointment id"})
	}
	txs, err := h.financial.ListTransactionsByAppointment(c.Context(), appmiddleware.TenantID(c), appointmentID)
	if err != nil {
		return respondErr(c, err)
	}
	return c.JSON(txs)
}
