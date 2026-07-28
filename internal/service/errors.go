package service

import "errors"

// ErrValidation wraps input validation failures. Handlers map it to 400;
// wrap with fmt.Errorf("%w: detail", ErrValidation) to attach the reason.
var ErrValidation = errors.New("service: validation failed")

// ErrConflict flags a well-formed request that clashes with existing state
// (e.g. a tenant slug already taken). Handlers map it to 409.
var ErrConflict = errors.New("service: conflict")

// ErrStorageNotConfigured is returned by PatientDocumentService when no R2
// credentials were supplied at startup. Handlers map it to 503 — the
// request itself is fine, the deployment just isn't ready for it yet.
var ErrStorageNotConfigured = errors.New("service: document storage is not configured")

// ErrMemedNotConfigured is returned by MemedService.GetPrescriberToken when
// no MEMED_API_KEY/MEMED_SECRET_KEY were supplied at startup. Handlers map
// it to 503, same reasoning as ErrStorageNotConfigured.
var ErrMemedNotConfigured = errors.New("service: memed integration is not configured")
