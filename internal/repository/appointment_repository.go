package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/castilho/health-os/internal/domain/models"
)

type AppointmentRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, appt *models.Appointment) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Appointment, error)
	ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID, from, to time.Time) ([]models.Appointment, error)
	ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.Appointment, error)
	// TransitionStatus is the only way status may change: it locks the row,
	// validates the move against models.AppointmentStatus.CanTransitionTo,
	// updates the denormalized timestamp for the new state, and appends an
	// AppointmentStatusLog row — all inside one transaction, so the audit
	// trail can never drift from the appointment's actual status.
	TransitionStatus(ctx context.Context, tenantID, id uuid.UUID, to models.AppointmentStatus, changedByID uuid.UUID, reason string) (*models.Appointment, error)
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

func (r *appointmentRepository) Create(ctx context.Context, tenantID uuid.UUID, appt *models.Appointment) error {
	appt.TenantID = tenantID
	if appt.Status == "" {
		appt.Status = models.StatusScheduled
	}
	return r.db.WithContext(ctx).Create(appt).Error
}

func (r *appointmentRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Appointment, error) {
	var appt models.Appointment
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &appt, nil
}

func (r *appointmentRepository) ListByProfessional(ctx context.Context, tenantID, professionalID uuid.UUID, from, to time.Time) ([]models.Appointment, error) {
	var appts []models.Appointment
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND professional_id = ? AND scheduled_at BETWEEN ? AND ?", tenantID, professionalID, from, to).
		Order("scheduled_at").
		Find(&appts).Error
	if err != nil {
		return nil, err
	}
	return appts, nil
}

func (r *appointmentRepository) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.Appointment, error) {
	var appts []models.Appointment
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("scheduled_at DESC").
		Find(&appts).Error
	if err != nil {
		return nil, err
	}
	return appts, nil
}

func (r *appointmentRepository) TransitionStatus(ctx context.Context, tenantID, id uuid.UUID, to models.AppointmentStatus, changedByID uuid.UUID, reason string) (*models.Appointment, error) {
	var result models.Appointment

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var appt models.Appointment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND id = ?", tenantID, id).
			First(&appt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		from := appt.Status
		if !from.CanTransitionTo(to) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
		}

		now := time.Now()
		appt.Status = to
		switch to {
		case models.StatusConfirmed:
			appt.ConfirmedAt = &now
		case models.StatusWaiting:
			appt.CheckedInAt = &now
		case models.StatusInProgress:
			appt.StartedAt = &now
		case models.StatusCompleted:
			appt.CompletedAt = &now
		case models.StatusCancelled:
			appt.CancelledAt = &now
			appt.CancellationReason = reason
		case models.StatusNoShow:
			appt.NoShowAt = &now
		}

		if err := tx.Model(&models.Appointment{}).
			Where("tenant_id = ? AND id = ?", tenantID, id).
			Select("*").
			Omit("id", "tenant_id", "created_at", "patient_id", "professional_id", "scheduled_at", "duration_min").
			Updates(&appt).Error; err != nil {
			return err
		}

		logEntry := models.AppointmentStatusLog{
			TenantModel:   models.TenantModel{TenantID: tenantID},
			AppointmentID: id,
			FromStatus:    from,
			ToStatus:      to,
			ChangedByID:   changedByID,
			Reason:        reason,
			ChangedAt:     now,
		}
		if err := tx.Create(&logEntry).Error; err != nil {
			return err
		}

		result = appt
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}
