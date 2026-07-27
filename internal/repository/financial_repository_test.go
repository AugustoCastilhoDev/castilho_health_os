package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestFinancialRuleRepository_FindApplicable_PrefersHigherPriority(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	repo := repository.NewFinancialRuleRepository(gdb)
	ctx := context.Background()

	wildcardPct := 0.5
	wildcard := &models.FinancialRule{
		ProfessionalID: doctor.ID,
		Type:           models.RuleTypePercentage,
		Percentage:     &wildcardPct,
		FeeDeduction:   models.DeductFeesBeforeSplit,
		Priority:       0,
		IsActive:       true,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, wildcard))

	procedureCode := "CONSULTA"
	specificPct := 0.8
	specific := &models.FinancialRule{
		ProfessionalID: doctor.ID,
		Type:           models.RuleTypePercentage,
		Percentage:     &specificPct,
		ProcedureCode:  &procedureCode,
		FeeDeduction:   models.DeductFeesBeforeSplit,
		Priority:       10,
		IsActive:       true,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, specific))

	// A transaction naming the specific procedure should resolve to the
	// higher-priority specific rule, not the wildcard.
	found, err := repo.FindApplicable(ctx, tenant.ID, doctor.ID, &procedureCode, nil)
	require.NoError(t, err)
	assert.Equal(t, specific.ID, found.ID)

	// No procedure code given: only the wildcard rule can match.
	found, err = repo.FindApplicable(ctx, tenant.ID, doctor.ID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, wildcard.ID, found.ID)
}

// Regression test for the exact bug caught during manual testing: a
// PROFESSIONAL_PAYOUT created without a payment method must serialize
// PaymentMethod as SQL NULL, not the empty string — the empty string
// violates the payment_method CHECK constraint (migration 000005).
func TestFinancialTransactionRepository_NilPaymentMethodDoesNotViolateCheckConstraint(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	repo := repository.NewFinancialTransactionRepository(gdb)
	ctx := context.Background()

	payout := &models.FinancialTransaction{
		ProfessionalID:   &doctor.ID,
		Type:             models.TransactionProfessionalPayout,
		GrossAmountCents: 14000,
		NetAmountCents:   14000,
		PaymentMethod:    nil,
	}
	err := repo.Create(ctx, tenant.ID, payout)
	require.NoError(t, err, "creating a transaction with a nil PaymentMethod must not violate the CHECK constraint")
}

func TestFinancialTransactionRepository_FindPayoutBySource_Idempotency(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	repo := repository.NewFinancialTransactionRepository(gdb)
	ctx := context.Background()

	income := &models.FinancialTransaction{
		ProfessionalID:   &doctor.ID,
		Type:             models.TransactionPatientPayment,
		GrossAmountCents: 20000,
		NetAmountCents:   20000,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, income))

	// Before any payout exists, lookup must report ErrNotFound so a
	// settlement service knows it's safe to create one.
	_, err := repo.FindPayoutBySource(ctx, tenant.ID, income.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)

	payout := &models.FinancialTransaction{
		ProfessionalID:      &doctor.ID,
		SourceTransactionID: &income.ID,
		Type:                models.TransactionProfessionalPayout,
		GrossAmountCents:    14000,
		NetAmountCents:      14000,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, payout))

	found, err := repo.FindPayoutBySource(ctx, tenant.ID, income.ID)
	require.NoError(t, err)
	assert.Equal(t, payout.ID, found.ID)
}
