package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
)

type StockService struct {
	items     repository.StockItemRepository
	movements repository.StockMovementRepository
}

func NewStockService(items repository.StockItemRepository, movements repository.StockMovementRepository) *StockService {
	return &StockService{items: items, movements: movements}
}

func validateItemFields(name, unit string, minQuantity *int) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if unit == "" {
		return fmt.Errorf("%w: unit is required", ErrValidation)
	}
	if minQuantity != nil && *minQuantity < 0 {
		return fmt.Errorf("%w: min_quantity must not be negative", ErrValidation)
	}
	return nil
}

// CreateItem takes an optional initialQuantity rather than accepting a
// caller-supplied QuantityOnHand directly — the running balance must only
// ever move through a StockMovement, so "I already have 20 of these"
// becomes a real IN movement (visible in the item's history) instead of a
// silent starting number nothing explains later.
func (s *StockService) CreateItem(ctx context.Context, tenantID uuid.UUID, item *models.StockItem, initialQuantity int, createdByID uuid.UUID) (*models.StockItem, error) {
	if err := validateItemFields(item.Name, item.Unit, item.MinQuantity); err != nil {
		return nil, err
	}
	if initialQuantity < 0 {
		return nil, fmt.Errorf("%w: initial_quantity must not be negative", ErrValidation)
	}
	item.QuantityOnHand = 0
	item.IsActive = true
	if err := s.items.Create(ctx, tenantID, item); err != nil {
		return nil, err
	}
	if initialQuantity > 0 {
		note := "Estoque inicial"
		updated, err := s.movements.RecordMovement(ctx, tenantID, &models.StockMovement{
			TenantModel: models.TenantModel{TenantID: tenantID},
			ItemID:      item.ID,
			Type:        models.StockMovementIn,
			Quantity:    initialQuantity,
			Note:        &note,
			CreatedByID: createdByID,
		})
		if err != nil {
			return nil, err
		}
		return updated, nil
	}
	return item, nil
}

func (s *StockService) GetItem(ctx context.Context, tenantID, id uuid.UUID) (*models.StockItem, error) {
	return s.items.FindByID(ctx, tenantID, id)
}

func (s *StockService) UpdateItem(ctx context.Context, tenantID uuid.UUID, item *models.StockItem) (*models.StockItem, error) {
	if err := validateItemFields(item.Name, item.Unit, item.MinQuantity); err != nil {
		return nil, err
	}
	if err := s.items.Update(ctx, tenantID, item); err != nil {
		return nil, err
	}
	return s.items.FindByID(ctx, tenantID, item.ID)
}

func (s *StockService) ListItems(ctx context.Context, tenantID uuid.UUID) ([]models.StockItem, error) {
	return s.items.List(ctx, tenantID)
}

func (s *StockService) RecordMovement(ctx context.Context, tenantID uuid.UUID, movement *models.StockMovement) (*models.StockItem, error) {
	if movement.ItemID == uuid.Nil {
		return nil, fmt.Errorf("%w: item_id is required", ErrValidation)
	}
	if movement.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be positive", ErrValidation)
	}
	switch movement.Type {
	case models.StockMovementIn, models.StockMovementOut:
	default:
		return nil, fmt.Errorf("%w: unknown movement type %q", ErrValidation, movement.Type)
	}
	return s.movements.RecordMovement(ctx, tenantID, movement)
}

func (s *StockService) ListMovementsByItem(ctx context.Context, tenantID, itemID uuid.UUID) ([]models.StockMovement, error) {
	return s.movements.ListByItem(ctx, tenantID, itemID)
}
