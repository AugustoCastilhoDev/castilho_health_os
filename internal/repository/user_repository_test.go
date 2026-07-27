package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestUserRepository_CreateAndFindByEmail(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewUserRepository(gdb)
	ctx := context.Background()

	user := &models.User{
		Name:         "Dra. Ana",
		Email:        "ana@example.com",
		PasswordHash: "hash",
		Role:         models.RoleDoctor,
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, user))

	found, err := repo.FindByEmail(ctx, tenant.ID, "ana@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
}

// Update does a full-row replace (Select("*")): any field the caller
// didn't set on the struct it passes gets written as its zero value. This
// locks that contract in explicitly, using CouncilNumber as the omitted
// field — the same class of bug the PaymentMethod CHECK-constraint
// failure was: a field silently zeroed instead of left alone.
func TestUserRepository_Update_IsFullRowReplace(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewUserRepository(gdb)
	ctx := context.Background()

	councilNumber := "12345"
	user := &models.User{
		Name:          "Dr. Joao",
		Email:         "joao@example.com",
		PasswordHash:  "hash",
		Role:          models.RoleDoctor,
		IsActive:      true,
		CouncilNumber: &councilNumber,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, user))

	// Update with CouncilNumber left nil — full replace means it clears.
	user.Name = "Dr. Joao Silva"
	user.CouncilNumber = nil
	require.NoError(t, repo.Update(ctx, tenant.ID, user))

	reloaded, err := repo.FindByID(ctx, tenant.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Dr. Joao Silva", reloaded.Name)
	assert.Nil(t, reloaded.CouncilNumber)
}

func TestUserRepository_List_FiltersByRoleAndIncludesInactive(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewUserRepository(gdb)
	ctx := context.Background()

	doctor := testutil.NewUser(t, gdb, tenant.ID, models.RoleDoctor)
	testutil.NewUser(t, gdb, tenant.ID, models.RoleReceptionist)

	inactiveDoctor := &models.User{
		Name:         "Inactive Doctor",
		Email:        "inactive@example.com",
		PasswordHash: "hash",
		Role:         models.RoleDoctor,
		IsActive:     false,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, inactiveDoctor))

	doctors, err := repo.List(ctx, tenant.ID, ptr(models.RoleDoctor))
	require.NoError(t, err)

	ids := make([]string, len(doctors))
	for i, u := range doctors {
		ids[i] = u.ID.String()
	}
	assert.Contains(t, ids, doctor.ID.String())
	assert.Contains(t, ids, inactiveDoctor.ID.String(), "List must include inactive users so admins can find and reactivate them")
	assert.Len(t, doctors, 2)
}

func ptr[T any](v T) *T { return &v }
