package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
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

const logoPresignExpiry = 15 * time.Minute

// TenantService is usable with storage == nil (e.g. no R2 account
// configured yet), same nil-safe convention as PatientDocumentService —
// every logo method checks explicitly and returns ErrStorageNotConfigured
// rather than a nil-pointer panic.
type TenantService struct {
	db      *gorm.DB
	tenants repository.TenantRepository
	storage ObjectStorage
}

func NewTenantService(db *gorm.DB, tenants repository.TenantRepository, storage ObjectStorage) *TenantService {
	return &TenantService{db: db, tenants: tenants, storage: storage}
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

// Get returns the tenant record for a session's own tenant_id (from JWT
// claims) — used by the "which clinic am I logged into" endpoint, never for
// looking up an arbitrary tenant.
func (s *TenantService) Get(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return s.tenants.FindByID(ctx, id)
}

type UpdateTenantProfileInput struct {
	Name          string
	Document      string
	Email         string
	Phone         string
	AddressStreet *string
	AddressCity   *string
	AddressState  *string
	AddressZip    *string
}

// UpdateProfile changes clinic profile/letterhead fields only — Slug, Type,
// IsActive, Settings and LogoKey are loaded from the current row and left
// untouched, the same load-then-mutate shape as UserService.Update
// protecting PasswordHash. This is deliberately not a PUT-style full
// replace: Slug is the login identifier (too risky to let a profile-edit
// screen touch) and LogoKey has its own dedicated upload flow.
func (s *TenantService) UpdateProfile(ctx context.Context, tenantID uuid.UUID, in UpdateTenantProfileInput) (*models.Tenant, error) {
	if in.Name == "" || in.Document == "" || in.Email == "" {
		return nil, fmt.Errorf("%w: name, document and email are required", ErrValidation)
	}
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	tenant.Name = in.Name
	tenant.Document = in.Document
	tenant.Email = in.Email
	tenant.Phone = in.Phone
	tenant.AddressStreet = in.AddressStreet
	tenant.AddressCity = in.AddressCity
	tenant.AddressState = in.AddressState
	tenant.AddressZip = in.AddressZip
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

// CreateLogoUploadURL mints a presigned PUT URL for the clinic logo — same
// two-step pattern as PatientDocumentService.CreateUploadURL, just keyed
// under "tenants/{id}/logo/" instead of a patient's folder.
func (s *TenantService) CreateLogoUploadURL(ctx context.Context, tenantID uuid.UUID, fileName, contentType string) (uploadURL, fileKey string, err error) {
	if s.storage == nil {
		return "", "", ErrStorageNotConfigured
	}
	if fileName == "" || contentType == "" {
		return "", "", fmt.Errorf("%w: file_name and content_type are required", ErrValidation)
	}
	fileKey = fmt.Sprintf("tenants/%s/logo/%s-%s", tenantID, uuid.New(), fileName)
	uploadURL, err = s.storage.PresignUpload(ctx, fileKey, contentType, logoPresignExpiry)
	if err != nil {
		return "", "", err
	}
	return uploadURL, fileKey, nil
}

// SetLogo persists the file_key once the browser's PUT to R2 has finished,
// then best-effort deletes whatever logo it's replacing — a failure to
// clean up the old object is logged, not returned, since the new logo is
// already saved from the caller's perspective (same tolerance as
// PatientDocumentService.Delete's R2 cleanup).
func (s *TenantService) SetLogo(ctx context.Context, tenantID uuid.UUID, fileKey string) (*models.Tenant, error) {
	if fileKey == "" {
		return nil, fmt.Errorf("%w: file_key is required", ErrValidation)
	}
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	oldKey := tenant.LogoKey
	tenant.LogoKey = &fileKey
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	if oldKey != nil && s.storage != nil {
		if err := s.storage.DeleteObject(ctx, *oldKey); err != nil {
			log.Printf("tenant: failed to delete old logo object %q: %v", *oldKey, err)
		}
	}
	return tenant, nil
}

// DeleteLogo clears the field first (that's what makes the logo disappear
// from the app) then best-effort removes the R2 object, mirroring
// PatientDocumentService.Delete's ordering and error tolerance.
func (s *TenantService) DeleteLogo(ctx context.Context, tenantID uuid.UUID) error {
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if tenant.LogoKey == nil {
		return nil
	}
	key := *tenant.LogoKey
	tenant.LogoKey = nil
	if err := s.tenants.Update(ctx, tenant); err != nil {
		return err
	}
	if s.storage != nil {
		if err := s.storage.DeleteObject(ctx, key); err != nil {
			log.Printf("tenant: failed to delete logo object %q: %v", key, err)
		}
	}
	return nil
}

// LogoDownloadURL returns "" (not an error) when no logo is set — the
// Configurações screen just shows a placeholder in that case, no need for a
// dedicated not-found error the frontend would have to special-case.
func (s *TenantService) LogoDownloadURL(ctx context.Context, tenantID uuid.UUID) (string, error) {
	if s.storage == nil {
		return "", ErrStorageNotConfigured
	}
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if tenant.LogoKey == nil {
		return "", nil
	}
	return s.storage.PresignDownload(ctx, *tenant.LogoKey, logoPresignExpiry)
}

// GetLogoBytes downloads the logo directly (not a presigned URL) for
// server-side embedding into a letterhead PDF. Returns (nil, nil) — not an
// error — whenever storage isn't configured or no logo is set, so
// DocumentTemplateHandler.Generate can treat "no logo available" as a
// no-op rather than special-casing two different reasons.
func (s *TenantService) GetLogoBytes(ctx context.Context, tenantID uuid.UUID) ([]byte, error) {
	if s.storage == nil {
		return nil, nil
	}
	tenant, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant.LogoKey == nil {
		return nil, nil
	}
	return s.storage.DownloadObject(ctx, *tenant.LogoKey)
}
