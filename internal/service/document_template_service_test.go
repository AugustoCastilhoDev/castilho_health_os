package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestDocumentTemplateService_Create_Validation(t *testing.T) {
	cases := []struct {
		name string
		tmpl *models.DocumentTemplate
	}{
		{"missing name", &models.DocumentTemplate{Type: models.TemplateTypeCertificate, Content: "x"}},
		{"missing content", &models.DocumentTemplate{Name: "Atestado", Type: models.TemplateTypeCertificate}},
		{"unknown type", &models.DocumentTemplate{Name: "Atestado", Type: "BOGUS", Content: "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewDocumentTemplateService(&fakeDocumentTemplateRepo{})
			err := svc.Create(context.Background(), uuid.New(), tc.tmpl)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestDocumentTemplateService_Generate_RejectsInactiveTemplate(t *testing.T) {
	repo := &fakeDocumentTemplateRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
			return &models.DocumentTemplate{
				Name:     "Atestado",
				Type:     models.TemplateTypeCertificate,
				Content:  "Atesto que {{patient_name}} esteve presente.",
				IsActive: false,
			}, nil
		},
	}
	svc := service.NewDocumentTemplateService(repo)

	_, _, err := svc.Generate(context.Background(), uuid.New(), uuid.New(), map[string]string{"patient_name": "Mariana"})
	require.ErrorIs(t, err, service.ErrValidation)
}

func TestDocumentTemplateService_Generate_ProducesNonEmptyPDF(t *testing.T) {
	repo := &fakeDocumentTemplateRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
			return &models.DocumentTemplate{
				Name:     "Atestado",
				Type:     models.TemplateTypeCertificate,
				Content:  "Atesto que {{patient_name}} esteve presente em {{date}}.",
				IsActive: true,
			}, nil
		},
	}
	svc := service.NewDocumentTemplateService(repo)

	pdfBytes, tmpl, err := svc.Generate(context.Background(), uuid.New(), uuid.New(), map[string]string{
		"patient_name": "Mariana Costa",
		"date":         "28/07/2026",
	})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	require.Equal(t, "Atestado", tmpl.Name)
	// %PDF is the magic header every valid PDF file starts with.
	require.Equal(t, "%PDF", string(pdfBytes[:4]))
}
