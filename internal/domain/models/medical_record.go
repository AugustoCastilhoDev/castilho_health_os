package models

import (
	"time"

	"github.com/google/uuid"
)

// MedicalRecordType matches the sector-specific evolution styles the
// clinic's professionals write in — the Content itself is free-form
// (JSON/HTML from the rich-text editor), so this is purely a
// classification/filtering dimension, not a schema switch.
type MedicalRecordType string

const (
	RecordTypeMedical    MedicalRecordType = "MEDICA"
	RecordTypeDental     MedicalRecordType = "ODONTOLOGICA"
	RecordTypePsychology MedicalRecordType = "PSICOLOGICA"
	RecordTypePsychiatry MedicalRecordType = "PSIQUIATRICA"
)

// MedicalRecord is one prontuário entry (an evolução) — an append-mostly
// clinical note written by a professional about a patient. Content is
// opaque here (JSON or HTML produced by the internal rich-text editor);
// this table only owns identity, classification and the lock state, never
// interprets the content itself.
//
// AppointmentID is optional: most evolutions are written during/after a
// scheduled encounter, but a professional may also add a note outside the
// agenda entirely (e.g. a phone follow-up), so this must not be a required
// FK.
type MedicalRecord struct {
	TenantModel
	PatientID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient        Patient      `gorm:"foreignKey:PatientID" json:"-"`
	ProfessionalID uuid.UUID    `gorm:"type:uuid;not null;index" json:"professional_id"`
	Professional   User         `gorm:"foreignKey:ProfessionalID" json:"-"`
	AppointmentID  *uuid.UUID   `gorm:"type:uuid;index" json:"appointment_id,omitempty"`
	Appointment    *Appointment `gorm:"foreignKey:AppointmentID" json:"-"`

	Type MedicalRecordType `gorm:"type:varchar(20);not null;index" json:"type"`

	// Content is the editor's serialized output (JSON doc or sanitized
	// HTML — a front-end decision, not enforced here). Never templated or
	// parsed server-side; PDF generation reads this as opaque text plus
	// whatever structured fields the editor's format exposes.
	Content string `gorm:"type:text;not null" json:"content"`

	// IsLocked marks a finalized record immutable for legal/compliance
	// reasons (a prontuário can't be silently rewritten after the fact).
	// This struct doesn't enforce that on its own — GORM will happily
	// UPDATE a locked row — the service layer must check IsLocked before
	// ever calling Update, the same way TransitionStatus is the only path
	// that may change an Appointment's status.
	IsLocked   bool       `gorm:"not null;default:false;index" json:"is_locked"`
	LockedAt   *time.Time `json:"locked_at,omitempty"`
	LockedByID *uuid.UUID `gorm:"type:uuid" json:"locked_by_id,omitempty"`
}
