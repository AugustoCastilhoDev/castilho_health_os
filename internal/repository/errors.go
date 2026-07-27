// Package repository is the persistence layer: one file per aggregate,
// GORM underneath, no business/pricing logic. The one exception is
// AppointmentRepository.TransitionStatus, which enforces the appointment
// state machine and writes its audit log atomically — that's a data
// integrity invariant (the log must never be skippable), not a business
// rule, so it belongs here rather than in a service layer above.
package repository

import "errors"

var (
	// ErrNotFound is returned when a lookup by ID (or other unique key)
	// matches no row, so callers check this instead of reaching into
	// gorm.ErrRecordNotFound directly.
	ErrNotFound = errors.New("repository: record not found")

	// ErrInvalidTransition is returned when a requested appointment status
	// change is not a legal move per models.AppointmentStatus.CanTransitionTo.
	ErrInvalidTransition = errors.New("repository: invalid appointment status transition")
)
