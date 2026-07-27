package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestPatientService_Create_RejectsEmptyName(t *testing.T) {
	repo := &fakePatientRepo{} // no createFn stubbed — must not be called
	svc := service.NewPatientService(repo)

	err := svc.Create(context.Background(), uuid.New(), &models.Patient{Name: ""})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestPatientService_Search_ClampsLimit(t *testing.T) {
	var gotLimit, gotOffset int
	repo := &fakePatientRepo{
		searchFn: func(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]models.Patient, error) {
			gotLimit, gotOffset = limit, offset
			return nil, nil
		},
	}
	svc := service.NewPatientService(repo)

	_, err := svc.Search(context.Background(), uuid.New(), "x", 0, -5)
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit, "non-positive limit should fall back to the default")
	assert.Equal(t, 0, gotOffset, "negative offset should clamp to zero")

	_, err = svc.Search(context.Background(), uuid.New(), "x", 1000, 10)
	require.NoError(t, err)
	assert.Equal(t, 20, gotLimit, "limit above the cap should fall back to the default")
}
