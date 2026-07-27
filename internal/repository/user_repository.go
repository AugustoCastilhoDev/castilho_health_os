package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/castilho/health-os/internal/domain/models"
)

// UserRepository methods take tenantID explicitly and fold it into every
// WHERE clause — even though models.User also carries TenantID — so a
// caller can never read/write across tenants just by forgetting to set a
// field. tenantID must come from the authenticated session, never from a
// client-supplied payload.
type UserRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, user *models.User) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error)
	Update(ctx context.Context, tenantID uuid.UUID, user *models.User) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	// List returns tenant users, optionally filtered by role (nil = all
	// roles). Includes inactive users deliberately — an admin managing
	// staff needs to find and reactivate someone, not just see who's live.
	List(ctx context.Context, tenantID uuid.UUID, role *models.UserRole) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, tenantID uuid.UUID, user *models.User) error {
	user.TenantID = tenantID
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND email = ?", tenantID, email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update persists every column on user, including zero values (e.g.
// IsActive=false or clearing a council field to nil) — plain GORM Updates()
// silently skips zero-value fields, which would make "deactivate this user"
// impossible to express, so this uses Select("*") + Omit(...) to force a
// full-row update while still protecting id/tenant_id/created_at.
func (r *userRepository) Update(ctx context.Context, tenantID uuid.UUID, user *models.User) error {
	user.TenantID = tenantID
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, user.ID).
		Select("*").
		Omit("id", "tenant_id", "created_at").
		Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, role *models.UserRole) ([]models.User, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if role != nil {
		q = q.Where("role = ?", *role)
	}
	var users []models.User
	if err := q.Order("name").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
