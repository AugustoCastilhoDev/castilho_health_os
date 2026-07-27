// Package models contains the GORM entities shared by every module of
// Castilho Health OS. Money is always stored as int64 cents (never float64)
// to avoid rounding drift in repasse/commission math. UUIDs are generated in
// Go (not left to a Postgres default) so the ID is available before INSERT
// and the schema has no dependency on the pgcrypto/uuid-ossp extensions.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded by entities that are NOT scoped to a tenant
// (currently only Tenant itself).
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns the UUID in application code so callers can rely on
// ID being populated immediately after a zero-value struct is created.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// TenantModel is embedded by every entity that belongs to a clinic.
// TenantID must always come from the authenticated session/JWT claims in the
// service layer, never from client-supplied payload fields — this is the
// isolation boundary the whole multi-tenant model relies on.
type TenantModel struct {
	BaseModel
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:idx_tenant" json:"tenant_id"`
}
