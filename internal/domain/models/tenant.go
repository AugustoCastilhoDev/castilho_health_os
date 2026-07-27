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

	// Settings holds clinic-specific configuration that changes per tenant
	// without a migration, e.g.:
	//   {"deduct_card_fee_before_split": true, "no_show_grace_minutes": 15,
	//    "whatsapp_confirmation_hours_before": 24}
	Settings datatypes.JSONMap `gorm:"type:jsonb" json:"settings,omitempty"`

	Users    []User    `gorm:"foreignKey:TenantID" json:"-"`
	Patients []Patient `gorm:"foreignKey:TenantID" json:"-"`
}
