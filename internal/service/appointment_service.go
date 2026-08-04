package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type AppointmentService struct {
	appointments repository.AppointmentRepository
}

func NewAppointmentService(appointments repository.AppointmentRepository) *AppointmentService {
	return &AppointmentService{appointments: appointments}
}

func (s *AppointmentService) Create(ctx context.Context, tenantID uuid.UUID, appt *models.Appointment) error {
	if appt.PatientID == uuid.Nil || appt.ProfessionalID == uuid.Nil {
		return fmt.Errorf("%w: patient_id and professional_id are required", ErrValidation)
	}
	if appt.ScheduledAt.IsZero() {
		return fmt.Errorf("%w: scheduled_at is required", ErrValidation)
	}
	if appt.DurationMin <= 0 {
		appt.DurationMin = 30
	}
	return s.appointments.Create(ctx, tenantID, appt)
}

func (s *AppointmentService) Get(ctx context.Context, tenantID, id uuid.UUID) (*models.Appointment, error) {
	return s.appointments.FindByID(ctx, tenantID, id)
}

func (s *AppointmentService) ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID, from, to time.Time) ([]models.Appointment, error) {
	return s.appointments.ListByProfessional(ctx, tenantID, professionalID, from, to)
}

func (s *AppointmentService) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.Appointment, error) {
	return s.appointments.ListByPatient(ctx, tenantID, patientID)
}

// Transition is the only way a handler may change an appointment's status;
// it just forwards to the repository, whose transaction enforces the state
// machine and writes the audit log — this layer only owns input shape
// validation (unknown status, cancel without a reason).
func (s *AppointmentService) Transition(ctx context.Context, tenantID, id uuid.UUID, to models.AppointmentStatus, changedByID uuid.UUID, reason, cid string) (*models.Appointment, error) {
	switch to {
	case models.StatusScheduled, models.StatusConfirmed, models.StatusCancelled, models.StatusWaiting, models.StatusInProgress, models.StatusCompleted, models.StatusNoShow:
	default:
		return nil, fmt.Errorf("%w: unknown status %q", ErrValidation, to)
	}
	if to == models.StatusCancelled && reason == "" {
		return nil, fmt.Errorf("%w: reason is required to cancel an appointment", ErrValidation)
	}
	return s.appointments.TransitionStatus(ctx, tenantID, id, to, changedByID, reason, cid)
}
