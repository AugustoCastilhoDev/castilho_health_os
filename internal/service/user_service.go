package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/auth"
	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func validRole(role models.UserRole) bool {
	switch role {
	case models.RoleTenantAdmin, models.RoleDoctor, models.RoleDentist, models.RoleReceptionist, models.RoleFinance,
		models.RolePsychologist, models.RolePsychiatrist:
		return true
	default:
		return false
	}
}

type CreateUserInput struct {
	Name          string
	Email         string
	Password      string
	Role          models.UserRole
	CouncilType   *string
	CouncilNumber *string
	CouncilState  *string
	// CPF/BirthDate/Sex/Phone are only stored for health professionals — see
	// Memed prescriber registration (internal/memed) for why they exist.
	CPF       *string
	BirthDate *time.Time
	Sex       *models.UserSex
	Phone     *string
}

func (s *UserService) Create(ctx context.Context, tenantID uuid.UUID, in CreateUserInput) (*models.User, error) {
	if in.Name == "" || in.Email == "" {
		return nil, fmt.Errorf("%w: name and email are required", ErrValidation)
	}
	if len(in.Password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", ErrValidation)
	}
	if !validRole(in.Role) {
		return nil, fmt.Errorf("%w: unknown role %q", ErrValidation, in.Role)
	}

	if _, err := s.users.FindByEmail(ctx, tenantID, in.Email); err == nil {
		return nil, fmt.Errorf("%w: email %q is already in use", ErrConflict, in.Email)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		Role:         in.Role,
		IsActive:     true,
	}
	if in.Role.IsHealthProfessional() {
		user.CouncilType = in.CouncilType
		user.CouncilNumber = in.CouncilNumber
		user.CouncilState = in.CouncilState
		user.CPF = in.CPF
		user.BirthDate = in.BirthDate
		user.Sex = in.Sex
		user.Phone = in.Phone
	}

	if err := s.users.Create(ctx, tenantID, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Get(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	return s.users.FindByID(ctx, tenantID, id)
}

func (s *UserService) List(ctx context.Context, tenantID uuid.UUID, role *models.UserRole) ([]models.User, error) {
	return s.users.List(ctx, tenantID, role)
}

type UpdateUserInput struct {
	Name          string
	Email         string
	Role          models.UserRole
	IsActive      bool
	CouncilType   *string
	CouncilNumber *string
	CouncilState  *string
	CPF           *string
	BirthDate     *time.Time
	Sex           *models.UserSex
	Phone         *string
}

// Update loads the existing row and mutates only the editable profile
// fields before saving, instead of building a fresh *models.User from the
// request body. The repository's Update does a full-row replace
// (Select("*")) so any field absent from the request DTO would otherwise
// be zeroed out — for most resources that's fine (PUT-replaces-everything
// is the documented contract), but User carries PasswordHash, which must
// never be touched here; ChangeOwnPassword/ResetPassword own that instead.
func (s *UserService) Update(ctx context.Context, tenantID, id uuid.UUID, in UpdateUserInput) (*models.User, error) {
	if in.Name == "" || in.Email == "" {
		return nil, fmt.Errorf("%w: name and email are required", ErrValidation)
	}
	if !validRole(in.Role) {
		return nil, fmt.Errorf("%w: unknown role %q", ErrValidation, in.Role)
	}

	existing, err := s.users.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	existing.Name = in.Name
	existing.Email = in.Email
	existing.Role = in.Role
	existing.IsActive = in.IsActive
	if in.Role.IsHealthProfessional() {
		existing.CouncilType = in.CouncilType
		existing.CouncilNumber = in.CouncilNumber
		existing.CouncilState = in.CouncilState
		existing.CPF = in.CPF
		existing.BirthDate = in.BirthDate
		existing.Sex = in.Sex
		existing.Phone = in.Phone
	} else {
		existing.CouncilType = nil
		existing.CouncilNumber = nil
		existing.CouncilState = nil
		existing.CPF = nil
		existing.BirthDate = nil
		existing.Sex = nil
		existing.Phone = nil
	}

	if err := s.users.Update(ctx, tenantID, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// ChangeOwnPassword requires the current password — this is the
// self-service path (any authenticated user changing their own login).
func (s *UserService) ChangeOwnPassword(ctx context.Context, tenantID, userID uuid.UUID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: new password must be at least 8 characters", ErrValidation)
	}
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(user.PasswordHash, oldPassword) {
		return fmt.Errorf("%w: current password is incorrect", ErrValidation)
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.users.Update(ctx, tenantID, user)
}

// ResetPassword is the admin path — no old password required, for staff
// who forgot theirs (there's no email-based reset flow yet).
func (s *UserService) ResetPassword(ctx context.Context, tenantID, userID uuid.UUID, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: new password must be at least 8 characters", ErrValidation)
	}
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.users.Update(ctx, tenantID, user)
}
