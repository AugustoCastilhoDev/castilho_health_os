package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

// ObjectStorage is the slice of *storage.R2Client this service actually
// needs — defined here (not imported from internal/storage) so tests can
// fake it without touching a real bucket, and so this package doesn't
// depend on the AWS SDK at all.
type ObjectStorage interface {
	PresignUpload(ctx context.Context, fileKey, contentType string, expiresIn time.Duration) (string, error)
	PresignDownload(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error)
	DeleteObject(ctx context.Context, fileKey string) error
}

const presignExpiry = 15 * time.Minute

// PatientDocumentService is usable with storage == nil (e.g. no R2 account
// configured yet) — every method that needs the bucket checks explicitly
// and returns ErrStorageNotConfigured rather than a nil-pointer panic, so
// the rest of the app keeps working without a Cloudflare account.
type PatientDocumentService struct {
	documents repository.PatientDocumentRepository
	storage   ObjectStorage
}

func NewPatientDocumentService(documents repository.PatientDocumentRepository, storage ObjectStorage) *PatientDocumentService {
	return &PatientDocumentService{documents: documents, storage: storage}
}

// CreateUploadURL mints a presigned PUT URL and the file_key the browser
// will upload to directly — the file's bytes never pass through this
// server. The tenant/patient-scoped prefix is a defense-in-depth isolation
// boundary inside the shared bucket, on top of (never instead of) the
// tenant_id column checked on every query.
func (s *PatientDocumentService) CreateUploadURL(ctx context.Context, tenantID, patientID uuid.UUID, fileName, contentType string) (uploadURL, fileKey string, err error) {
	if s.storage == nil {
		return "", "", ErrStorageNotConfigured
	}
	if fileName == "" {
		return "", "", fmt.Errorf("%w: file_name is required", ErrValidation)
	}
	if contentType == "" {
		return "", "", fmt.Errorf("%w: mime_type is required", ErrValidation)
	}
	fileKey = fmt.Sprintf("tenants/%s/patients/%s/%s-%s", tenantID, patientID, uuid.New(), fileName)
	uploadURL, err = s.storage.PresignUpload(ctx, fileKey, contentType, presignExpiry)
	if err != nil {
		return "", "", err
	}
	return uploadURL, fileKey, nil
}

// Create persists the metadata row once the browser has finished the PUT
// to the presigned URL — this is the only step that touches Postgres.
func (s *PatientDocumentService) Create(ctx context.Context, tenantID uuid.UUID, doc *models.PatientDocument) error {
	if doc.PatientID == uuid.Nil {
		return fmt.Errorf("%w: patient_id is required", ErrValidation)
	}
	if doc.FileKey == "" || doc.FileName == "" {
		return fmt.Errorf("%w: file_key and file_name are required", ErrValidation)
	}
	if doc.FileSize <= 0 {
		return fmt.Errorf("%w: file_size must be positive", ErrValidation)
	}
	return s.documents.Create(ctx, tenantID, doc)
}

func (s *PatientDocumentService) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.PatientDocument, error) {
	return s.documents.ListByPatient(ctx, tenantID, patientID)
}

// DownloadURL resolves a document's own file_key (never trusted from the
// caller) into a short-lived, read-only presigned URL.
func (s *PatientDocumentService) DownloadURL(ctx context.Context, tenantID, id uuid.UUID) (string, error) {
	if s.storage == nil {
		return "", ErrStorageNotConfigured
	}
	doc, err := s.documents.FindByID(ctx, tenantID, id)
	if err != nil {
		return "", err
	}
	return s.storage.PresignDownload(ctx, doc.FileKey, presignExpiry)
}

// Delete removes the metadata row first — that's what makes the document
// invisible to the app — then best-effort cleans up the R2 object. A
// failure on the R2 side is logged, not returned: the row is already gone,
// so the delete the caller asked for did succeed from their perspective.
func (s *PatientDocumentService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	doc, err := s.documents.FindByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if err := s.documents.Delete(ctx, tenantID, id); err != nil {
		return err
	}
	if s.storage != nil {
		if err := s.storage.DeleteObject(ctx, doc.FileKey); err != nil {
			log.Printf("patient_document: failed to delete object %q from storage: %v", doc.FileKey, err)
		}
	}
	return nil
}
