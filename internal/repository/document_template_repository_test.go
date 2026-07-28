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

func TestDocumentTemplateRepository_CreateUpdateList(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenantA := testutil.NewTenant(t, gdb)
	tenantB := testutil.NewTenant(t, gdb)
	repo := repository.NewDocumentTemplateRepository(gdb)
	ctx := context.Background()

	tmplA := &models.DocumentTemplate{
		Name:    "Atestado padrão",
		Type:    models.TemplateTypeCertificate,
		Content: "Atesto que {{patient_name}} esteve em consulta em {{date}}.",
	}
	require.NoError(t, repo.Create(ctx, tenantA.ID, tmplA))
	assert.True(t, tmplA.IsActive)

	tmplB := &models.DocumentTemplate{
		Name:    "Outro tenant",
		Type:    models.TemplateTypeOther,
		Content: "irrelevante",
	}
	require.NoError(t, repo.Create(ctx, tenantB.ID, tmplB))

	// Tenant isolation: listing tenantA never surfaces tenantB's template.
	listA, err := repo.ListByTenant(ctx, tenantA.ID)
	require.NoError(t, err)
	require.Len(t, listA, 1)
	assert.Equal(t, tmplA.ID, listA[0].ID)

	tmplA.Name = "Atestado revisado"
	tmplA.IsActive = false
	require.NoError(t, repo.Update(ctx, tenantA.ID, tmplA))

	found, err := repo.FindByID(ctx, tenantA.ID, tmplA.ID)
	require.NoError(t, err)
	assert.Equal(t, "Atestado revisado", found.Name)
	assert.False(t, found.IsActive)

	// Cross-tenant lookup by ID must behave as not-found, never leak data.
	_, err = repo.FindByID(ctx, tenantB.ID, tmplA.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
