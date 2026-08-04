package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestOdontogramaRepository_CreateAndListByPatient(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	dentist := testutil.NewUser(t, gdb, tenant.ID, models.RoleDentist)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewOdontogramaRepository(gdb)
	ctx := context.Background()

	entry := &models.OdontogramaEntry{
		PatientID:    patient.ID,
		ToothNumber:  "26",
		Condition:    models.ToothConditionCavity,
		RecordedByID: dentist.ID,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, entry))

	list, err := repo.ListByPatient(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "26", list[0].ToothNumber)
	assert.Equal(t, models.ToothConditionCavity, list[0].Condition)
}

// CurrentChart must return only the latest entry per tooth, not the full
// history — this is the regression guard for the DISTINCT ON query, since a
// bare Find would return every entry for teeth touched more than once.
func TestOdontogramaRepository_CurrentChart_ReturnsLatestEntryPerTooth(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	dentist := testutil.NewUser(t, gdb, tenant.ID, models.RoleDentist)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewOdontogramaRepository(gdb)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, tenant.ID, &models.OdontogramaEntry{
		PatientID:    patient.ID,
		ToothNumber:  "36",
		Condition:    models.ToothConditionCavity,
		RecordedByID: dentist.ID,
	}))
	time.Sleep(10 * time.Millisecond) // ensure a distinct created_at ordering
	require.NoError(t, repo.Create(ctx, tenant.ID, &models.OdontogramaEntry{
		PatientID:    patient.ID,
		ToothNumber:  "36",
		Condition:    models.ToothConditionRestored,
		RecordedByID: dentist.ID,
	}))
	require.NoError(t, repo.Create(ctx, tenant.ID, &models.OdontogramaEntry{
		PatientID:    patient.ID,
		ToothNumber:  "47",
		Condition:    models.ToothConditionMissing,
		RecordedByID: dentist.ID,
	}))

	chart, err := repo.CurrentChart(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	require.Len(t, chart, 2)

	byTooth := make(map[string]models.ToothCondition, len(chart))
	for _, e := range chart {
		byTooth[e.ToothNumber] = e.Condition
	}
	assert.Equal(t, models.ToothConditionRestored, byTooth["36"])
	assert.Equal(t, models.ToothConditionMissing, byTooth["47"])

	full, err := repo.ListByPatient(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	assert.Len(t, full, 3, "full history must still contain every entry")
}
