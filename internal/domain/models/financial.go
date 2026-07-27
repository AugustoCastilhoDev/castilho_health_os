package models

import (
	"time"

	"github.com/google/uuid"
)

// FinancialRuleType selects how a professional's repasse is computed.
type FinancialRuleType string

const (
	RuleTypePercentage          FinancialRuleType = "PERCENTAGE"
	RuleTypeFixedPerAppointment FinancialRuleType = "FIXED_PER_APPOINTMENT"
	RuleTypeFixedPerProcedure   FinancialRuleType = "FIXED_PER_PROCEDURE"
)

// FeeDeductionPolicy controls whether admin/card fees are subtracted from
// the gross amount before the professional's percentage is applied, per
// clinic configuration.
type FeeDeductionPolicy string

const (
	// DeductFeesBeforeSplit: professional's cut is computed on (gross - fee).
	DeductFeesBeforeSplit FeeDeductionPolicy = "BEFORE_SPLIT"
	// DeductFeesAfterSplit: professional's cut is computed on the full gross;
	// the fee comes out of the clinic's remaining share only.
	DeductFeesAfterSplit FeeDeductionPolicy = "AFTER_SPLIT"
)

// FinancialRule configures how a given professional is paid. Multiple rules
// may exist per professional (e.g. one for private/particular, one per
// insurance plan, one for a specific high-value procedure); ProcedureCode
// and InsurancePlan let a rule narrow its scope, and Priority breaks ties
// when more than one rule matches the same transaction — the calculation
// service must pick the highest-priority match, never merge or sum rules.
type FinancialRule struct {
	TenantModel
	ProfessionalID uuid.UUID `gorm:"type:uuid;not null;index" json:"professional_id"`
	Professional   User      `gorm:"foreignKey:ProfessionalID" json:"-"`

	Type FinancialRuleType `gorm:"type:varchar(30);not null" json:"type"`

	// Percentage as a fraction (0.7000 = 70%). Set only when Type == PERCENTAGE.
	Percentage *float64 `gorm:"type:numeric(5,4)" json:"percentage,omitempty"`
	// FixedAmountCents. Set only when Type is one of the FIXED_* variants.
	FixedAmountCents *int64 `json:"fixed_amount_cents,omitempty"`

	// Scope: nil on either field means "matches anything not claimed by a
	// more specific rule" for that dimension.
	ProcedureCode *string `gorm:"type:varchar(50);index" json:"procedure_code,omitempty"`
	InsurancePlan *string `gorm:"type:varchar(100);index" json:"insurance_plan,omitempty"` // nil = particular

	FeeDeduction FeeDeductionPolicy `gorm:"type:varchar(20);not null;default:'BEFORE_SPLIT'" json:"fee_deduction"`

	Priority int  `gorm:"not null;default:0" json:"priority"`
	IsActive bool `gorm:"not null;default:true" json:"is_active"`
}

// TransactionType distinguishes money coming from the patient from money
// going out to the professional. They are modeled as the same entity
// because both share amount/fee/status/payment-method plumbing, and a
// payout transaction always references the income transaction that funded
// it (SourceTransactionID).
type TransactionType string

const (
	TransactionPatientPayment     TransactionType = "PATIENT_PAYMENT"     // receivable (in)
	TransactionProfessionalPayout TransactionType = "PROFESSIONAL_PAYOUT" // payable / repasse (out)
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "PENDING"
	TransactionPaid      TransactionStatus = "PAID"
	TransactionCancelled TransactionStatus = "CANCELLED"
)

type PaymentMethod string

const (
	PaymentCash         PaymentMethod = "CASH"
	PaymentDebitCard    PaymentMethod = "DEBIT_CARD"
	PaymentCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentPix          PaymentMethod = "PIX"
	PaymentBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentInsurance    PaymentMethod = "INSURANCE" // convênio
)

// FinancialTransaction is one ledger entry, either patient-facing (income)
// or professional-facing (repasse/expense). A PROFESSIONAL_PAYOUT row must
// only be created once its Appointment.Status == COMPLETED and the source
// PATIENT_PAYMENT.Status == PAID (or the clinic's "pay on production billed"
// rule is satisfied) — that gating lives in the settlement service, not
// here, but SourceTransactionID and AppointmentID are what let it enforce
// that check.
type FinancialTransaction struct {
	TenantModel
	AppointmentID  *uuid.UUID `gorm:"type:uuid;index" json:"appointment_id,omitempty"`
	PatientID      *uuid.UUID `gorm:"type:uuid;index" json:"patient_id,omitempty"`
	ProfessionalID *uuid.UUID `gorm:"type:uuid;index" json:"professional_id,omitempty"`

	// SourceTransactionID: set only on PROFESSIONAL_PAYOUT rows, pointing at
	// the PATIENT_PAYMENT row it was derived from.
	SourceTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"source_transaction_id,omitempty"`
	FinancialRuleID     *uuid.UUID `gorm:"type:uuid" json:"financial_rule_id,omitempty"`

	Type   TransactionType   `gorm:"type:varchar(30);not null;index" json:"type"`
	Status TransactionStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`

	// All amounts in cents (int64) — never float64 — to keep repasse math exact.
	GrossAmountCents int64 `gorm:"not null" json:"gross_amount_cents"`
	FeeAmountCents   int64 `gorm:"not null;default:0" json:"fee_amount_cents"` // e.g. card acquirer fee
	NetAmountCents   int64 `gorm:"not null" json:"net_amount_cents"`           // what actually settles

	// PaymentMethod is a pointer because it's genuinely optional (e.g. a
	// PROFESSIONAL_PAYOUT row may not have a disbursement method decided
	// yet) — as a plain string, GORM would insert '' instead of SQL NULL
	// for the zero value, which the payment_method CHECK constraint rejects.
	PaymentMethod *PaymentMethod `gorm:"type:varchar(20)" json:"payment_method,omitempty"`

	// ProcedureCode/InsurancePlan let a PATIENT_PAYMENT carry the same scope
	// dimensions FinancialRule matches on, so the settlement service can call
	// FindApplicable with real data instead of always falling back to a
	// professional's wildcard rule.
	ProcedureCode *string `gorm:"type:varchar(50);index" json:"procedure_code,omitempty"`
	InsurancePlan *string `gorm:"type:varchar(100);index" json:"insurance_plan,omitempty"`

	DueDate *time.Time `json:"due_date,omitempty"`
	PaidAt  *time.Time `json:"paid_at,omitempty"`

	Notes string `gorm:"type:text" json:"notes,omitempty"`
}
