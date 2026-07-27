package models

import "github.com/google/uuid"

// PatientDocument is metadata only — the file bytes live in Cloudflare R2,
// never in Postgres and never on this app server's disk. FileKey is the R2
// object key (not a URL): every read/write goes through a short-lived
// presigned URL generated on demand (see internal/storage), so a leaked
// FileKey alone grants no access and nothing here needs to double as a
// secret. Keying uploads as "tenants/{tenant_id}/patients/{patient_id}/..."
// (enforced by whatever service builds FileKey, not by this struct) adds a
// defense-in-depth isolation boundary at the storage layer, on top of the
// tenant_id scoping every query already does in Postgres.
type PatientDocument struct {
	TenantModel
	PatientID    uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient      Patient   `gorm:"foreignKey:PatientID" json:"-"`
	UploadedByID uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by_id"`
	UploadedBy   User      `gorm:"foreignKey:UploadedByID" json:"-"`

	FileKey  string `gorm:"type:varchar(500);not null;uniqueIndex" json:"file_key"`
	FileName string `gorm:"type:varchar(255);not null" json:"file_name"`
	FileSize int64  `gorm:"not null" json:"file_size"`
	MimeType string `gorm:"type:varchar(150);not null" json:"mime_type"`

	Description string `gorm:"type:text" json:"description,omitempty"`
}
