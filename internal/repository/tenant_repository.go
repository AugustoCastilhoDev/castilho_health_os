package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// TenantRepository is intentionally not tenant-scoped — it manages the
// tenants themselves (onboarding, subdomain/slug resolution at login).
type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
	List(ctx context.Context) ([]models.Tenant, error)
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	result := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenant.ID).
		Select("*").
		Omit("id", "created_at").
		Updates(tenant)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *tenantRepository) List(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := r.db.WithContext(ctx).Order("name").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
