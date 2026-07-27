package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

type PatientRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Patient, error)
	Update(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// Search matches name (case-insensitive, partial) or an exact document
	// (CPF) match — the two lookups receptionists actually use at the front
	// desk.
	Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error)
}

type patientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepository{db: db}
}

func (r *patientRepository) Create(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error {
	patient.TenantID = tenantID
	return r.db.WithContext(ctx).Create(patient).Error
}

func (r *patientRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) Update(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error {
	patient.TenantID = tenantID
	result := r.db.WithContext(ctx).
		Model(&models.Patient{}).
		Where("tenant_id = ? AND id = ?", tenantID, patient.ID).
		Select("*").
		Omit("id", "tenant_id", "created_at").
		Updates(patient)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *patientRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&models.Patient{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *patientRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error) {
	var patients []models.Patient
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (name ILIKE ? OR document = ?)", tenantID, "%"+query+"%", query).
		Order("name").
		Limit(limit).
		Offset(offset).
		Find(&patients).Error
	if err != nil {
		return nil, err
	}
	return patients, nil
}
