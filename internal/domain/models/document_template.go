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
}
