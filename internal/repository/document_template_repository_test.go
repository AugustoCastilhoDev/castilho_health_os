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

// Regression guard for a real bug found live via Playwright: a template
// created with every layout checkbox explicitly unchecked (Go `false`)
// still rendered the full letterhead, because GORM's `default:true` tag
// made it treat "false" (the zero value) as "not set" and silently insert
// the column's DB default (true) instead. IncludeHeader et al. deliberately
// carry no `default` gorm tag now — Create must persist the exact Go value
// it's given, explicit false included.
func TestDocumentTemplateRepository_Create_PersistsExplicitFalseFlags(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewDocumentTemplateRepository(gdb)
	ctx := context.Background()

	tmpl := &models.DocumentTemplate{
		Name:    "Todas as flags desligadas",
		Type:    models.TemplateTypeOther,
		Content: "irrelevante",
		// IncludeHeader/Footer/Signature/Stamp left at their Go zero value
		// (false) — exactly what an "uncheck every box" form submission
		// produces.
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, tmpl))

	found, err := repo.FindByID(ctx, tenant.ID, tmpl.ID)
	require.NoError(t, err)
	assert.False(t, found.IncludeHeader)
	assert.False(t, found.IncludeFooter)
	assert.False(t, found.IncludeSignature)
	assert.False(t, found.IncludeStamp)
}

// Update must persist an explicit "uncheck this block" choice against a
// template that started with every flag on.
func TestDocumentTemplateRepository_Update_PersistsLayoutFlags(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	repo := repository.NewDocumentTemplateRepository(gdb)
	ctx := context.Background()

	tmpl := &models.DocumentTemplate{
		Name: "Modelo", Type: models.TemplateTypeOther, Content: "x",
		IncludeHeader: true, IncludeFooter: true, IncludeSignature: true, IncludeStamp: true,
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, tmpl))

	tmpl.IncludeHeader = false
	tmpl.IncludeStamp = false
	require.NoError(t, repo.Update(ctx, tenant.ID, tmpl))

	found, err := repo.FindByID(ctx, tenant.ID, tmpl.ID)
	require.NoError(t, err)
	assert.False(t, found.IncludeHeader)
	assert.True(t, found.IncludeFooter, "untouched flags must survive the update")
	assert.True(t, found.IncludeSignature)
	assert.False(t, found.IncludeStamp)
}
