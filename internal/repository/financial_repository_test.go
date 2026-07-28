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

func TestFinancialTransactionRepository_ListByTenant_FiltersAndIsTenantScoped(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenantA := testutil.NewTenant(t, gdb)
	tenantB := testutil.NewTenant(t, gdb)
	doctorA := testutil.NewUser(t, gdb, tenantA.ID, models.RoleDoctor)
	doctorB := testutil.NewUser(t, gdb, tenantB.ID, models.RoleDoctor)
	repo := repository.NewFinancialTransactionRepository(gdb)
	ctx := context.Background()

	pendingIncome := &models.FinancialTransaction{
		ProfessionalID:   &doctorA.ID,
		Type:             models.TransactionPatientPayment,
		GrossAmountCents: 10000,
		NetAmountCents:   10000,
	}
	require.NoError(t, repo.Create(ctx, tenantA.ID, pendingIncome))

	paidPayout := &models.FinancialTransaction{
		ProfessionalID:   &doctorA.ID,
		Type:             models.TransactionProfessionalPayout,
		GrossAmountCents: 7000,
		NetAmountCents:   7000,
		Status:           models.TransactionPaid,
	}
	require.NoError(t, repo.Create(ctx, tenantA.ID, paidPayout))

	otherTenantIncome := &models.FinancialTransaction{
		ProfessionalID:   &doctorB.ID,
		Type:             models.TransactionPatientPayment,
		GrossAmountCents: 99999,
		NetAmountCents:   99999,
	}
	require.NoError(t, repo.Create(ctx, tenantB.ID, otherTenantIncome))

	// No filter: only tenantA's two rows, never tenantB's.
	all, total, err := repo.ListByTenant(ctx, tenantA.ID, repository.TransactionFilter{Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, all, 2)

	// Filter by type: only the PATIENT_PAYMENT row.
	incomeType := models.TransactionPatientPayment
	incomes, total, err := repo.ListByTenant(ctx, tenantA.ID, repository.TransactionFilter{Type: &incomeType, Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, incomes, 1)
	assert.Equal(t, pendingIncome.ID, incomes[0].ID)

	// Filter by status: only the PAID payout.
	paidStatus := models.TransactionPaid
	paid, total, err := repo.ListByTenant(ctx, tenantA.ID, repository.TransactionFilter{Status: &paidStatus, Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, paid, 1)
	assert.Equal(t, paidPayout.ID, paid[0].ID)

	// Pagination: limit 1 still reports the true total of 2.
	page, total, err := repo.ListByTenant(ctx, tenantA.ID, repository.TransactionFilter{Limit: 1, Offset: 0})
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, page, 1)
}
