package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestOdontogramaService_RecordEntry_RejectsMissingPatientID(t *testing.T) {
	s := service.NewOdontogramaService(&fakeOdontogramaRepo{})
	err := s.RecordEntry(context.Background(), uuid.New(), &models.OdontogramaEntry{
		ToothNumber: "11",
		Condition:   models.ToothConditionHealthy,
	})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestOdontogramaService_RecordEntry_RejectsInvalidToothNumber(t *testing.T) {
	s := service.NewOdontogramaService(&fakeOdontogramaRepo{})
	err := s.RecordEntry(context.Background(), uuid.New(), &models.OdontogramaEntry{
		PatientID:   uuid.New(),
		ToothNumber: "99",
		Condition:   models.ToothConditionHealthy,
	})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestOdontogramaService_RecordEntry_RejectsUnknownCondition(t *testing.T) {
	s := service.NewOdontogramaService(&fakeOdontogramaRepo{})
	err := s.RecordEntry(context.Background(), uuid.New(), &models.OdontogramaEntry{
		PatientID:   uuid.New(),
		ToothNumber: "11",
		Condition:   models.ToothCondition("INVENTADO"),
	})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestOdontogramaService_RecordEntry_DelegatesValidEntry(t *testing.T) {
	tenantID := uuid.New()
	var createdEntry *models.OdontogramaEntry
	repo := &fakeOdontogramaRepo{
		createFn: func(ctx context.Context, gotTenantID uuid.UUID, entry *models.OdontogramaEntry) error {
			assert.Equal(t, tenantID, gotTenantID)
			createdEntry = entry
			return nil
		},
	}
	s := service.NewOdontogramaService(repo)
	entry := &models.OdontogramaEntry{
		PatientID:   uuid.New(),
		ToothNumber: "85",
		Condition:   models.ToothConditionCavity,
	}
	require.NoError(t, s.RecordEntry(context.Background(), tenantID, entry))
	require.NotNil(t, createdEntry)
	assert.Equal(t, entry.ToothNumber, createdEntry.ToothNumber)
}

func TestOdontogramaService_ListByPatient_Delegates(t *testing.T) {
	tenantID, patientID := uuid.New(), uuid.New()
	want := []models.OdontogramaEntry{{ToothNumber: "18", Condition: models.ToothConditionMissing}}
	repo := &fakeOdontogramaRepo{
		listByPatientFn: func(ctx context.Context, gotTenantID, gotPatientID uuid.UUID) ([]models.OdontogramaEntry, error) {
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, patientID, gotPatientID)
			return want, nil
		},
	}
	s := service.NewOdontogramaService(repo)
	got, err := s.ListByPatient(context.Background(), tenantID, patientID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestOdontogramaService_CurrentChart_Delegates(t *testing.T) {
	tenantID, patientID := uuid.New(), uuid.New()
	wantErr := errors.New("boom")
	repo := &fakeOdontogramaRepo{
		currentChartFn: func(ctx context.Context, gotTenantID, gotPatientID uuid.UUID) ([]models.OdontogramaEntry, error) {
			return nil, wantErr
		},
	}
	s := service.NewOdontogramaService(repo)
	_, err := s.CurrentChart(context.Background(), tenantID, patientID)
	require.ErrorIs(t, err, wantErr)
}
