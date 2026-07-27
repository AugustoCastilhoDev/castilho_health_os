package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestTenantRepository_CreateAndFindBySlug(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)

	repo := repository.NewTenantRepository(gdb)
	found, err := repo.FindBySlug(context.Background(), tenant.Slug)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, found.ID)
	assert.Equal(t, tenant.Name, found.Name)
}

func TestTenantRepository_FindBySlug_NotFound(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	repo := repository.NewTenantRepository(gdb)

	_, err := repo.FindBySlug(context.Background(), "does-not-exist-"+uuid.NewString())
	require.ErrorIs(t, err, repository.ErrNotFound)
}
