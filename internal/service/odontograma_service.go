package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type OdontogramaService struct {
	entries repository.OdontogramaRepository
}

func NewOdontogramaService(entries repository.OdontogramaRepository) *OdontogramaService {
	return &OdontogramaService{entries: entries}
}

func (s *OdontogramaService) RecordEntry(ctx context.Context, tenantID uuid.UUID, entry *models.OdontogramaEntry) error {
	if entry.PatientID == uuid.Nil {
		return fmt.Errorf("%w: patient_id is required", ErrValidation)
	}
	if !models.IsValidToothNumber(entry.ToothNumber) {
		return fmt.Errorf("%w: invalid tooth_number %q", ErrValidation, entry.ToothNumber)
	}
	if !entry.Condition.Valid() {
		return fmt.Errorf("%w: unknown condition %q", ErrValidation, entry.Condition)
	}
	return s.entries.Create(ctx, tenantID, entry)
}

func (s *OdontogramaService) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	return s.entries.ListByPatient(ctx, tenantID, patientID)
}

func (s *OdontogramaService) CurrentChart(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.OdontogramaEntry, error) {
	return s.entries.CurrentChart(ctx, tenantID, patientID)
}
