package models

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentStatus is the agenda lifecycle state, per the clinic's fixed
// workflow (scheduling -> check-in -> attendance -> billing trigger).
type AppointmentStatus string

const (
	StatusScheduled  AppointmentStatus = "SCHEDULED"
	StatusConfirmed  AppointmentStatus = "CONFIRMED"
	StatusCancelled  AppointmentStatus = "CANCELLED"
	StatusWaiting    AppointmentStatus = "WAITING"
	StatusInProgress AppointmentStatus = "IN_PROGRESS"
	StatusCompleted  AppointmentStatus = "COMPLETED"
	StatusNoShow     AppointmentStatus = "NO_SHOW"
)

// validTransitions is the single source of truth for legal status moves.
// Enforce it in the service layer (e.g. Appointment.TransitionTo) so a
// direct Save() can never skip a state or resurrect a closed appointment;
// COMPLETED/CANCELLED/NO_SHOW are terminal because they are billing/audit
// triggers that must not be reopened silently.
var validTransitions = map[AppointmentStatus][]AppointmentStatus{
	StatusScheduled:  {StatusConfirmed, StatusCancelled, StatusWaiting, StatusNoShow},
	StatusConfirmed:  {StatusWaiting, StatusCancelled, StatusNoShow},
	StatusWaiting:    {StatusInProgress, StatusCancelled},
	StatusInProgress: {StatusCompleted},
	StatusCompleted:  {},
	StatusCancelled:  {},
	StatusNoShow:     {},
}

// CanTransitionTo reports whether target is a legal next state from s.
func (s AppointmentStatus) CanTransitionTo(target AppointmentStatus) bool {
	for _, allowed := range validTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// Appointment is one scheduled slot for a patient with a professional.
type Appointment struct {
	TenantModel
	PatientID      uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient        Patient   `gorm:"foreignKey:PatientID" json:"-"`
	ProfessionalID uuid.UUID `gorm:"type:uuid;not null;index" json:"professional_id"`
	Professional   User      `gorm:"foreignKey:ProfessionalID" json:"-"`

	ScheduledAt time.Time         `gorm:"not null;index" json:"scheduled_at"`
	DurationMin int               `gorm:"not null;default:30" json:"duration_min"`
	Status      AppointmentStatus `gorm:"type:varchar(20);not null;default:'SCHEDULED';index" json:"status"`

	// Denormalized timestamps for cheap reads (dashboards, wait-time KPIs).
	// The full append-only trail with actor + reason lives in
	// AppointmentStatusLog; these columns just cache "when did we last enter
	// state X" and are written by the same service call that appends a log
	// row, never independently.
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CheckedInAt *time.Time `json:"checked_in_at,omitempty"` // -> WAITING
	StartedAt   *time.Time `json:"started_at,omitempty"`    // -> IN_PROGRESS
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	NoShowAt    *time.Time `json:"no_show_at,omitempty"`

	CancellationReason string `gorm:"type:text" json:"cancellation_reason,omitempty"`

	// CID mirrors MedicalRecord.CID — same free-text, unvalidated diagnosis
	// code — kept on the appointment itself (not only on a linked
	// MedicalRecord) so a future insurance/TISS billing integration can read
	// it per encounter without joining through the prontuário. Set via
	// TransitionStatus (typically when moving to COMPLETED); left untouched
	// on transitions where no cid is supplied.
	CID *string `gorm:"column:cid;type:varchar(255)" json:"cid,omitempty"`

	StatusLogs []AppointmentStatusLog `gorm:"foreignKey:AppointmentID" json:"-"`
}

// AppointmentStatusLog is the append-only audit trail: one row per status
// change, ever. Never updated or deleted — that's what makes wait-time
// calculations (WAITING -> IN_PROGRESS) and compliance audits trustworthy.
// It embeds TenantModel (not just BaseModel) so tenant_id is queryable
// directly on this table — required if/when the database adds row-level
// security per tenant, instead of only being reachable through a join on
// appointments.
type AppointmentStatusLog struct {
	TenantModel
	AppointmentID uuid.UUID         `gorm:"type:uuid;not null;index" json:"appointment_id"`
	FromStatus    AppointmentStatus `gorm:"type:varchar(20)" json:"from_status,omitempty"`
	ToStatus      AppointmentStatus `gorm:"type:varchar(20);not null" json:"to_status"`
	// ChangedByID is the acting User, or a well-known system actor (e.g. a
	// fixed UUID for "WHATSAPP_WEBHOOK") when the transition was automated.
	ChangedByID uuid.UUID `gorm:"type:uuid;not null" json:"changed_by_id"`
	Reason      string    `gorm:"type:text" json:"reason,omitempty"`
	ChangedAt   time.Time `gorm:"not null;index" json:"changed_at"`
}
