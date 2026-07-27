package service_test

import (
	"context"
	"testing"

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

			svc := service.NewTenantService(nil, &fakeTenantRepo{})
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
	svc := service.NewTenantService(nil, tenants)

	_, _, err := svc.Register(context.Background(), validRegisterInput())
	require.ErrorIs(t, err, service.ErrConflict)
}

func TestTenantService_Register_PropagatesUnexpectedLookupError(t *testing.T) {
	tenants := &fakeTenantRepo{
		findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
			return nil, context.DeadlineExceeded // some non-ErrNotFound failure
		},
	}
	svc := service.NewTenantService(nil, tenants)

	_, _, err := svc.Register(context.Background(), validRegisterInput())
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
