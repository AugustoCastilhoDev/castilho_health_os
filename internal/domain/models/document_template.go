package models

// DocumentTemplateType groups templates by the kind of document they
// eventually render into (a locally-generated PDF, not a Memed
// prescription — that's MemedPrescriptionLog's domain).
type DocumentTemplateType string

const (
	TemplateTypeCertificate DocumentTemplateType = "ATESTADO"   // atestado médico/odontológico
	TemplateTypeDeclaration DocumentTemplateType = "DECLARACAO" // declaração de comparecimento
	TemplateTypeReport      DocumentTemplateType = "LAUDO"      // laudo psicológico/técnico
	TemplateTypeOther       DocumentTemplateType = "OUTRO"
)

// DocumentTemplate is a clinic-level (not patient-level) reusable text
// body for recurring documents — tenant-scoped only, no Patient/User FK.
type DocumentTemplate struct {
	TenantModel
	Name string               `gorm:"type:varchar(255);not null" json:"name"`
	Type DocumentTemplateType `gorm:"type:varchar(20);not null;index" json:"type"`

	// Content holds the template body with {{tag}}-style placeholders
	// (e.g. {{patient_name}}, {{professional_name}}, {{cid}}, {{days_off}})
	// that the PDF-generation step resolves against a specific
	// patient/appointment/professional before rendering. This table only
	// stores the reusable text, never a filled-in instance — a generated
	// PDF is either handed straight to the user or, if it should be kept
	// on file, saved back as a PatientDocument in R2.
	Content string `gorm:"type:text;not null" json:"content"`

	IsActive bool `gorm:"not null;default:true" json:"is_active"`

	// Layout toggles for the optional letterhead blocks internal/pdf.Render
	// can draw around Content — independent of {{tag}} substitution, and
	// independent of each other except IncludeStamp, which only has an
	// effect when IncludeSignature also draws the signature line it sits
	// below. Deliberately NO `default:true` gorm tag here, even though the
	// migration's column default is TRUE: GORM omits a field from the
	// INSERT whenever its Go value equals the type's zero value AND the
	// tag carries a `default`, which for a plain bool makes "the user
	// explicitly unchecked this box" indistinguishable from "never set" —
	// every unchecked checkbox silently came back true. Found live via
	// Playwright (a template created with every box unchecked still
	// rendered the full letterhead). Without the tag, GORM always sends
	// the real Go value, so Create matches the checkbox state exactly; the
	// DB column default only still matters for a hypothetical raw-SQL
	// insert that bypasses this struct entirely.
	IncludeHeader    bool `gorm:"not null" json:"include_header"`
	IncludeFooter    bool `gorm:"not null" json:"include_footer"`
	IncludeSignature bool `gorm:"not null" json:"include_signature"`
	IncludeStamp     bool `gorm:"not null" json:"include_stamp"`
}
