package models

import "time"

// Patient holds registration data shared by the medical and dental modules.
// Clinical record content (PEP/prontuário) is intentionally not modeled
// here yet — it belongs to its own auditable, append-only structure and is
// a separate step from this base schema.
type Patient struct {
	TenantModel
	Name      string     `gorm:"type:varchar(255);not null;index" json:"name"`
	Document  string     `gorm:"type:varchar(20);index" json:"document,omitempty"` // CPF
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Phone     string     `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Email     string     `gorm:"type:varchar(255)" json:"email,omitempty"`

	AddressZip    string `gorm:"type:varchar(10)" json:"address_zip,omitempty"`
	AddressStreet string `gorm:"type:varchar(255)" json:"address_street,omitempty"`
	AddressCity   string `gorm:"type:varchar(100)" json:"address_city,omitempty"`
	AddressState  string `gorm:"type:varchar(2)" json:"address_state,omitempty"`

	Appointments []Appointment `gorm:"foreignKey:PatientID" json:"-"`
}
