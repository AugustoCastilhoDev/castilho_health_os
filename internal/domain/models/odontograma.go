package models

import (
	"fmt"

	"github.com/google/uuid"
)

// ToothCondition is the whole-tooth status recorded by an OdontogramaEntry.
// v1 tracks only whole-tooth state — a per-surface (mesial/distal/oclusal/
// vestibular/lingual) model was considered and deferred: a coarser chart is
// still clinically useful and far simpler to build and read at a glance.
type ToothCondition string

const (
	ToothConditionHealthy             ToothCondition = "SAUDAVEL"
	ToothConditionCavity              ToothCondition = "CARIE"
	ToothConditionRestored            ToothCondition = "RESTAURADO"
	ToothConditionMissing             ToothCondition = "AUSENTE"
	ToothConditionRootCanal           ToothCondition = "CANAL"
	ToothConditionCrown               ToothCondition = "COROA"
	ToothConditionImplant             ToothCondition = "IMPLANTE"
	ToothConditionFractured           ToothCondition = "FRATURADO"
	ToothConditionExtractionIndicated ToothCondition = "A_EXTRAIR"
)

// Valid reports whether c is one of the fixed conditions above — used by the
// service layer the same way movement/record-type enums are checked
// elsewhere in this app.
func (c ToothCondition) Valid() bool {
	switch c {
	case ToothConditionHealthy, ToothConditionCavity, ToothConditionRestored, ToothConditionMissing,
		ToothConditionRootCanal, ToothConditionCrown, ToothConditionImplant, ToothConditionFractured,
		ToothConditionExtractionIndicated:
		return true
	default:
		return false
	}
}

// validToothNumbers is the fixed FDI-notation whitelist this app accepts:
// quadrants 1-4 (8 teeth each) are the 32 permanent teeth, quadrants 5-8 (5
// teeth each) are the 20 deciduous teeth. Both sets are always available —
// a patient can have entries in both at once, which is clinically normal
// during the mixed-dentition years — rather than switching the chart by the
// patient's age. The Postgres CHECK constraint on this column only
// restricts the two-digit shape (see migration 000011); this map is the
// actual whitelist enforced in the service layer.
var validToothNumbers = func() map[string]bool {
	set := make(map[string]bool, 52)
	for _, quadrant := range []int{1, 2, 3, 4} {
		for position := 1; position <= 8; position++ {
			set[fmt.Sprintf("%d%d", quadrant, position)] = true
		}
	}
	for _, quadrant := range []int{5, 6, 7, 8} {
		for position := 1; position <= 5; position++ {
			set[fmt.Sprintf("%d%d", quadrant, position)] = true
		}
	}
	return set
}()

// IsValidToothNumber reports whether n is a known FDI tooth code (permanent
// or deciduous).
func IsValidToothNumber(n string) bool {
	return validToothNumbers[n]
}

// OdontogramaEntry is one append-only finding/procedure record against a
// single tooth — same "ledger, never edited" shape as StockMovement. There
// is deliberately no separate "current chart" row: a tooth's displayed
// status is always the most recent entry for that (patient, tooth_number)
// pair, computed with DISTINCT ON in OdontogramaRepository.CurrentChart.
// Unlike StockItem.QuantityOnHand, a condition change is a plain
// replacement rather than a numeric delta, so there's no running balance
// that needs a locked transaction to stay consistent.
type OdontogramaEntry struct {
	TenantModel
	PatientID uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	// ToothNumber uses FDI two-digit notation ("11".."48" permanent,
	// "51".."85" deciduous) — validated against IsValidToothNumber, never
	// against an open string.
	ToothNumber  string         `gorm:"type:varchar(2);not null;index" json:"tooth_number"`
	Condition    ToothCondition `gorm:"type:varchar(20);not null" json:"condition"`
	Note         *string        `gorm:"type:text" json:"note,omitempty"`
	RecordedByID uuid.UUID      `gorm:"type:uuid;not null" json:"recorded_by_id"`
}

func (OdontogramaEntry) TableName() string { return "odontograma_entries" }
