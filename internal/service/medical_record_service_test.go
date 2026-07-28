package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

func TestMedicalRecordService_Create_Validation(t *testing.T) {
	cases := []struct {
		name   string
		record *models.MedicalRecord
	}{
		{"missing patient_id", &models.MedicalRecord{ProfessionalID: uuid.New(), Type: models.RecordTypeMedical, Content: "x"}},
		{"missing professional_id", &models.MedicalRecord{PatientID: uuid.New(), Type: models.RecordTypeMedical, Content: "x"}},
		{"missing content", &models.MedicalRecord{PatientID: uuid.New(), ProfessionalID: uuid.New(), Type: models.RecordTypeMedical}},
		{"unknown type", &models.MedicalRecord{PatientID: uuid.New(), ProfessionalID: uuid.New(), Type: "BOGUS", Content: "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewMedicalRecordService(&fakeMedicalRecordRepo{})
			err := svc.Create(context.Background(), uuid.New(), tc.record)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestMedicalRecordService_Update_PropagatesLockedError(t *testing.T) {
	repo := &fakeMedicalRecordRepo{
		updateFn: func(ctx context.Context, tenantID uuid.UUID, r *models.MedicalRecord) error {
			return repository.ErrLocked
		},
	}
	svc := service.NewMedicalRecordService(repo)

	err := svc.Update(context.Background(), uuid.New(), &models.MedicalRecord{
		Type:    models.RecordTypeMedical,
		Content: "tentando editar depois de travado",
	})
	require.ErrorIs(t, err, repository.ErrLocked)
}

func TestMedicalRecordService_Update_RejectsEmptyContent(t *testing.T) {
	svc := service.NewMedicalRecordService(&fakeMedicalRecordRepo{})
	err := svc.Update(context.Background(), uuid.New(), &models.MedicalRecord{Type: models.RecordTypeMedical})
	require.ErrorIs(t, err, service.ErrValidation)
}
