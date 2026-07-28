package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// MemedPrescriptionLogRepository is an audit trail, not a system of record
// (see the model's doc comment) — there is deliberately no Update-by-ID or
// hard Delete here, only the narrow CancelByMemedID transition a
// prescricaoExcluida event can trigger.
type MemedPrescriptionLogRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, log *models.MemedPrescriptionLog) error
	FindByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error)
	ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MemedPrescriptionLog, error)
	CancelByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) error
}

type memedPrescriptionLogRepository struct {
	db *gorm.DB
}

func NewMemedPrescriptionLogRepository(db *gorm.DB) MemedPrescriptionLogRepository {
	return &memedPrescriptionLogRepository{db: db}
}

func (r *memedPrescriptionLogRepository) Create(ctx context.Context, tenantID uuid.UUID, log *models.MemedPrescriptionLog) error {
	log.TenantID = tenantID
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *memedPrescriptionLogRepository) FindByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error) {
	var log models.MemedPrescriptionLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND memed_prescription_id = ?", tenantID, memedPrescriptionID).
		First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

func (r *memedPrescriptionLogRepository) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MemedPrescriptionLog, error) {
	var logs []models.MemedPrescriptionLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("issued_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *memedPrescriptionLogRepository) CancelByMemedID(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) error {
	result := r.db.WithContext(ctx).
		Model(&models.MemedPrescriptionLog{}).
		Where("tenant_id = ? AND memed_prescription_id = ?", tenantID, memedPrescriptionID).
		Update("status", models.MemedPrescriptionCancelled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
