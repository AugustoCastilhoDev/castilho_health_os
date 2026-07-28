package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/service"
)

func councilStr(s string) *string { return &s }

func healthProfessional() *models.User {
	cpf := "12345678901"
	birth := time.Date(1985, 3, 10, 0, 0, 0, 0, time.UTC)
	return &models.User{
		TenantModel:   models.TenantModel{BaseModel: models.BaseModel{ID: uuid.New()}},
		Name:          "Joana Souza",
		Email:         "joana@example.com",
		Role:          models.RoleDoctor,
		CouncilType:   councilStr("CRM"),
		CouncilNumber: councilStr("12345"),
		CouncilState:  councilStr("SP"),
		CPF:           &cpf,
		BirthDate:     &birth,
	}
}

func TestMemedService_GetPrescriberToken_RequiresClient(t *testing.T) {
	svc := service.NewMemedService(&fakeMemedPrescriptionLogRepo{}, &fakeUserRepo{}, nil)
	_, err := svc.GetPrescriberToken(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, service.ErrMemedNotConfigured)
}

func TestMemedService_GetPrescriberToken_RejectsNonHealthProfessional(t *testing.T) {
	staff := &models.User{TenantModel: models.TenantModel{BaseModel: models.BaseModel{ID: uuid.New()}}, Role: models.RoleReceptionist}
	users := &fakeUserRepo{findByIDFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
		return staff, nil
	}}
	svc := service.NewMemedService(&fakeMemedPrescriptionLogRepo{}, users, &fakeMemedClient{})
	_, err := svc.GetPrescriberToken(context.Background(), uuid.New(), staff.ID)
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestMemedService_GetPrescriberToken_RequiresCompleteProfile(t *testing.T) {
	incomplete := &models.User{TenantModel: models.TenantModel{BaseModel: models.BaseModel{ID: uuid.New()}}, Role: models.RoleDoctor}
	users := &fakeUserRepo{findByIDFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
		return incomplete, nil
	}}
	svc := service.NewMemedService(&fakeMemedPrescriptionLogRepo{}, users, &fakeMemedClient{})
	_, err := svc.GetPrescriberToken(context.Background(), uuid.New(), incomplete.ID)
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestMemedService_GetPrescriberToken_BuildsPrescriberFromUser(t *testing.T) {
	doctor := healthProfessional()
	users := &fakeUserRepo{findByIDFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
		return doctor, nil
	}}
	var captured service.MemedPrescriber
	client := &fakeMemedClient{fetchOrRegisterTokenFn: func(ctx context.Context, p service.MemedPrescriber) (string, error) {
		captured = p
		return "memed-token-123", nil
	}}
	svc := service.NewMemedService(&fakeMemedPrescriptionLogRepo{}, users, client)

	token, err := svc.GetPrescriberToken(context.Background(), uuid.New(), doctor.ID)
	require.NoError(t, err)
	assert.Equal(t, "memed-token-123", token)
	assert.Equal(t, doctor.ID.String(), captured.ExternalID)
	assert.Equal(t, "Joana", captured.Name)
	assert.Equal(t, "Souza", captured.Surname)
	assert.Equal(t, "12345678901", captured.CPF)
	assert.Equal(t, "CRM", captured.BoardCode)
	assert.Equal(t, "12345", captured.BoardNumber)
	assert.Equal(t, "SP", captured.BoardState)
}

func TestMemedService_RecordIssuance_Validation(t *testing.T) {
	svc := service.NewMemedService(&fakeMemedPrescriptionLogRepo{}, &fakeUserRepo{}, nil)

	_, err := svc.RecordIssuance(context.Background(), uuid.New(), uuid.New(), uuid.New(), "")
	require.ErrorIs(t, err, service.ErrValidation)

	_, err = svc.RecordIssuance(context.Background(), uuid.New(), uuid.Nil, uuid.New(), "presc-1")
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestMemedService_RecordIssuance_RejectsDuplicateMemedID(t *testing.T) {
	existing := &models.MemedPrescriptionLog{MemedPrescriptionID: "presc-1"}
	repo := &fakeMemedPrescriptionLogRepo{
		findByMemedIDFn: func(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error) {
			return existing, nil
		},
	}
	svc := service.NewMemedService(repo, &fakeUserRepo{}, nil)
	_, err := svc.RecordIssuance(context.Background(), uuid.New(), uuid.New(), uuid.New(), "presc-1")
	require.ErrorIs(t, err, service.ErrConflict)
}

func TestMemedService_RecordIssuance_PersistsLog(t *testing.T) {
	patientID, professionalID := uuid.New(), uuid.New()
	var created *models.MemedPrescriptionLog
	repo := &fakeMemedPrescriptionLogRepo{
		findByMemedIDFn: func(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error) {
			return nil, repository.ErrNotFound
		},
		createFn: func(ctx context.Context, tenantID uuid.UUID, l *models.MemedPrescriptionLog) error {
			created = l
			return nil
		},
	}
	svc := service.NewMemedService(repo, &fakeUserRepo{}, nil)

	log, err := svc.RecordIssuance(context.Background(), uuid.New(), patientID, professionalID, "presc-2")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, patientID, log.PatientID)
	assert.Equal(t, professionalID, log.ProfessionalID)
	assert.Equal(t, "presc-2", log.MemedPrescriptionID)
	assert.Equal(t, models.MemedPrescriptionIssued, log.Status)
}
