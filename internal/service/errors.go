package service

import "errors"

// ErrValidation wraps input validation failures. Handlers map it to 400;
// wrap with fmt.Errorf("%w: detail", ErrValidation) to attach the reason.
var ErrValidation = errors.New("service: validation failed")

// ErrConflict flags a well-formed request that clashes with existing state
// (e.g. a tenant slug already taken). Handlers map it to 409.
var ErrConflict = errors.New("service: conflict")
