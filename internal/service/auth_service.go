// Package service holds application/business logic that coordinates
// multiple repositories — the layer between HTTP handlers and persistence.
package service

import (
	"context"
	"errors"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/repository"
)

// ErrInvalidCredentials is returned for every login failure mode (unknown
// tenant, unknown email, wrong password, inactive account) — deliberately
// generic so a failed login never reveals which of those it was, which
// would let an attacker enumerate valid tenant slugs or emails.
var ErrInvalidCredentials = errors.New("service: invalid credentials")

type AuthService struct {
	tenants repository.TenantRepository
	users   repository.UserRepository
	issuer  *auth.JWTIssuer
}

func NewAuthService(tenants repository.TenantRepository, users repository.UserRepository, issuer *auth.JWTIssuer) *AuthService {
	return &AuthService{tenants: tenants, users: users, issuer: issuer}
}

func (s *AuthService) Login(ctx context.Context, tenantSlug, email, password string) (string, error) {
	tenant, err := s.tenants.FindBySlug(ctx, tenantSlug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if !tenant.IsActive {
		return "", ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, tenant.ID, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if !user.IsActive || !auth.CheckPassword(user.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	return s.issuer.Issue(user.ID, tenant.ID, user.Role)
}
