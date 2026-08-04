package models

import "gorm.io/datatypes"

// TenantType controls which modules the frontend exposes for a clinic.
type TenantType string

const (
	TenantTypeMedica TenantType = "MEDICA"
	TenantTypeOdonto TenantType = "ODONTO"
	TenantTypeMista  TenantType = "MISTA"
)

// Tenant represents one clinic. It is the root of the multi-tenant tree —
// every other business table carries a TenantID that must resolve to a row
// here.
type Tenant struct {
	BaseModel
	Name     string     `gorm:"type:varchar(255);not null" json:"name"`
	Slug     string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"` // subdomínio: {slug}.castilhohealth.com.br
	Type     TenantType `gorm:"type:varchar(20);not null" json:"type"`
	Document string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"document"` // CNPJ
	Email    string     `gorm:"type:varchar(255);not null" json:"email"`
	Phone    string     `gorm:"type:varchar(20)" json:"phone,omitempty"`
	IsActive bool       `gorm:"not null;default:true" json:"is_active"`

	// Letterhead data for generated PDFs (internal/pdf.Render) — all
	// optional, filled in via the Configurações screen. AddressXxx are kept
	// as separate columns (not one free-text line) to match the pattern
	// already used for Patient's address fields.
	AddressStreet *string `gorm:"type:varchar(255)" json:"address_street,omitempty"`
	AddressCity   *string `gorm:"type:varchar(100)" json:"address_city,omitempty"`
	AddressState  *string `gorm:"type:varchar(2)" json:"address_state,omitempty"`
	AddressZip    *string `gorm:"type:varchar(10)" json:"address_zip,omitempty"`
	// LogoKey is an R2 object key, never a public URL — resolved through a
	// presigned URL on demand (frontend preview) or downloaded server-side
	// into PDF bytes (letterhead header), same access pattern as
	// PatientDocument.FileKey.
	LogoKey *string `gorm:"type:varchar(500)" json:"logo_key,omitempty"`

	// Settings holds clinic-specific configuration that changes per tenant
	// without a migration, e.g.:
	//   {"deduct_card_fee_before_split": true, "no_show_grace_minutes": 15,
	//    "whatsapp_confirmation_hours_before": 24}
	Settings datatypes.JSONMap `gorm:"type:jsonb" json:"settings,omitempty"`

	Users    []User    `gorm:"foreignKey:TenantID" json:"-"`
	Patients []Patient `gorm:"foreignKey:TenantID" json:"-"`
}
