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

// Regression guard for the letterhead feature: address/logo fields must
// round-trip through Update, and a field left nil (e.g. no logo set yet)
// must not error the whole update.
func TestTenantRepository_Update_PersistsAddressAndLogo(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewTenantRepository(gdb)
	ctx := context.Background()

	street, city, state, zip := "Rua Exemplo, 100", "São Paulo", "SP", "01000-000"
	tenant.AddressStreet = &street
	tenant.AddressCity = &city
	tenant.AddressState = &state
	tenant.AddressZip = &zip
	require.NoError(t, repo.Update(ctx, tenant))

	found, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.NotNil(t, found.AddressStreet)
	assert.Equal(t, street, *found.AddressStreet)
	require.NotNil(t, found.AddressCity)
	assert.Equal(t, city, *found.AddressCity)
	assert.Nil(t, found.LogoKey, "logo_key must stay nil until a logo is actually set")

	logoKey := "tenants/x/logo/test.png"
	found.LogoKey = &logoKey
	require.NoError(t, repo.Update(ctx, found))

	final, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.NotNil(t, final.LogoKey)
	assert.Equal(t, logoKey, *final.LogoKey)
}
