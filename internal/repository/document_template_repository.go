package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// DocumentTemplateRepository stores clinic-level, reusable document bodies
// (atestados, declarações, laudos) — tenant-scoped only, no patient/user FK.
type DocumentTemplateRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error)
	Update(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.DocumentTemplate, error)
}

type documentTemplateRepository struct {
	db *gorm.DB
}

func NewDocumentTemplateRepository(db *gorm.DB) DocumentTemplateRepository {
	return &documentTemplateRepository{db: db}
}

func (r *documentTemplateRepository) Create(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error {
	tmpl.TenantID = tenantID
	return r.db.WithContext(ctx).Create(tmpl).Error
}

func (r *documentTemplateRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
	var tmpl models.DocumentTemplate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&tmpl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tmpl, nil
}

func (r *documentTemplateRepository) Update(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error {
	result := r.db.WithContext(ctx).
		Model(&models.DocumentTemplate{}).
		Where("tenant_id = ? AND id = ?", tenantID, tmpl.ID).
		Select("*").
		Omit("id", "tenant_id", "created_at").
		Updates(tmpl)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *documentTemplateRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.DocumentTemplate, error) {
	var templates []models.DocumentTemplate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("name").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}
