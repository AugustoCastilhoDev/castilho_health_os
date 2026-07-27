package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestPatientRepository_CreateAndSearch(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewPatientRepository(gdb)
	ctx := context.Background()

	patient := &models.Patient{
		Name:     "Maria da Silva",
		Document: "12345678900",
		Phone:    "11988887777",
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, patient))

	byName, err := repo.Search(ctx, tenant.ID, "maria", 10, 0)
	require.NoError(t, err)
	require.Len(t, byName, 1)
	assert.Equal(t, patient.ID, byName[0].ID)

	byDocument, err := repo.Search(ctx, tenant.ID, "12345678900", 10, 0)
	require.NoError(t, err)
	require.Len(t, byDocument, 1)
	assert.Equal(t, patient.ID, byDocument[0].ID)

	noMatch, err := repo.Search(ctx, tenant.ID, "nonexistent-name-xyz", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, noMatch)
}

// Same full-row-replace contract as UserRepository.Update, exercised here
// with BirthDate: a PUT that omits it must null it out, not leave the old
// value in place.
func TestPatientRepository_Update_ClearsOmittedFields(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewPatientRepository(gdb)
	ctx := context.Background()

	birthDate := time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC)
	patient := &models.Patient{Name: "Carlos", BirthDate: &birthDate}
	require.NoError(t, repo.Create(ctx, tenant.ID, patient))

	patient.Name = "Carlos Updated"
	patient.BirthDate = nil
	require.NoError(t, repo.Update(ctx, tenant.ID, patient))

	reloaded, err := repo.FindByID(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	assert.Equal(t, "Carlos Updated", reloaded.Name)
	assert.Nil(t, reloaded.BirthDate)
}
