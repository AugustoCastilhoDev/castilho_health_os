package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestAppointmentService_Create_Validation(t *testing.T) {
	repo := &fakeAppointmentRepo{} // no createFn — must not be called on validation failure

	cases := []struct {
		name string
		appt *models.Appointment
	}{
		{"missing patient_id", &models.Appointment{ProfessionalID: uuid.New(), ScheduledAt: time.Now()}},
		{"missing professional_id", &models.Appointment{PatientID: uuid.New(), ScheduledAt: time.Now()}},
		{"missing scheduled_at", &models.Appointment{PatientID: uuid.New(), ProfessionalID: uuid.New()}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewAppointmentService(repo)
			err := svc.Create(context.Background(), uuid.New(), tc.appt)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestAppointmentService_Create_DefaultsDuration(t *testing.T) {
	var created *models.Appointment
	repo := &fakeAppointmentRepo{
		createFn: func(ctx context.Context, tenantID uuid.UUID, a *models.Appointment) error {
			created = a
			return nil
		},
	}
	svc := service.NewAppointmentService(repo)

	appt := &models.Appointment{PatientID: uuid.New(), ProfessionalID: uuid.New(), ScheduledAt: time.Now()}
	require.NoError(t, svc.Create(context.Background(), uuid.New(), appt))
	assert.Equal(t, 30, created.DurationMin)
}

func TestAppointmentService_Transition_RejectsUnknownStatus(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo)

	_, err := svc.Transition(context.Background(), uuid.New(), uuid.New(), models.AppointmentStatus("NOT_A_REAL_STATUS"), uuid.New(), "")
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestAppointmentService_Transition_CancelRequiresReason(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo)

	_, err := svc.Transition(context.Background(), uuid.New(), uuid.New(), models.StatusCancelled, uuid.New(), "")
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestAppointmentService_Transition_ForwardsValidCallToRepo(t *testing.T) {
	tenantID, apptID, changedBy := uuid.New(), uuid.New(), uuid.New()
	var gotReason string
	repo := &fakeAppointmentRepo{
		transitionFn: func(ctx context.Context, tid, id uuid.UUID, to models.AppointmentStatus, cb uuid.UUID, reason string) (*models.Appointment, error) {
			assert.Equal(t, tenantID, tid)
			assert.Equal(t, apptID, id)
			assert.Equal(t, models.StatusCancelled, to)
			assert.Equal(t, changedBy, cb)
			gotReason = reason
			return &models.Appointment{Status: to}, nil
		},
	}
	svc := service.NewAppointmentService(repo)

	result, err := svc.Transition(context.Background(), tenantID, apptID, models.StatusCancelled, changedBy, "patient requested")
	require.NoError(t, err)
	assert.Equal(t, models.StatusCancelled, result.Status)
	assert.Equal(t, "patient requested", gotReason)
}
