package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

func newTestIssuer() *auth.JWTIssuer {
	return auth.NewJWTIssuer("test-secret-at-least-32-characters-long", time.Hour)
}

// Every login failure mode must come back as the same generic
// ErrInvalidCredentials — the whole point is that a caller can't tell
// "wrong tenant" from "wrong password" and enumerate valid accounts.
func TestAuthService_Login_FailureModesReturnGenericError(t *testing.T) {
	tenantID := uuid.New()
	activeTenant := &models.Tenant{BaseModel: models.BaseModel{ID: tenantID}, IsActive: true, Slug: "acme"}
	inactiveTenant := &models.Tenant{BaseModel: models.BaseModel{ID: tenantID}, IsActive: false, Slug: "acme"}

	correctHash, err := auth.HashPassword("correct-password")
	require.NoError(t, err)
	activeUser := &models.User{
		TenantModel:  models.TenantModel{BaseModel: models.BaseModel{ID: uuid.New()}, TenantID: tenantID},
		PasswordHash: correctHash,
		Role:         models.RoleDoctor,
		IsActive:     true,
	}
	inactiveUser := &models.User{
		TenantModel:  models.TenantModel{BaseModel: models.BaseModel{ID: uuid.New()}, TenantID: tenantID},
		PasswordHash: correctHash,
		Role:         models.RoleDoctor,
		IsActive:     false,
	}

	cases := []struct {
		name     string
		tenants  *fakeTenantRepo
		users    *fakeUserRepo
		password string
	}{
		{
			name: "unknown tenant",
			tenants: &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
				return nil, repository.ErrNotFound
			}},
			users:    &fakeUserRepo{},
			password: "correct-password",
		},
		{
			name: "inactive tenant",
			tenants: &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
				return inactiveTenant, nil
			}},
			users:    &fakeUserRepo{},
			password: "correct-password",
		},
		{
			name: "unknown user",
			tenants: &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
				return activeTenant, nil
			}},
			users: &fakeUserRepo{findByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
				return nil, repository.ErrNotFound
			}},
			password: "correct-password",
		},
		{
			name: "wrong password",
			tenants: &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
				return activeTenant, nil
			}},
			users: &fakeUserRepo{findByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
				return activeUser, nil
			}},
			password: "wrong-password",
		},
		{
			name: "inactive user",
			tenants: &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
				return activeTenant, nil
			}},
			users: &fakeUserRepo{findByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
				return inactiveUser, nil
			}},
			password: "correct-password",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewAuthService(tc.tenants, tc.users, newTestIssuer())
			_, err := svc.Login(context.Background(), "acme", "user@example.com", tc.password)
			require.ErrorIs(t, err, service.ErrInvalidCredentials)
		})
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	tenant := &models.Tenant{BaseModel: models.BaseModel{ID: tenantID}, IsActive: true, Slug: "acme"}

	hash, err := auth.HashPassword("correct-password")
	require.NoError(t, err)
	user := &models.User{
		TenantModel:  models.TenantModel{BaseModel: models.BaseModel{ID: userID}, TenantID: tenantID},
		PasswordHash: hash,
		Role:         models.RoleDoctor,
		IsActive:     true,
	}

	tenants := &fakeTenantRepo{findBySlugFn: func(ctx context.Context, slug string) (*models.Tenant, error) {
		assert.Equal(t, "acme", slug)
		return tenant, nil
	}}
	users := &fakeUserRepo{findByEmailFn: func(ctx context.Context, tid uuid.UUID, email string) (*models.User, error) {
		assert.Equal(t, tenantID, tid)
		assert.Equal(t, "user@example.com", email)
		return user, nil
	}}

	issuer := newTestIssuer()
	svc := service.NewAuthService(tenants, users, issuer)

	token, err := svc.Login(context.Background(), "acme", "user@example.com", "correct-password")
	require.NoError(t, err)

	claims, err := issuer.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, models.RoleDoctor, claims.Role)
}
