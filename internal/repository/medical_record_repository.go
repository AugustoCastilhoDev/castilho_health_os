package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// MedicalRecordRepository stores prontuário entries. Once a record is
// locked, Update must refuse to touch it — Lock is the only path back into
// mutating a record, and even that only reads/writes the lock fields
// themselves, never Content.
type MedicalRecordRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.MedicalRecord, error)
	ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MedicalRecord, error)
	// Update only ever affects a row that is still unlocked — enforced in
	// the same WHERE clause as the read, so a lock issued concurrently with
	// an edit can't be silently overwritten by a stale write.
	Update(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error
	// Lock finalizes a record so Update can never touch it again. It's a
	// no-op (not an error) if the record is already locked.
	Lock(ctx context.Context, tenantID, id, lockedByID uuid.UUID) (*models.MedicalRecord, error)
}

type medicalRecordRepository struct {
	db *gorm.DB
}

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

func (r *medicalRecordRepository) Create(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error {
	record.TenantID = tenantID
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.MedicalRecord, error) {
	var record models.MedicalRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *medicalRecordRepository) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MedicalRecord, error) {
	var records []models.MedicalRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *medicalRecordRepository) Update(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error {
	result := r.db.WithContext(ctx).
		Model(&models.MedicalRecord{}).
		Where("tenant_id = ? AND id = ? AND is_locked = false", tenantID, record.ID).
		Updates(map[string]any{
			"type":    record.Type,
			"content": record.Content,
			"cid":     record.CID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Zero rows matches either "doesn't exist" or "exists but locked" —
		// look it up to report the right one back to the caller.
		existing, err := r.FindByID(ctx, tenantID, record.ID)
		if err != nil {
			return err
		}
		if existing.IsLocked {
			return ErrLocked
		}
		return ErrNotFound
	}
	return nil
}

func (r *medicalRecordRepository) Lock(ctx context.Context, tenantID, id, lockedByID uuid.UUID) (*models.MedicalRecord, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.MedicalRecord{}).
		Where("tenant_id = ? AND id = ? AND is_locked = false", tenantID, id).
		Updates(map[string]any{
			"is_locked":    true,
			"locked_at":    now,
			"locked_by_id": lockedByID,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	// RowsAffected == 0 here just means it was already locked (or a
	// concurrent Lock call won the race) — either way the end state is the
	// same, so this returns the current row rather than erroring.
	return r.FindByID(ctx, tenantID, id)
}
