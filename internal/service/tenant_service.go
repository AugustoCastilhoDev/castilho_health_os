package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type RegisterTenantInput struct {
	TenantName     string
	TenantSlug     string
	TenantType     models.TenantType
	TenantDocument string
	TenantEmail    string
	TenantPhone    string
	AdminName      string
	AdminEmail     string
	AdminPassword  string
}

type TenantService struct {
	db      *gorm.DB
	tenants repository.TenantRepository
}

func NewTenantService(db *gorm.DB, tenants repository.TenantRepository) *TenantService {
	return &TenantService{db: db, tenants: tenants}
}

// Register onboards a brand-new clinic: creates the Tenant and its first
// TENANT_ADMIN user in a single transaction — a tenant without an admin
// (or vice versa) must never be observable by any other request.
func (s *TenantService) Register(ctx context.Context, in RegisterTenantInput) (*models.Tenant, *models.User, error) {
	if in.TenantName == "" || in.TenantSlug == "" || in.TenantDocument == "" || in.TenantEmail == "" {
		return nil, nil, fmt.Errorf("%w: tenant_name, tenant_slug, tenant_document and tenant_email are required", ErrValidation)
	}
	if in.AdminName == "" || in.AdminEmail == "" || len(in.AdminPassword) < 8 {
		return nil, nil, fmt.Errorf("%w: admin_name, admin_email and an admin_password of at least 8 characters are required", ErrValidation)
	}
	switch in.TenantType {
	case models.TenantTypeMedica, models.TenantTypeOdonto, models.TenantTypeMista:
	default:
		return nil, nil, fmt.Errorf("%w: tenant_type must be MEDICA, ODONTO or MISTA", ErrValidation)
	}

	if _, err := s.tenants.FindBySlug(ctx, in.TenantSlug); err == nil {
		return nil, nil, fmt.Errorf("%w: slug %q is already in use", ErrConflict, in.TenantSlug)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, nil, err
	}

	passwordHash, err := auth.HashPassword(in.AdminPassword)
	if err != nil {
		return nil, nil, err
	}

	tenant := &models.Tenant{
		Name:     in.TenantName,
		Slug:     in.TenantSlug,
		Type:     in.TenantType,
		Document: in.TenantDocument,
		Email:    in.TenantEmail,
		Phone:    in.TenantPhone,
		IsActive: true,
	}
	admin := &models.User{
		Name:         in.AdminName,
		Email:        in.AdminEmail,
		PasswordHash: passwordHash,
		Role:         models.RoleTenantAdmin,
		IsActive:     true,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := repository.NewTenantRepository(tx).Create(ctx, tenant); err != nil {
			return err
		}
		return repository.NewUserRepository(tx).Create(ctx, tenant.ID, admin)
	})
	if err != nil {
		return nil, nil, err
	}

	return tenant, admin, nil
}
