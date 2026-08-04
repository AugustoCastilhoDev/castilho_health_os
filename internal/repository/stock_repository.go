package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/castilho/health-os/internal/domain/models"
)

type StockItemRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.StockItem, error)
	// FindByName does a case-insensitive match — used by CSV import to skip
	// a row whose name already exists in the tenant's stock rather than
	// creating a duplicate. Returns ErrNotFound like FindByID when no match.
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.StockItem, error)
	// Update replaces name/unit/min_quantity/is_active only — never
	// quantity_on_hand, which is owned exclusively by
	// StockMovementRepository.RecordMovement.
	Update(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error
	List(ctx context.Context, tenantID uuid.UUID) ([]models.StockItem, error)
}

type stockItemRepository struct {
	db *gorm.DB
}

func NewStockItemRepository(db *gorm.DB) StockItemRepository {
	return &stockItemRepository{db: db}
}

func (r *stockItemRepository) Create(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error {
	item.TenantID = tenantID
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *stockItemRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*models.StockItem, error) {
	var item models.StockItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *stockItemRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.StockItem, error) {
	var item models.StockItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(name) = LOWER(?)", tenantID, name).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *stockItemRepository) Update(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) error {
	item.TenantID = tenantID
	result := r.db.WithContext(ctx).
		Model(&models.StockItem{}).
		Where("tenant_id = ? AND id = ?", tenantID, item.ID).
		Updates(map[string]any{
			"name":         item.Name,
			"unit":         item.Unit,
			"min_quantity": item.MinQuantity,
			"is_active":    item.IsActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *stockItemRepository) List(ctx context.Context, tenantID uuid.UUID) ([]models.StockItem, error) {
	var items []models.StockItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("name").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

type StockMovementRepository interface {
	// RecordMovement locks the item row, applies the quantity delta, and
	// inserts the ledger entry atomically — the same "lock, validate,
	// mutate, log" shape as AppointmentRepository.TransitionStatus. Returns
	// the item's new state alongside the persisted movement so the caller
	// doesn't need a second round trip to show the updated balance.
	RecordMovement(ctx context.Context, tenantID uuid.UUID, movement *models.StockMovement) (*models.StockItem, error)
	ListByItem(ctx context.Context, tenantID, itemID uuid.UUID) ([]models.StockMovement, error)
}

type stockMovementRepository struct {
	db *gorm.DB
}

func NewStockMovementRepository(db *gorm.DB) StockMovementRepository {
	return &stockMovementRepository{db: db}
}

func (r *stockMovementRepository) RecordMovement(ctx context.Context, tenantID uuid.UUID, movement *models.StockMovement) (*models.StockItem, error) {
	var item models.StockItem

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND id = ?", tenantID, movement.ItemID).
			First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		switch movement.Type {
		case models.StockMovementIn:
			item.QuantityOnHand += movement.Quantity
		case models.StockMovementOut:
			if item.QuantityOnHand < movement.Quantity {
				return ErrInsufficientStock
			}
			item.QuantityOnHand -= movement.Quantity
		default:
			return fmt.Errorf("repository: unknown stock movement type %q", movement.Type)
		}

		if err := tx.Model(&models.StockItem{}).
			Where("tenant_id = ? AND id = ?", tenantID, item.ID).
			Update("quantity_on_hand", item.QuantityOnHand).Error; err != nil {
			return err
		}

		movement.TenantID = tenantID
		return tx.Create(movement).Error
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *stockMovementRepository) ListByItem(ctx context.Context, tenantID, itemID uuid.UUID) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND item_id = ?", tenantID, itemID).
		Order("created_at DESC").
		Find(&movements).Error
	if err != nil {
		return nil, err
	}
	return movements, nil
}
