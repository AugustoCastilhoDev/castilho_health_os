package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

func TestUserService_Create_RejectsShortPassword(t *testing.T) {
	repo := &fakeUserRepo{
		findByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := service.NewUserService(repo)

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateUserInput{
		Name: "X", Email: "x@example.com", Password: "short", Role: models.RoleDoctor,
	})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestUserService_Create_RejectsDuplicateEmail(t *testing.T) {
	repo := &fakeUserRepo{
		findByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
			return &models.User{}, nil // already exists
		},
	}
	svc := service.NewUserService(repo)

	_, err := svc.Create(context.Background(), uuid.New(), service.CreateUserInput{
		Name: "X", Email: "taken@example.com", Password: "longenough1", Role: models.RoleDoctor,
	})
	require.ErrorIs(t, err, service.ErrConflict)
}

// The critical regression case: Update must never wipe PasswordHash, even
// though the update DTO has no password field at all. This is exactly the
// bug class the PaymentMethod CHECK-constraint failure was — a field
// silently zeroed because it wasn't part of the request shape.
func TestUserService_Update_PreservesPasswordHash(t *testing.T) {
	tenantID, userID := uuid.New(), uuid.New()
	existing := &models.User{
		TenantModel:  models.TenantModel{BaseModel: models.BaseModel{ID: userID}, TenantID: tenantID},
		Name:         "Old Name",
		Email:        "user@example.com",
		PasswordHash: "original-bcrypt-hash",
		Role:         models.RoleDoctor,
		IsActive:     true,
	}

	var saved *models.User
	repo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, tid, id uuid.UUID) (*models.User, error) {
			return existing, nil
		},
		updateFn: func(ctx context.Context, tid uuid.UUID, u *models.User) error {
			saved = u
			return nil
		},
	}
	svc := service.NewUserService(repo)

	_, err := svc.Update(context.Background(), tenantID, userID, service.UpdateUserInput{
		Name: "New Name", Email: "user@example.com", Role: models.RoleDoctor, IsActive: true,
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "original-bcrypt-hash", saved.PasswordHash, "profile update must not touch the password hash")
	assert.Equal(t, "New Name", saved.Name)
}

func TestUserService_ChangeOwnPassword_RejectsWrongOldPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-old-password")
	require.NoError(t, err)
	existing := &models.User{PasswordHash: hash}

	repo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, tid, id uuid.UUID) (*models.User, error) { return existing, nil },
	}
	svc := service.NewUserService(repo)

	err = svc.ChangeOwnPassword(context.Background(), uuid.New(), uuid.New(), "wrong-old-password", "new-password-123")
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestUserService_ChangeOwnPassword_Success(t *testing.T) {
	hash, err := auth.HashPassword("correct-old-password")
	require.NoError(t, err)
	existing := &models.User{PasswordHash: hash}

	var saved *models.User
	repo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, tid, id uuid.UUID) (*models.User, error) { return existing, nil },
		updateFn: func(ctx context.Context, tid uuid.UUID, u *models.User) error {
			saved = u
			return nil
		},
	}
	svc := service.NewUserService(repo)

	err = svc.ChangeOwnPassword(context.Background(), uuid.New(), uuid.New(), "correct-old-password", "new-password-123")
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.True(t, auth.CheckPassword(saved.PasswordHash, "new-password-123"))
}

// ResetPassword is the admin path: it must succeed without the caller
// knowing (or the service checking) any old password.
func TestUserService_ResetPassword_DoesNotRequireOldPassword(t *testing.T) {
	existing := &models.User{PasswordHash: "irrelevant-old-hash"}

	var saved *models.User
	repo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, tid, id uuid.UUID) (*models.User, error) { return existing, nil },
		updateFn: func(ctx context.Context, tid uuid.UUID, u *models.User) error {
			saved = u
			return nil
		},
	}
	svc := service.NewUserService(repo)

	err := svc.ResetPassword(context.Background(), uuid.New(), uuid.New(), "brand-new-password")
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.True(t, auth.CheckPassword(saved.PasswordHash, "brand-new-password"))
}
