package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestPatientDocumentService_CreateUploadURL_RequiresStorage(t *testing.T) {
	svc := service.NewPatientDocumentService(&fakePatientDocumentRepo{}, nil)
	_, _, err := svc.CreateUploadURL(context.Background(), uuid.New(), uuid.New(), "exame.pdf", "application/pdf")
	require.ErrorIs(t, err, service.ErrStorageNotConfigured)
}

func TestPatientDocumentService_CreateUploadURL_ScopesFileKeyToTenantAndPatient(t *testing.T) {
	tenantID := uuid.New()
	patientID := uuid.New()
	var capturedKey string
	storage := &fakeObjectStorage{
		presignUploadFn: func(ctx context.Context, fileKey, contentType string, expiresIn time.Duration) (string, error) {
			capturedKey = fileKey
			return "https://r2.example/presigned-put", nil
		},
	}
	svc := service.NewPatientDocumentService(&fakePatientDocumentRepo{}, storage)

	url, fileKey, err := svc.CreateUploadURL(context.Background(), tenantID, patientID, "exame.pdf", "application/pdf")
	require.NoError(t, err)
	assert.Equal(t, "https://r2.example/presigned-put", url)
	assert.Equal(t, fileKey, capturedKey)
	assert.Contains(t, fileKey, "tenants/"+tenantID.String()+"/patients/"+patientID.String()+"/")
	assert.Contains(t, fileKey, "exame.pdf")
}

func TestPatientDocumentService_Create_Validation(t *testing.T) {
	cases := []struct {
		name string
		doc  *models.PatientDocument
	}{
		{"missing patient_id", &models.PatientDocument{FileKey: "k", FileName: "f", FileSize: 10}},
		{"missing file_key", &models.PatientDocument{PatientID: uuid.New(), FileName: "f", FileSize: 10}},
		{"non-positive file_size", &models.PatientDocument{PatientID: uuid.New(), FileKey: "k", FileName: "f", FileSize: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewPatientDocumentService(&fakePatientDocumentRepo{}, nil)
			err := svc.Create(context.Background(), uuid.New(), tc.doc)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestPatientDocumentService_DownloadURL_RequiresStorage(t *testing.T) {
	svc := service.NewPatientDocumentService(&fakePatientDocumentRepo{}, nil)
	_, err := svc.DownloadURL(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, service.ErrStorageNotConfigured)
}

func TestPatientDocumentService_DownloadURL_ResolvesStoredFileKey(t *testing.T) {
	doc := &models.PatientDocument{FileKey: "tenants/x/patients/y/exame.pdf"}
	repo := &fakePatientDocumentRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error) {
			return doc, nil
		},
	}
	var capturedKey string
	storage := &fakeObjectStorage{
		presignDownloadFn: func(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error) {
			capturedKey = fileKey
			return "https://r2.example/presigned-get", nil
		},
	}
	svc := service.NewPatientDocumentService(repo, storage)

	url, err := svc.DownloadURL(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "https://r2.example/presigned-get", url)
	assert.Equal(t, doc.FileKey, capturedKey)
}

// Regression guard: deleting the metadata row must succeed even if the R2
// cleanup call fails — the row being gone is what the caller actually asked
// for, and an orphaned object is a logged warning, not a user-facing error.
func TestPatientDocumentService_Delete_SucceedsEvenIfStorageCleanupFails(t *testing.T) {
	doc := &models.PatientDocument{FileKey: "tenants/x/patients/y/exame.pdf"}
	deleted := false
	repo := &fakePatientDocumentRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.PatientDocument, error) {
			return doc, nil
		},
		deleteFn: func(ctx context.Context, tenantID, id uuid.UUID) error {
			deleted = true
			return nil
		},
	}
	storage := &fakeObjectStorage{
		deleteObjectFn: func(ctx context.Context, fileKey string) error {
			return assert.AnError
		},
	}
	svc := service.NewPatientDocumentService(repo, storage)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.True(t, deleted)
}
