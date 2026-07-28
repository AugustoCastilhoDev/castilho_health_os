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

func TestMemedPrescriptionLogRepository_CreateFindListCancel(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewMemedPrescriptionLogRepository(gdb)
	ctx := context.Background()

	log := &models.MemedPrescriptionLog{
		PatientID:           patient.ID,
		ProfessionalID:      doctor.ID,
		MemedPrescriptionID: "memed-" + tenant.ID.String(),
		Status:              models.MemedPrescriptionIssued,
		IssuedAt:            time.Now(),
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, log))

	found, err := repo.FindByMemedID(ctx, tenant.ID, log.MemedPrescriptionID)
	require.NoError(t, err)
	assert.Equal(t, log.ID, found.ID)
	assert.Equal(t, models.MemedPrescriptionIssued, found.Status)

	list, err := repo.ListByPatient(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, log.ID, list[0].ID)

	require.NoError(t, repo.CancelByMemedID(ctx, tenant.ID, log.MemedPrescriptionID))
	found, err = repo.FindByMemedID(ctx, tenant.ID, log.MemedPrescriptionID)
	require.NoError(t, err)
	assert.Equal(t, models.MemedPrescriptionCancelled, found.Status)

	err = repo.CancelByMemedID(ctx, tenant.ID, "does-not-exist")
	require.ErrorIs(t, err, repository.ErrNotFound)

	_, err = repo.FindByMemedID(ctx, tenant.ID, "does-not-exist")
	require.ErrorIs(t, err, repository.ErrNotFound)
}
