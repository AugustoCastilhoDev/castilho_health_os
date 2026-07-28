package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type FinancialService struct {
	rules        repository.FinancialRuleRepository
	transactions repository.FinancialTransactionRepository
}

func NewFinancialService(rules repository.FinancialRuleRepository, transactions repository.FinancialTransactionRepository) *FinancialService {
	return &FinancialService{rules: rules, transactions: transactions}
}

// validateRule enforces the invariants both CreateRule and UpdateRule need:
// a rule's Percentage/FixedAmountCents must actually match its Type, since
// the settlement service trusts whichever field its Type says to read
// (see calculatePayoutCents) without re-checking the other is empty.
func validateRule(rule *models.FinancialRule) error {
	if rule.ProfessionalID == uuid.Nil {
		return fmt.Errorf("%w: professional_id is required", ErrValidation)
	}
	switch rule.Type {
	case models.RuleTypePercentage:
		if rule.Percentage == nil || *rule.Percentage <= 0 || *rule.Percentage > 1 {
			return fmt.Errorf("%w: percentage must be between 0 and 1 (exclusive of 0) for PERCENTAGE rules", ErrValidation)
		}
	case models.RuleTypeFixedPerAppointment, models.RuleTypeFixedPerProcedure:
		if rule.FixedAmountCents == nil || *rule.FixedAmountCents <= 0 {
			return fmt.Errorf("%w: fixed_amount_cents must be positive for fixed-amount rules", ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unknown rule type %q", ErrValidation, rule.Type)
	}
	return nil
}

func (s *FinancialService) CreateRule(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error {
	if err := validateRule(rule); err != nil {
		return err
	}
	if rule.FeeDeduction == "" {
		rule.FeeDeduction = models.DeductFeesBeforeSplit
	}
	rule.IsActive = true
	return s.rules.Create(ctx, tenantID, rule)
}

func (s *FinancialService) GetRule(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialRule, error) {
	return s.rules.FindByID(ctx, tenantID, id)
}

// UpdateRule is a full replace, same PUT convention as the rest of the API
// — a field the caller omits (e.g. is_active) gets zeroed, so the frontend
// must always send the complete rule, not a partial patch.
func (s *FinancialService) UpdateRule(ctx context.Context, tenantID uuid.UUID, rule *models.FinancialRule) error {
	if err := validateRule(rule); err != nil {
		return err
	}
	if rule.FeeDeduction == "" {
		rule.FeeDeduction = models.DeductFeesBeforeSplit
	}
	return s.rules.Update(ctx, tenantID, rule)
}

func (s *FinancialService) ListRulesByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID) ([]models.FinancialRule, error) {
	return s.rules.ListByProfessional(ctx, tenantID, professionalID)
}

// CreateTransaction takes the requester's role because the allowed action
// depends on it: registering money coming IN (PATIENT_PAYMENT) is a normal
// front-desk task, but recording money going OUT to a professional
// (PROFESSIONAL_PAYOUT) is restricted to admin/finance — a route-level RBAC
// check can't see the request body's Type field, so this has to live here.
func (s *FinancialService) CreateTransaction(ctx context.Context, tenantID uuid.UUID, requesterRole models.UserRole, tx *models.FinancialTransaction) error {
	if tx.GrossAmountCents <= 0 {
		return fmt.Errorf("%w: gross_amount_cents must be positive", ErrValidation)
	}
	switch tx.Type {
	case models.TransactionPatientPayment:
	case models.TransactionProfessionalPayout:
		if requesterRole != models.RoleTenantAdmin && requesterRole != models.RoleFinance {
			return fmt.Errorf("%w: only TENANT_ADMIN or FINANCE may register a professional payout", ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unknown transaction type %q", ErrValidation, tx.Type)
	}
	if tx.NetAmountCents == 0 {
		tx.NetAmountCents = tx.GrossAmountCents - tx.FeeAmountCents
	}
	return s.transactions.Create(ctx, tenantID, tx)
}

func (s *FinancialService) GetTransaction(ctx context.Context, tenantID, id uuid.UUID) (*models.FinancialTransaction, error) {
	return s.transactions.FindByID(ctx, tenantID, id)
}

func (s *FinancialService) MarkPaid(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.transactions.MarkPaid(ctx, tenantID, id, time.Now())
}

func (s *FinancialService) ListTransactionsByAppointment(ctx context.Context, tenantID, appointmentID uuid.UUID) ([]models.FinancialTransaction, error) {
	return s.transactions.ListByAppointment(ctx, tenantID, appointmentID)
}

const (
	defaultTransactionPageSize = 20
	maxTransactionPageSize     = 100
)

// ListTransactions backs the Financeiro screen's ledger. Limit is defaulted
// and capped here rather than trusted from the request, since it flows
// straight into a SQL LIMIT.
func (s *FinancialService) ListTransactions(ctx context.Context, tenantID uuid.UUID, filter repository.TransactionFilter) ([]models.FinancialTransaction, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultTransactionPageSize
	} else if filter.Limit > maxTransactionPageSize {
		filter.Limit = maxTransactionPageSize
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.transactions.ListByTenant(ctx, tenantID, filter)
}
