package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

// ErrSettlementNotReady means the appointment hasn't yet met both gating
// conditions for an automatic repasse: Status == COMPLETED and a
// PATIENT_PAYMENT transaction with Status == PAID. It is not a failure —
// callers that trigger settlement opportunistically (on every COMPLETED
// transition, on every mark-paid) should ignore it.
var ErrSettlementNotReady = errors.New("settlement: appointment not yet eligible for payout")

// SettlementService generates a PROFESSIONAL_PAYOUT transaction once an
// appointment is COMPLETED and its funding PATIENT_PAYMENT is PAID. It reads
// from AppointmentRepository and FinancialRuleRepository directly, rather
// than depending on AppointmentService/FinancialService, so those two
// services stay free to be constructed (and unit-tested) without it.
type SettlementService struct {
	appointments repository.AppointmentRepository
	rules        repository.FinancialRuleRepository
	transactions repository.FinancialTransactionRepository
}

func NewSettlementService(appointments repository.AppointmentRepository, rules repository.FinancialRuleRepository, transactions repository.FinancialTransactionRepository) *SettlementService {
	return &SettlementService{appointments: appointments, rules: rules, transactions: transactions}
}

// Settle is idempotent: calling it repeatedly for the same appointment after
// a payout already exists just returns that payout again, so it's safe to
// call opportunistically from both the appointment-transition path and the
// mark-paid path without double-paying a professional.
func (s *SettlementService) Settle(ctx context.Context, tenantID, appointmentID uuid.UUID) (*models.FinancialTransaction, error) {
	appt, err := s.appointments.FindByID(ctx, tenantID, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.Status != models.StatusCompleted {
		return nil, ErrSettlementNotReady
	}

	txs, err := s.transactions.ListByAppointment(ctx, tenantID, appointmentID)
	if err != nil {
		return nil, err
	}
	var source *models.FinancialTransaction
	for i := range txs {
		if txs[i].Type == models.TransactionPatientPayment && txs[i].Status == models.TransactionPaid {
			source = &txs[i]
			break
		}
	}
	if source == nil {
		return nil, ErrSettlementNotReady
	}

	existing, err := s.transactions.FindPayoutBySource(ctx, tenantID, source.ID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	rule, err := s.rules.FindApplicable(ctx, tenantID, appt.ProfessionalID, source.ProcedureCode, source.InsurancePlan)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: no active financial rule configured for this professional/scope", ErrValidation)
		}
		return nil, err
	}

	payoutCents, err := calculatePayoutCents(rule, source)
	if err != nil {
		return nil, err
	}

	payout := &models.FinancialTransaction{
		AppointmentID:       &appointmentID,
		PatientID:           source.PatientID,
		ProfessionalID:      &appt.ProfessionalID,
		SourceTransactionID: &source.ID,
		FinancialRuleID:     &rule.ID,
		Type:                models.TransactionProfessionalPayout,
		Status:              models.TransactionPending,
		GrossAmountCents:    payoutCents,
		NetAmountCents:      payoutCents,
	}
	if err := s.transactions.Create(ctx, tenantID, payout); err != nil {
		return nil, err
	}
	return payout, nil
}

// calculatePayoutCents applies the rule's type and, for percentage rules,
// its FeeDeductionPolicy against the funding transaction. Fixed-amount
// rules ignore the fee policy entirely — a flat repasse per
// appointment/procedure is owed regardless of what the acquirer charged.
func calculatePayoutCents(rule *models.FinancialRule, source *models.FinancialTransaction) (int64, error) {
	switch rule.Type {
	case models.RuleTypePercentage:
		if rule.Percentage == nil {
			return 0, fmt.Errorf("%w: percentage rule %s has no percentage set", ErrValidation, rule.ID)
		}
		base := source.GrossAmountCents
		if rule.FeeDeduction == models.DeductFeesBeforeSplit {
			base = source.NetAmountCents
		}
		return int64(math.Round(float64(base) * (*rule.Percentage))), nil
	case models.RuleTypeFixedPerAppointment, models.RuleTypeFixedPerProcedure:
		if rule.FixedAmountCents == nil {
			return 0, fmt.Errorf("%w: fixed rule %s has no fixed_amount_cents set", ErrValidation, rule.ID)
		}
		return *rule.FixedAmountCents, nil
	default:
		return 0, fmt.Errorf("%w: unknown financial rule type %q", ErrValidation, rule.Type)
	}
}
