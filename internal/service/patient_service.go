package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type PatientService struct {
	patients repository.PatientRepository
}

func NewPatientService(patients repository.PatientRepository) *PatientService {
	return &PatientService{patients: patients}
}

func (s *PatientService) Create(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error {
	if patient.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	return s.patients.Create(ctx, tenantID, patient)
}

func (s *PatientService) Get(ctx context.Context, tenantID, id uuid.UUID) (*models.Patient, error) {
	return s.patients.FindByID(ctx, tenantID, id)
}

func (s *PatientService) Update(ctx context.Context, tenantID uuid.UUID, patient *models.Patient) error {
	if patient.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	return s.patients.Update(ctx, tenantID, patient)
}

func (s *PatientService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.patients.Delete(ctx, tenantID, id)
}

func (s *PatientService) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.patients.Search(ctx, tenantID, query, limit, offset)
}
