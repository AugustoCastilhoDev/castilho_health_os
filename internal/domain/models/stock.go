package models

import "github.com/google/uuid"

// StockItem is a supply/material the clinic keeps count of (gauze, gloves,
// anesthetic vials, ...). QuantityOnHand is the running balance derived
// from StockMovement rows — it's denormalized onto the item (rather than
// always summed from movements) so the low-stock list can be a plain index
// scan; StockMovementRepository.RecordMovement is the only writer, and it
// keeps both in sync inside one DB transaction.
type StockItem struct {
	TenantModel
	Name           string `gorm:"type:varchar(255);not null" json:"name"`
	Unit           string `gorm:"type:varchar(20);not null" json:"unit"` // "un", "cx", "par", "frasco", ...
	QuantityOnHand int    `gorm:"not null;default:0" json:"quantity_on_hand"`
	// MinQuantity is nil when the clinic hasn't set a low-stock threshold
	// for this item yet — nil means "don't warn", not "warn at zero".
	MinQuantity *int `json:"min_quantity,omitempty"`
	IsActive    bool `gorm:"not null;default:true" json:"is_active"`
}

func (StockItem) TableName() string { return "stock_items" }

// StockMovementType is deliberately just IN/OUT for v1 — no separate
// "adjustment" variant, since a signed adjustment would need its own
// direction concept anyway. A stock count correction is just an OUT (or IN)
// movement with a note explaining why.
type StockMovementType string

const (
	StockMovementIn  StockMovementType = "IN"
	StockMovementOut StockMovementType = "OUT"
)

// StockMovement is one append-only ledger entry against a StockItem.
// Quantity is always positive — Type carries the direction — matching the
// same "magnitude + direction" shape as FinancialTransaction rather than a
// signed delta, so a movement's meaning never depends on interpreting a
// sign.
type StockMovement struct {
	TenantModel
	ItemID      uuid.UUID         `gorm:"type:uuid;not null;index" json:"item_id"`
	Type        StockMovementType `gorm:"type:varchar(10);not null" json:"type"`
	Quantity    int               `gorm:"not null" json:"quantity"`
	Note        *string           `gorm:"type:text" json:"note,omitempty"`
	CreatedByID uuid.UUID         `gorm:"type:uuid;not null" json:"created_by_id"`
}

func (StockMovement) TableName() string { return "stock_movements" }
