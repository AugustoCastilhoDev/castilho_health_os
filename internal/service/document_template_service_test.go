package service_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
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

	_, _, err := svc.Generate(context.Background(), uuid.New(), uuid.New(), map[string]string{"patient_name": "Mariana"}, service.DocumentBranding{})
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
	}, service.DocumentBranding{})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	require.Equal(t, "Atestado", tmpl.Name)
	// %PDF is the magic header every valid PDF file starts with.
	require.Equal(t, "%PDF", string(pdfBytes[:4]))
}

// Regression guard for the letterhead feature: a template with every layout
// flag enabled plus a full branding payload (including a logo) must still
// render successfully — this is the path most likely to panic on a bad
// fpdf call (image registration, dash pattern, negative Y footer offset)
// that the flags-all-false case above would never exercise.
func TestDocumentTemplateService_Generate_RendersWithFullLetterhead(t *testing.T) {
	repo := &fakeDocumentTemplateRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
			return &models.DocumentTemplate{
				Name:             "Atestado",
				Type:             models.TemplateTypeCertificate,
				Content:          "Atesto que {{patient_name}} esteve presente em {{date}}.",
				IsActive:         true,
				IncludeHeader:    true,
				IncludeFooter:    true,
				IncludeSignature: true,
				IncludeStamp:     true,
			}, nil
		},
	}
	svc := service.NewDocumentTemplateService(repo)

	pdfBytes, _, err := svc.Generate(context.Background(), uuid.New(), uuid.New(), map[string]string{
		"patient_name": "Mariana Costa",
		"date":         "28/07/2026",
	}, service.DocumentBranding{
		ClinicName:               "Clínica Demo",
		ClinicDocument:           "12.345.678/0001-00",
		ClinicAddress:            "Rua Exemplo, 100 - São Paulo/SP - CEP 01000-000",
		ClinicPhone:              "(11) 4000-0000",
		ClinicEmail:              "contato@clinicademo.com.br",
		LogoBytes:                encodeTinyPNG(t),
		LogoMimeType:             "image/png",
		SignatureCity:            "São Paulo",
		SignatureDate:            "28 de julho de 2026",
		ProfessionalName:         "Dra. Ana Souza",
		ProfessionalRegistration: "CRM/SP 123456",
	})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	require.Equal(t, "%PDF", string(pdfBytes[:4]))
}

// Regression guard: fpdf's built-in PNG decoder can panic (not just return
// an error) on a malformed image — confirmed live when a bad logo crashed
// generation entirely. A tenant's broken logo upload must degrade to "no
// logo in the header," never break every document they try to generate.
func TestDocumentTemplateService_Generate_SurvivesMalformedLogo(t *testing.T) {
	repo := &fakeDocumentTemplateRepo{
		findFn: func(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
			return &models.DocumentTemplate{
				Name:          "Atestado",
				Type:          models.TemplateTypeCertificate,
				Content:       "Atesto que {{patient_name}} esteve presente.",
				IsActive:      true,
				IncludeHeader: true,
			}, nil
		},
	}
	svc := service.NewDocumentTemplateService(repo)

	pdfBytes, _, err := svc.Generate(context.Background(), uuid.New(), uuid.New(), map[string]string{
		"patient_name": "Mariana Costa",
	}, service.DocumentBranding{
		ClinicName:   "Clínica Demo",
		LogoBytes:    []byte("not a real png"),
		LogoMimeType: "image/png",
	})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	require.Equal(t, "%PDF", string(pdfBytes[:4]))
}

func encodeTinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}
