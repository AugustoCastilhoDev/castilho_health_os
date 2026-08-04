package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

type OdontogramaRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, entry *models.OdontogramaEntry) error
	ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error)
	// CurrentChart returns the single most recent entry per tooth_number for
	// the patient — the derived chart the frontend renders. There's no
	// current-state table to read from (see OdontogramaEntry's doc comment),
	// so this is a DISTINCT ON query instead of a plain Find.
	CurrentChart(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error)
}

type odontogramaRepository struct {
	db *gorm.DB
}

func NewOdontogramaRepository(db *gorm.DB) OdontogramaRepository {
	return &odontogramaRepository{db: db}
}

func (r *odontogramaRepository) Create(ctx context.Context, tenantID uuid.UUID, entry *models.OdontogramaEntry) error {
	entry.TenantID = tenantID
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *odontogramaRepository) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	var entries []models.OdontogramaEntry
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("created_at DESC").
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *odontogramaRepository) CurrentChart(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	var entries []models.OdontogramaEntry
	err := r.db.WithContext(ctx).
		Select("DISTINCT ON (tooth_number) *").
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("tooth_number, created_at DESC").
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}
