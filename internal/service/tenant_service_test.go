package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

// Register's happy path needs a real *gorm.DB (it opens a transaction to
// create the Tenant and its admin User atomically) and is covered as an
// integration concern by the repository tests instead. Everything tested
// here returns before that transaction is ever opened, so passing a nil
// *gorm.DB is safe — these paths never dereference it.

func validRegisterInput() service.RegisterTenantInput {
	return service.RegisterTenantInput{
		TenantName:     "Clinica Teste",
		TenantSlug:     "clinica-teste",
		TenantType:     models.TenantTypeMista,
		TenantDocument: "12345678000199",
		TenantEmail:    "clinica@example.com",
		AdminName:      "Admin",
		AdminEmail:     "admin@example.com",
		AdminPassword:  "longenough1",
	}
}

func TestTenantService_Register_Validation(t *testing.T) {
	mutate := map[string]func(in *service.RegisterTenantInput){
		"missing tenant name":  func(in *service.RegisterTenantInput) { in.TenantName = "" },
		"missing slug":         func(in *service.RegisterTenantInput) { in.TenantSlug = "" },
		"missing document":     func(in *service.RegisterTenantInput) { in.TenantDocument = "" },
		"missing tenant email": func(in *service.RegisterTenantInput) { in.TenantEmail = "" },
		"missing admin name":   func(in *service.RegisterTenantInput) { in.AdminName = "" },
		"missing admin email":  func(in *service.RegisterTenantInput) { in.AdminEmail = "" },
		"short admin password": func(in *service.RegisterTenantInput) { in.AdminPassword = "short" },
		"invalid tenant type":  func(in *service.RegisterTenantInput) { in.TenantType = "BOGUS" },
	}

	for name, mutateFn := range mutate {
		t.Run(name, func(t *testing.T) {
			in := validRegisterInput()
			mutateFn(&in)

			svc := service.NewTenantService(nil, &fakeTenantRepo{}, nil)
			_, _, err := svc.Register(context.Background(), in)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestTenantService_Register_RejectsTakenSlug(t *testing.T) {
	tenants := &fakeTenantRepo{
		findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
			return &models.Tenant{Slug: slug}, nil // slug already exists
		},
	}
	svc := service.NewTenantService(nil, tenants, nil)

	_, _, err := svc.Register(context.Background(), validRegisterInput())
	require.ErrorIs(t, err, service.ErrConflict)
}

func TestTenantService_Register_PropagatesUnexpectedLookupError(t *testing.T) {
	tenants := &fakeTenantRepo{
		findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
			return nil, context.DeadlineExceeded // some non-ErrNotFound failure
		},
	}
	svc := service.NewTenantService(nil, tenants, nil)

	_, _, err := svc.Register(context.Background(), validRegisterInput())
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func validUpdateProfileInput() service.UpdateTenantProfileInput {
	return service.UpdateTenantProfileInput{
		Name:     "Clínica Atualizada",
		Document: "12345678000199",
		Email:    "clinica@example.com",
		Phone:    "11999999999",
	}
}

func TestTenantService_UpdateProfile_Validation(t *testing.T) {
	mutate := map[string]func(in *service.UpdateTenantProfileInput){
		"missing name":     func(in *service.UpdateTenantProfileInput) { in.Name = "" },
		"missing document": func(in *service.UpdateTenantProfileInput) { in.Document = "" },
		"missing email":    func(in *service.UpdateTenantProfileInput) { in.Email = "" },
	}
	for name, mutateFn := range mutate {
		t.Run(name, func(t *testing.T) {
			in := validUpdateProfileInput()
			mutateFn(&in)

			svc := service.NewTenantService(nil, &fakeTenantRepo{}, nil)
			_, err := svc.UpdateProfile(context.Background(), uuid.New(), in)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

// Regression guard: UpdateProfile must load the current row and mutate only
// the profile fields — Slug/Type/IsActive/LogoKey must survive untouched,
// the same load-then-mutate shape that protects User.PasswordHash.
func TestTenantService_UpdateProfile_PreservesUntouchedFields(t *testing.T) {
	tenantID := uuid.New()
	logoKey := "tenants/x/logo/existing.png"
	existing := &models.Tenant{
		BaseModel: models.BaseModel{ID: tenantID},
		Name:      "Old Name",
		Slug:      "old-slug",
		Type:      models.TenantTypeMedica,
		Document:  "old-doc",
		Email:     "old@example.com",
		IsActive:  true,
		LogoKey:   &logoKey,
	}
	var saved *models.Tenant
	tenants := &fakeTenantRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
			return existing, nil
		},
		updateFn: func(ctx context.Context, tenant *models.Tenant) error {
			saved = tenant
			return nil
		},
	}
	svc := service.NewTenantService(nil, tenants, nil)

	_, err := svc.UpdateProfile(context.Background(), tenantID, validUpdateProfileInput())
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "Clínica Atualizada", saved.Name)
	assert.Equal(t, "old-slug", saved.Slug, "Slug must not change via profile update")
	assert.Equal(t, models.TenantTypeMedica, saved.Type, "Type must not change via profile update")
	assert.True(t, saved.IsActive, "IsActive must not be zeroed")
	require.NotNil(t, saved.LogoKey)
	assert.Equal(t, logoKey, *saved.LogoKey, "LogoKey has its own upload flow, must not be touched here")
}

func TestTenantService_CreateLogoUploadURL_RequiresStorage(t *testing.T) {
	svc := service.NewTenantService(nil, &fakeTenantRepo{}, nil)
	_, _, err := svc.CreateLogoUploadURL(context.Background(), uuid.New(), "logo.png", "image/png")
	require.ErrorIs(t, err, service.ErrStorageNotConfigured)
}

// SetLogo must clean up the previous logo object once the new one is
// safely saved — same best-effort-cleanup ordering as
// PatientDocumentService.Delete.
func TestTenantService_SetLogo_DeletesOldLogoObject(t *testing.T) {
	tenantID := uuid.New()
	oldKey := "tenants/x/logo/old.png"
	tenant := &models.Tenant{BaseModel: models.BaseModel{ID: tenantID}, LogoKey: &oldKey}
	tenants := &fakeTenantRepo{
		findByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Tenant, error) { return tenant, nil },
		updateFn:   func(ctx context.Context, t *models.Tenant) error { return nil },
	}
	var deletedKey string
	storage := &fakeObjectStorage{
		deleteObjectFn: func(ctx context.Context, fileKey string) error {
			deletedKey = fileKey
			return nil
		},
	}
	svc := service.NewTenantService(nil, tenants, storage)

	updated, err := svc.SetLogo(context.Background(), tenantID, "tenants/x/logo/new.png")
	require.NoError(t, err)
	require.NotNil(t, updated.LogoKey)
	assert.Equal(t, "tenants/x/logo/new.png", *updated.LogoKey)
	assert.Equal(t, oldKey, deletedKey, "the previous logo object must be cleaned up")
}
