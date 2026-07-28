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

func TestMedicalRecordRepository_CreateAndListByPatient(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewMedicalRecordRepository(gdb)
	ctx := context.Background()

	record := &models.MedicalRecord{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		Type:           models.RecordTypeMedical,
		Content:        "Paciente relata dor de cabeça leve.",
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, record))

	found, err := repo.FindByID(ctx, tenant.ID, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "Paciente relata dor de cabeça leve.", found.Content)
	assert.False(t, found.IsLocked)

	list, err := repo.ListByPatient(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, record.ID, list[0].ID)
}

func TestMedicalRecordRepository_UpdateRejectedOnceLocked(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewMedicalRecordRepository(gdb)
	ctx := context.Background()

	record := &models.MedicalRecord{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		Type:           models.RecordTypeMedical,
		Content:        "rascunho inicial",
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, record))

	record.Content = "rascunho revisado"
	require.NoError(t, repo.Update(ctx, tenant.ID, record))
	found, err := repo.FindByID(ctx, tenant.ID, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "rascunho revisado", found.Content)

	locked, err := repo.Lock(ctx, tenant.ID, record.ID, doctor.ID)
	require.NoError(t, err)
	assert.True(t, locked.IsLocked)
	assert.NotNil(t, locked.LockedAt)
	require.NotNil(t, locked.LockedByID)
	assert.Equal(t, doctor.ID, *locked.LockedByID)

	record.Content = "tentativa de reescrever depois de travado"
	err = repo.Update(ctx, tenant.ID, record)
	require.ErrorIs(t, err, repository.ErrLocked)

	// Content on the row itself must be untouched by the rejected update.
	stillLocked, err := repo.FindByID(ctx, tenant.ID, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "rascunho revisado", stillLocked.Content)
}

func TestMedicalRecordRepository_LockIsIdempotent(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewMedicalRecordRepository(gdb)
	ctx := context.Background()

	record := &models.MedicalRecord{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		Type:           models.RecordTypeMedical,
		Content:        "nota",
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, record))

	first, err := repo.Lock(ctx, tenant.ID, record.ID, doctor.ID)
	require.NoError(t, err)
	second, err := repo.Lock(ctx, tenant.ID, record.ID, doctor.ID)
	require.NoError(t, err)
	assert.True(t, second.IsLocked)
	assert.Equal(t, *first.LockedByID, *second.LockedByID)
}
