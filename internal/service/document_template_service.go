package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/pdf"
	"github.com/castilho/health-os/internal/repository"
)

type DocumentTemplateService struct {
	templates repository.DocumentTemplateRepository
}

func NewDocumentTemplateService(templates repository.DocumentTemplateRepository) *DocumentTemplateService {
	return &DocumentTemplateService{templates: templates}
}

func validateDocumentTemplate(tmpl *models.DocumentTemplate) error {
	if tmpl.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if tmpl.Content == "" {
		return fmt.Errorf("%w: content is required", ErrValidation)
	}
	switch tmpl.Type {
	case models.TemplateTypeCertificate, models.TemplateTypeDeclaration, models.TemplateTypeReport, models.TemplateTypeOther:
	default:
		return fmt.Errorf("%w: unknown template type %q", ErrValidation, tmpl.Type)
	}
	return nil
}

func (s *DocumentTemplateService) Create(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error {
	if err := validateDocumentTemplate(tmpl); err != nil {
		return err
	}
	tmpl.IsActive = true
	return s.templates.Create(ctx, tenantID, tmpl)
}

func (s *DocumentTemplateService) Get(ctx context.Context, tenantID, id uuid.UUID) (*models.DocumentTemplate, error) {
	return s.templates.FindByID(ctx, tenantID, id)
}

// Update is a full replace, same PUT convention as FinancialRule — the
// frontend always sends the complete template, including is_active.
func (s *DocumentTemplateService) Update(ctx context.Context, tenantID uuid.UUID, tmpl *models.DocumentTemplate) error {
	if err := validateDocumentTemplate(tmpl); err != nil {
		return err
	}
	return s.templates.Update(ctx, tenantID, tmpl)
}

func (s *DocumentTemplateService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.DocumentTemplate, error) {
	return s.templates.ListByTenant(ctx, tenantID)
}

var placeholderPattern = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

// resolvePlaceholders substitutes {{tag}} tokens the caller has a value
// for; a tag with no matching var is left as literal text in the output —
// silently blanking it risks issuing a document that reads as complete but
// is actually missing information (e.g. a CID code).
func resolvePlaceholders(content string, vars map[string]string) string {
	return placeholderPattern.ReplaceAllStringFunc(content, func(match string) string {
		key := placeholderPattern.FindStringSubmatch(match)[1]
		if v, ok := vars[key]; ok {
			return v
		}
		return match
	})
}

// DocumentBranding is the letterhead data Generate can't resolve itself —
// it needs the Tenant/User repositories the handler already has access to
// (patient/professional lookups happen there, see DocumentTemplateHandler.
// Generate). The template's own IncludeHeader/Footer/Signature/Stamp flags
// are not part of this struct — Generate reads those off tmpl directly,
// since the caller only supplies raw facts, never layout decisions.
type DocumentBranding struct {
	ClinicName     string
	ClinicDocument string
	ClinicAddress  string
	ClinicPhone    string
	ClinicEmail    string
	LogoBytes      []byte
	LogoMimeType   string
	SignatureCity  string
	SignatureDate  string

	ProfessionalName         string
	ProfessionalRegistration string
}

// Generate resolves an active template's placeholders against vars and
// renders the result to PDF bytes, wrapped in whichever letterhead blocks
// the template itself has enabled. Resolving vars/branding from a specific
// patient/professional/tenant is the caller's job (it needs repositories
// this service doesn't own) — this only does template lookup, substitution,
// merging the template's own layout flags onto branding, and rendering.
func (s *DocumentTemplateService) Generate(ctx context.Context, tenantID, templateID uuid.UUID, vars map[string]string, branding DocumentBranding) ([]byte, *models.DocumentTemplate, error) {
	tmpl, err := s.templates.FindByID(ctx, tenantID, templateID)
	if err != nil {
		return nil, nil, err
	}
	if !tmpl.IsActive {
		return nil, nil, fmt.Errorf("%w: template is inactive", ErrValidation)
	}
	body := resolvePlaceholders(tmpl.Content, vars)
	opts := pdf.RenderOptions{
		IncludeHeader:    tmpl.IncludeHeader,
		IncludeFooter:    tmpl.IncludeFooter,
		IncludeSignature: tmpl.IncludeSignature,
		IncludeStamp:     tmpl.IncludeStamp,

		ClinicName:     branding.ClinicName,
		ClinicDocument: branding.ClinicDocument,
		ClinicAddress:  branding.ClinicAddress,
		ClinicPhone:    branding.ClinicPhone,
		ClinicEmail:    branding.ClinicEmail,
		LogoBytes:      branding.LogoBytes,
		LogoMimeType:   branding.LogoMimeType,

		SignatureCity: branding.SignatureCity,
		SignatureDate: branding.SignatureDate,

		ProfessionalName:         branding.ProfessionalName,
		ProfessionalRegistration: branding.ProfessionalRegistration,
	}
	pdfBytes, err := pdf.Render(tmpl.Name, body, opts)
	if err != nil {
		return nil, nil, err
	}
	return pdfBytes, tmpl, nil
}
