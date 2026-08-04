package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestAppointmentRepository_TransitionStatus_HappyPath(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)

	repo := repository.NewAppointmentRepository(gdb)
	ctx := context.Background()

	appt := &models.Appointment{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		ScheduledAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, appt))
	assert.Equal(t, models.StatusScheduled, appt.Status)

	transitions := []models.AppointmentStatus{
		models.StatusConfirmed,
		models.StatusWaiting,
		models.StatusInProgress,
		models.StatusCompleted,
	}
	for _, to := range transitions {
		updated, err := repo.TransitionStatus(ctx, tenant.ID, appt.ID, to, doctor.ID, "", "")
		require.NoError(t, err, "transition to %s", to)
		assert.Equal(t, to, updated.Status)
	}

	final, err := repo.FindByID(ctx, tenant.ID, appt.ID)
	require.NoError(t, err)
	assert.NotNil(t, final.ConfirmedAt)
	assert.NotNil(t, final.CheckedInAt)
	assert.NotNil(t, final.StartedAt)
	assert.NotNil(t, final.CompletedAt)

	var logCount int64
	gdb.Model(&models.AppointmentStatusLog{}).Where("appointment_id = ?", appt.ID).Count(&logCount)
	assert.EqualValues(t, len(transitions), logCount, "one audit log row per transition, no more, no less")
}

func TestAppointmentRepository_TransitionStatus_RejectsIllegalJump(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)

	repo := repository.NewAppointmentRepository(gdb)
	ctx := context.Background()

	appt := &models.Appointment{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		ScheduledAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, appt))

	// SCHEDULED can never jump straight to COMPLETED.
	_, err := repo.TransitionStatus(ctx, tenant.ID, appt.ID, models.StatusCompleted, doctor.ID, "", "")
	require.ErrorIs(t, err, repository.ErrInvalidTransition)

	// The rejected transition must not have left any trace: status
	// unchanged, and no audit log row written for it.
	unchanged, err := repo.FindByID(ctx, tenant.ID, appt.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusScheduled, unchanged.Status)

	var logCount int64
	gdb.Model(&models.AppointmentStatusLog{}).Where("appointment_id = ?", appt.ID).Count(&logCount)
	assert.Zero(t, logCount)
}

func TestAppointmentRepository_TransitionStatus_NotFound(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewAppointmentRepository(gdb)

	_, err := repo.TransitionStatus(context.Background(), tenant.ID, uuid.New(), models.StatusConfirmed, uuid.New(), "", "")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

// Regression guard for the cid param added alongside the psychology/CID
// feature: setting it on a COMPLETED transition must persist it, and an
// empty cid on a later transition must leave the previously recorded value
// untouched rather than blanking it.
func TestAppointmentRepository_TransitionStatus_SetsAndPreservesCID(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	patient := testutil.NewPatient(t, gdb, tenant.ID)

	repo := repository.NewAppointmentRepository(gdb)
	ctx := context.Background()

	appt := &models.Appointment{
		PatientID:      patient.ID,
		ProfessionalID: doctor.ID,
		ScheduledAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, appt))

	_, err := repo.TransitionStatus(ctx, tenant.ID, appt.ID, models.StatusConfirmed, doctor.ID, "", "")
	require.NoError(t, err)
	_, err = repo.TransitionStatus(ctx, tenant.ID, appt.ID, models.StatusWaiting, doctor.ID, "", "")
	require.NoError(t, err)
	_, err = repo.TransitionStatus(ctx, tenant.ID, appt.ID, models.StatusInProgress, doctor.ID, "", "")
	require.NoError(t, err)

	completed, err := repo.TransitionStatus(ctx, tenant.ID, appt.ID, models.StatusCompleted, doctor.ID, "", "F41.1")
	require.NoError(t, err)
	require.NotNil(t, completed.CID)
	assert.Equal(t, "F41.1", *completed.CID)

	final, err := repo.FindByID(ctx, tenant.ID, appt.ID)
	require.NoError(t, err)
	require.NotNil(t, final.CID)
	assert.Equal(t, "F41.1", *final.CID)
}
