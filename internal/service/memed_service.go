package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

// MemedClient is the slice of *memed.Client this service needs — defined
// here (not imported from internal/memed) so tests can fake it, mirroring
// ObjectStorage's split from internal/storage in patient_document_service.go.
type MemedClient interface {
	FetchOrRegisterToken(ctx context.Context, p MemedPrescriber) (string, error)
}

// MemedPrescriber mirrors memed.Prescriber field-for-field. It's redeclared
// in this package rather than imported so this package doesn't depend on
// internal/memed (and, transitively, on nothing outside the standard
// library) — the same reasoning as MemedClient itself.
type MemedPrescriber struct {
	ExternalID  string
	Name        string
	Surname     string
	CPF         string
	BoardCode   string
	BoardNumber string
	BoardState  string
	BirthDate   time.Time
	Email       string
	Phone       string
	Sex         string
}

// MemedService is usable with client == nil (no MEMED_API_KEY/SECRET_KEY
// configured yet, e.g. CI or a fresh clone) — GetPrescriberToken returns
// ErrMemedNotConfigured instead of a nil-pointer panic, the same pattern
// PatientDocumentService uses for R2. RecordIssuance/ListByPatient/Cancel
// never need the client at all — they only ever touch the local audit log.
type MemedService struct {
	logs   repository.MemedPrescriptionLogRepository
	users  repository.UserRepository
	client MemedClient
}

func NewMemedService(logs repository.MemedPrescriptionLogRepository, users repository.UserRepository, client MemedClient) *MemedService {
	return &MemedService{logs: logs, users: users, client: client}
}

// GetPrescriberToken fetches (or, on first use, registers) the Memed
// prescriber token for the given professional, so the frontend can load
// Memed's own prescription widget. The token is intentionally not cached —
// Memed's docs say it isn't fixed, and should be re-fetched per session.
func (s *MemedService) GetPrescriberToken(ctx context.Context, tenantID, professionalID uuid.UUID) (string, error) {
	if s.client == nil {
		return "", ErrMemedNotConfigured
	}
	user, err := s.users.FindByID(ctx, tenantID, professionalID)
	if err != nil {
		return "", err
	}
	if !user.Role.IsHealthProfessional() {
		return "", fmt.Errorf("%w: only DOCTOR/DENTIST users can issue Memed prescriptions", ErrValidation)
	}
	if user.CPF == nil || user.BirthDate == nil || user.CouncilType == nil || user.CouncilNumber == nil || user.CouncilState == nil {
		return "", fmt.Errorf("%w: this professional is missing cpf/birth_date/council registration required by Memed — complete their profile first", ErrValidation)
	}

	surname, firstName := splitName(user.Name)
	prescriber := MemedPrescriber{
		ExternalID:  user.ID.String(),
		Name:        firstName,
		Surname:     surname,
		CPF:         *user.CPF,
		BoardCode:   *user.CouncilType,
		BoardNumber: *user.CouncilNumber,
		BoardState:  *user.CouncilState,
		BirthDate:   *user.BirthDate,
		Email:       user.Email,
	}
	if user.Phone != nil {
		prescriber.Phone = *user.Phone
	}
	if user.Sex != nil {
		prescriber.Sex = string(*user.Sex)
	}

	return s.client.FetchOrRegisterToken(ctx, prescriber)
}

// splitName divides a full name into (first name, rest) the way Memed's
// nome/sobrenome fields expect — this app stores a single Name string
// (see models.User), so there's no better source of truth than a naive
// split on the first space.
func splitName(fullName string) (surname, firstName string) {
	for i, r := range fullName {
		if r == ' ' {
			return fullName[i+1:], fullName[:i]
		}
	}
	return fullName, fullName
}

// RecordIssuance persists the audit-only log row once the frontend's Memed
// widget reports a prescription was issued (the prescricaoImpressa event) —
// this service never sees the prescription's actual content.
func (s *MemedService) RecordIssuance(ctx context.Context, tenantID, patientID, professionalID uuid.UUID, memedPrescriptionID string) (*models.MemedPrescriptionLog, error) {
	if memedPrescriptionID == "" {
		return nil, fmt.Errorf("%w: memed_prescription_id is required", ErrValidation)
	}
	if patientID == uuid.Nil {
		return nil, fmt.Errorf("%w: patient_id is required", ErrValidation)
	}

	if _, err := s.logs.FindByMemedID(ctx, tenantID, memedPrescriptionID); err == nil {
		return nil, fmt.Errorf("%w: prescription %q was already logged", ErrConflict, memedPrescriptionID)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	log := &models.MemedPrescriptionLog{
		PatientID:           patientID,
		ProfessionalID:      professionalID,
		MemedPrescriptionID: memedPrescriptionID,
		Status:              models.MemedPrescriptionIssued,
		IssuedAt:            time.Now(),
	}
	if err := s.logs.Create(ctx, tenantID, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *MemedService) ListByPatient(ctx context.Context, tenantID, patientID uuid.UUID) ([]models.MemedPrescriptionLog, error) {
	return s.logs.ListByPatient(ctx, tenantID, patientID)
}

// Cancel marks a previously logged prescription CANCELLED — mirrors Memed's
// own prescricaoExcluida event (a professional deleting a prescription
// before it's dispensed).
func (s *MemedService) Cancel(ctx context.Context, tenantID uuid.UUID, memedPrescriptionID string) error {
	return s.logs.CancelByMemedID(ctx, tenantID, memedPrescriptionID)
}
