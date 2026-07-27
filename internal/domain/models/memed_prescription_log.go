package models

import (
	"time"

	"github.com/google/uuid"
)

// MemedPrescriptionStatus is deliberately small — Memed owns the real
// prescription lifecycle (issued, dispensed, etc.); this table only needs
// enough to know whether a given prescription is still valid for audit
// purposes.
type MemedPrescriptionStatus string

const (
	MemedPrescriptionIssued    MemedPrescriptionStatus = "ISSUED"
	MemedPrescriptionCancelled MemedPrescriptionStatus = "CANCELLED"
)

// MemedPrescriptionLog is an audit trail, not the system of record for
// prescription content: the Memed front-end SDK talks to Memed directly
// and issues the prescription there, so our backend never sees (and must
// never try to reconstruct) the medication list itself — only that
// professional X issued prescription Y for patient Z at time T. This is
// what lets the clinic answer "quem prescreveu o quê e quando" without our
// system having to be a party to the actual controlled-substance
// prescription.
type MemedPrescriptionLog struct {
	TenantModel
	PatientID      uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient        Patient   `gorm:"foreignKey:PatientID" json:"-"`
	ProfessionalID uuid.UUID `gorm:"type:uuid;not null;index" json:"professional_id"`
	Professional   User      `gorm:"foreignKey:ProfessionalID" json:"-"`

	// MemedPrescriptionID is Memed's own external identifier, returned by
	// their SDK after issuance — the join key if we ever need to reconcile
	// against Memed's API/dashboard.
	MemedPrescriptionID string                  `gorm:"type:varchar(100);not null;uniqueIndex" json:"memed_prescription_id"`
	Status              MemedPrescriptionStatus `gorm:"type:varchar(20);not null;default:'ISSUED';index" json:"status"`
	IssuedAt            time.Time               `gorm:"not null" json:"issued_at"`
}
