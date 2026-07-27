package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

func TestSettlementService_Settle_NotReadyWhenAppointmentNotCompleted(t *testing.T) {
	tenantID, apptID := uuid.New(), uuid.New()
	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{Status: models.StatusInProgress}, nil
		},
	}
	svc := service.NewSettlementService(appointments, &fakeFinancialRuleRepo{}, &fakeFinancialTransactionRepo{})

	_, err := svc.Settle(context.Background(), tenantID, apptID)
	require.ErrorIs(t, err, service.ErrSettlementNotReady)
}

func TestSettlementService_Settle_NotReadyWhenNoPaidPatientPayment(t *testing.T) {
	tenantID, apptID := uuid.New(), uuid.New()
	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{Status: models.StatusCompleted}, nil
		},
	}
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			// A PENDING patient payment doesn't fund a payout yet.
			return []models.FinancialTransaction{
				{Type: models.TransactionPatientPayment, Status: models.TransactionPending},
			}, nil
		},
	}
	svc := service.NewSettlementService(appointments, &fakeFinancialRuleRepo{}, transactions)

	_, err := svc.Settle(context.Background(), tenantID, apptID)
	require.ErrorIs(t, err, service.ErrSettlementNotReady)
}

func TestSettlementService_Settle_IdempotentWhenPayoutAlreadyExists(t *testing.T) {
	tenantID, apptID, professionalID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	existingPayout := &models.FinancialTransaction{Type: models.TransactionProfessionalPayout, GrossAmountCents: 7000}

	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{ProfessionalID: professionalID, Status: models.StatusCompleted}, nil
		},
	}
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			return []models.FinancialTransaction{
				{TenantModel: modelWithID(sourceID), Type: models.TransactionPatientPayment, Status: models.TransactionPaid, GrossAmountCents: 10000, NetAmountCents: 10000},
			}, nil
		},
		findPayoutBySourceFn: func(ctx context.Context, tid, srcID uuid.UUID) (*models.FinancialTransaction, error) {
			assert.Equal(t, sourceID, srcID)
			return existingPayout, nil
		},
		// createFn intentionally unset — must not be called when a payout already exists.
	}
	svc := service.NewSettlementService(appointments, &fakeFinancialRuleRepo{}, transactions)

	got, err := svc.Settle(context.Background(), tenantID, apptID)
	require.NoError(t, err)
	assert.Same(t, existingPayout, got)
}

func TestSettlementService_Settle_NoApplicableRuleIsValidationError(t *testing.T) {
	tenantID, apptID, professionalID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{ProfessionalID: professionalID, Status: models.StatusCompleted}, nil
		},
	}
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			return []models.FinancialTransaction{
				{TenantModel: modelWithID(sourceID), Type: models.TransactionPatientPayment, Status: models.TransactionPaid, GrossAmountCents: 10000, NetAmountCents: 10000},
			}, nil
		},
		findPayoutBySourceFn: func(ctx context.Context, tid, srcID uuid.UUID) (*models.FinancialTransaction, error) {
			return nil, repository.ErrNotFound
		},
	}
	rules := &fakeFinancialRuleRepo{
		findApplicableFn: func(ctx context.Context, tid, profID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := service.NewSettlementService(appointments, rules, transactions)

	_, err := svc.Settle(context.Background(), tenantID, apptID)
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestSettlementService_Settle_PercentageBeforeSplitUsesNetAmount(t *testing.T) {
	tenantID, apptID, professionalID, sourceID, ruleID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	pct := 0.7

	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{ProfessionalID: professionalID, Status: models.StatusCompleted}, nil
		},
	}
	var created *models.FinancialTransaction
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			return []models.FinancialTransaction{
				// gross 10000, fee 500 -> net 9500; BEFORE_SPLIT computes the
				// professional's cut on the net (post-fee) amount.
				{TenantModel: modelWithID(sourceID), Type: models.TransactionPatientPayment, Status: models.TransactionPaid, GrossAmountCents: 10000, FeeAmountCents: 500, NetAmountCents: 9500},
			}, nil
		},
		findPayoutBySourceFn: func(ctx context.Context, tid, srcID uuid.UUID) (*models.FinancialTransaction, error) {
			return nil, repository.ErrNotFound
		},
		createFn: func(ctx context.Context, tid uuid.UUID, tx *models.FinancialTransaction) error {
			created = tx
			return nil
		},
	}
	rules := &fakeFinancialRuleRepo{
		findApplicableFn: func(ctx context.Context, tid, profID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
			return &models.FinancialRule{
				TenantModel:  modelWithID(ruleID),
				Type:         models.RuleTypePercentage,
				Percentage:   &pct,
				FeeDeduction: models.DeductFeesBeforeSplit,
			}, nil
		},
	}
	svc := service.NewSettlementService(appointments, rules, transactions)

	payout, err := svc.Settle(context.Background(), tenantID, apptID)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.EqualValues(t, 6650, payout.NetAmountCents) // round(9500 * 0.7)
	assert.EqualValues(t, 6650, created.GrossAmountCents)
	assert.Equal(t, models.TransactionProfessionalPayout, created.Type)
	assert.Equal(t, models.TransactionPending, created.Status)
	assert.Equal(t, sourceID, *created.SourceTransactionID)
	assert.Equal(t, ruleID, *created.FinancialRuleID)
	assert.Equal(t, professionalID, *created.ProfessionalID)
	assert.Equal(t, apptID, *created.AppointmentID)
}

func TestSettlementService_Settle_PercentageAfterSplitUsesGrossAmount(t *testing.T) {
	tenantID, apptID, professionalID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	pct := 0.5

	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{ProfessionalID: professionalID, Status: models.StatusCompleted}, nil
		},
	}
	var created *models.FinancialTransaction
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			return []models.FinancialTransaction{
				{TenantModel: modelWithID(sourceID), Type: models.TransactionPatientPayment, Status: models.TransactionPaid, GrossAmountCents: 10000, FeeAmountCents: 500, NetAmountCents: 9500},
			}, nil
		},
		findPayoutBySourceFn: func(ctx context.Context, tid, srcID uuid.UUID) (*models.FinancialTransaction, error) {
			return nil, repository.ErrNotFound
		},
		createFn: func(ctx context.Context, tid uuid.UUID, tx *models.FinancialTransaction) error {
			created = tx
			return nil
		},
	}
	rules := &fakeFinancialRuleRepo{
		findApplicableFn: func(ctx context.Context, tid, profID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
			return &models.FinancialRule{Type: models.RuleTypePercentage, Percentage: &pct, FeeDeduction: models.DeductFeesAfterSplit}, nil
		},
	}
	svc := service.NewSettlementService(appointments, rules, transactions)

	_, err := svc.Settle(context.Background(), tenantID, apptID)
	require.NoError(t, err)
	assert.EqualValues(t, 5000, created.GrossAmountCents) // round(10000 * 0.5), ignores the fee
}

func TestSettlementService_Settle_FixedPerAppointmentIgnoresFeePolicy(t *testing.T) {
	tenantID, apptID, professionalID, sourceID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	fixed := int64(4000)

	appointments := &fakeAppointmentRepo{
		findFn: func(ctx context.Context, tid, id uuid.UUID) (*models.Appointment, error) {
			return &models.Appointment{ProfessionalID: professionalID, Status: models.StatusCompleted}, nil
		},
	}
	var created *models.FinancialTransaction
	transactions := &fakeFinancialTransactionRepo{
		listByAppointmentFn: func(ctx context.Context, tid, id uuid.UUID) ([]models.FinancialTransaction, error) {
			return []models.FinancialTransaction{
				{TenantModel: modelWithID(sourceID), Type: models.TransactionPatientPayment, Status: models.TransactionPaid, GrossAmountCents: 30000, FeeAmountCents: 1000, NetAmountCents: 29000},
			}, nil
		},
		findPayoutBySourceFn: func(ctx context.Context, tid, srcID uuid.UUID) (*models.FinancialTransaction, error) {
			return nil, repository.ErrNotFound
		},
		createFn: func(ctx context.Context, tid uuid.UUID, tx *models.FinancialTransaction) error {
			created = tx
			return nil
		},
	}
	rules := &fakeFinancialRuleRepo{
		findApplicableFn: func(ctx context.Context, tid, profID uuid.UUID, procedureCode, insurancePlan *string) (*models.FinancialRule, error) {
			return &models.FinancialRule{Type: models.RuleTypeFixedPerAppointment, FixedAmountCents: &fixed, FeeDeduction: models.DeductFeesBeforeSplit}, nil
		},
	}
	svc := service.NewSettlementService(appointments, rules, transactions)

	_, err := svc.Settle(context.Background(), tenantID, apptID)
	require.NoError(t, err)
	assert.EqualValues(t, 4000, created.GrossAmountCents)
	assert.EqualValues(t, 4000, created.NetAmountCents)
}

func modelWithID(id uuid.UUID) models.TenantModel {
	return models.TenantModel{BaseModel: models.BaseModel{ID: id}}
}
