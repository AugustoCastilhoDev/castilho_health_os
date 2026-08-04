package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type MedicalRecordService struct {
	records repository.MedicalRecordRepository
}

func NewMedicalRecordService(records repository.MedicalRecordRepository) *MedicalRecordService {
	return &MedicalRecordService{records: records}
}

func validateMedicalRecord(record *models.MedicalRecord) error {
	if record.PatientID == uuid.Nil {
		return fmt.Errorf("%w: patient_id is required", ErrValidation)
	}
	if record.ProfessionalID == uuid.Nil {
		return fmt.Errorf("%w: professional_id is required", ErrValidation)
	}
	if record.Content == "" {
		return fmt.Errorf("%w: content is required", ErrValidation)
	}
	switch record.Type {
	case models.RecordTypeMedical, models.RecordTypeDental, models.RecordTypePsychology, models.RecordTypePsychiatry:
	default:
		return fmt.Errorf("%w: unknown record type %q", ErrValidation, record.Type)
	}
	return nil
}

func (s *MedicalRecordService) Create(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error {
	if err := validateMedicalRecord(record); err != nil {
		return err
	}
	return s.records.Create(ctx, tenantID, record)
}

func (s *MedicalRecordService) Get(ctx context.Context, tenantID, id uuid.UUID) (*models.MedicalRecord, error) {
	return s.records.FindByID(ctx, tenantID, id)
}

func (s *MedicalRecordService) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MedicalRecord, error) {
	return s.records.ListByPatient(ctx, tenantID, patientID)
}

// Update leaves PatientID/ProfessionalID/AppointmentID alone — only
// Type/Content/CID are editable after creation — and relies entirely on the
// repository's is_locked-guarded WHERE clause to enforce immutability once
// locked, rather than re-checking IsLocked here (a check-then-act gap here
// could race with a concurrent Lock call the same way it would in SQL).
func (s *MedicalRecordService) Update(ctx context.Context, tenantID uuid.UUID, record *models.MedicalRecord) error {
	if record.Content == "" {
		return fmt.Errorf("%w: content is required", ErrValidation)
	}
	switch record.Type {
	case models.RecordTypeMedical, models.RecordTypeDental, models.RecordTypePsychology, models.RecordTypePsychiatry:
	default:
		return fmt.Errorf("%w: unknown record type %q", ErrValidation, record.Type)
	}
	return s.records.Update(ctx, tenantID, record)
}

func (s *MedicalRecordService) Lock(ctx context.Context, tenantID, id, lockedByID uuid.UUID) (*models.MedicalRecord, error) {
	return s.records.Lock(ctx, tenantID, id, lockedByID)
}
