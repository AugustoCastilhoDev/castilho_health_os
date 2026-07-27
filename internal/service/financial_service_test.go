package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestFinancialService_CreateRule_Validation(t *testing.T) {
	badPct := 1.5
	fixedRuleNoAmount := &models.FinancialRule{ProfessionalID: uuid.New(), Type: models.RuleTypeFixedPerAppointment}

	cases := []struct {
		name string
		rule *models.FinancialRule
	}{
		{"missing professional_id", &models.FinancialRule{Type: models.RuleTypePercentage}},
		{"percentage out of range", &models.FinancialRule{ProfessionalID: uuid.New(), Type: models.RuleTypePercentage, Percentage: &badPct}},
		{"percentage nil", &models.FinancialRule{ProfessionalID: uuid.New(), Type: models.RuleTypePercentage}},
		{"fixed amount missing", fixedRuleNoAmount},
		{"unknown type", &models.FinancialRule{ProfessionalID: uuid.New(), Type: "BOGUS"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeFinancialRuleRepo{}
			svc := service.NewFinancialService(repo, &fakeFinancialTransactionRepo{})
			err := svc.CreateRule(context.Background(), uuid.New(), tc.rule)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestFinancialService_CreateTransaction_PayoutRestrictedToAdminOrFinance(t *testing.T) {
	cases := []struct {
		role      models.UserRole
		wantError bool
	}{
		{models.RoleReceptionist, true},
		{models.RoleDoctor, true},
		{models.RoleTenantAdmin, false},
		{models.RoleFinance, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			repo := &fakeFinancialTransactionRepo{
				createFn: func(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error {
					return nil
				},
			}
			svc := service.NewFinancialService(&fakeFinancialRuleRepo{}, repo)

			tx := &models.FinancialTransaction{
				Type:             models.TransactionProfessionalPayout,
				GrossAmountCents: 14000,
			}
			err := svc.CreateTransaction(context.Background(), uuid.New(), tc.role, tx)
			if tc.wantError {
				require.ErrorIs(t, err, service.ErrValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFinancialService_CreateTransaction_PatientPaymentOpenToAnyRole(t *testing.T) {
	repo := &fakeFinancialTransactionRepo{
		createFn: func(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error { return nil },
	}
	svc := service.NewFinancialService(&fakeFinancialRuleRepo{}, repo)

	tx := &models.FinancialTransaction{Type: models.TransactionPatientPayment, GrossAmountCents: 20000}
	err := svc.CreateTransaction(context.Background(), uuid.New(), models.RoleReceptionist, tx)
	require.NoError(t, err)
}

func TestFinancialService_CreateTransaction_AutoCalculatesNetAmount(t *testing.T) {
	var saved *models.FinancialTransaction
	repo := &fakeFinancialTransactionRepo{
		createFn: func(ctx context.Context, tenantID uuid.UUID, tx *models.FinancialTransaction) error {
			saved = tx
			return nil
		},
	}
	svc := service.NewFinancialService(&fakeFinancialRuleRepo{}, repo)

	tx := &models.FinancialTransaction{
		Type:             models.TransactionPatientPayment,
		GrossAmountCents: 20000,
		FeeAmountCents:   500,
	}
	require.NoError(t, svc.CreateTransaction(context.Background(), uuid.New(), models.RoleReceptionist, tx))
	assert.EqualValues(t, 19500, saved.NetAmountCents)
}

func TestFinancialService_CreateTransaction_RejectsNonPositiveAmount(t *testing.T) {
	repo := &fakeFinancialTransactionRepo{}
	svc := service.NewFinancialService(&fakeFinancialRuleRepo{}, repo)

	tx := &models.FinancialTransaction{Type: models.TransactionPatientPayment, GrossAmountCents: 0}
	err := svc.CreateTransaction(context.Background(), uuid.New(), models.RoleReceptionist, tx)
	require.ErrorIs(t, err, service.ErrValidation)
}
