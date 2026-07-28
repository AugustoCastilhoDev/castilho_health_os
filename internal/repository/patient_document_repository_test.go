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

func TestPatientDocumentRepository_CreateListDelete(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	staff := testutil.NewUser(t, gdb, tenant.ID, models.RoleReceptionist)
	patient := testutil.NewPatient(t, gdb, tenant.ID)
	repo := repository.NewPatientDocumentRepository(gdb)
	ctx := context.Background()

	doc := &models.PatientDocument{
		PatientID:    patient.ID,
		UploadedByID: staff.ID,
		FileKey:      "tenants/" + tenant.ID.String() + "/patients/" + patient.ID.String() + "/exame.pdf",
		FileName:     "exame.pdf",
		FileSize:     2048,
		MimeType:     "application/pdf",
	}
	require.NoError(t, repo.Create(ctx, tenant.ID, doc))

	found, err := repo.FindByID(ctx, tenant.ID, doc.ID)
	require.NoError(t, err)
	assert.Equal(t, "exame.pdf", found.FileName)

	list, err := repo.ListByPatient(ctx, tenant.ID, patient.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, doc.ID, list[0].ID)

	require.NoError(t, repo.Delete(ctx, tenant.ID, doc.ID))
	_, err = repo.FindByID(ctx, tenant.ID, doc.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Delete(ctx, tenant.ID, doc.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
