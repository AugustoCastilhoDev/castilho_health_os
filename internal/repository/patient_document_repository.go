package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// PatientDocumentRepository stores metadata only — the file bytes live in
// R2 (see internal/storage), reached only through a presigned URL minted by
// PatientDocumentService.
type PatientDocumentRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, doc *models.PatientDocument) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error)
	ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.PatientDocument, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

type patientDocumentRepository struct {
	db *gorm.DB
}

func NewPatientDocumentRepository(db *gorm.DB) PatientDocumentRepository {
	return &patientDocumentRepository{db: db}
}

func (r *patientDocumentRepository) Create(ctx context.Context, tenantID uuid.UUID, doc *models.PatientDocument) error {
	doc.TenantID = tenantID
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *patientDocumentRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error) {
	var doc models.PatientDocument
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (r *patientDocumentRepository) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.PatientDocument, error) {
	var docs []models.PatientDocument
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND patient_id = ?", tenantID, patientID).
		Order("created_at DESC").
		Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *patientDocumentRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&models.PatientDocument{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
