package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/service"
)

func TestStockService_CreateItem_Validation(t *testing.T) {
	cases := []struct {
		name string
		item *models.StockItem
	}{
		{"missing name", &models.StockItem{Unit: "un"}},
		{"missing unit", &models.StockItem{Name: "Luva"}},
		{"negative min quantity", &models.StockItem{Name: "Luva", Unit: "un", MinQuantity: intPtr(-1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewStockService(&fakeStockItemRepo{}, &fakeStockMovementRepo{})
			_, err := svc.CreateItem(context.Background(), uuid.New(), tc.item, 0, uuid.New())
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestStockService_CreateItem_StartsAtZeroWithoutInitialQuantity(t *testing.T) {
	var created *models.StockItem
	items := &fakeStockItemRepo{
		createFn: func(_ context.Context, tenantID uuid.UUID, item *models.StockItem) error {
			item.ID = uuid.New()
			created = item
			return nil
		},
	}
	movements := &fakeStockMovementRepo{
		recordMovementFn: func(context.Context, uuid.UUID, *models.StockMovement) (*models.StockItem, error) {
			t.Fatal("must not record a movement when initialQuantity is 0")
			return nil, nil
		},
	}
	svc := service.NewStockService(items, movements)

	saved, err := svc.CreateItem(context.Background(), uuid.New(), &models.StockItem{Name: "Luva", Unit: "un"}, 0, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, saved.QuantityOnHand)
	assert.True(t, created.IsActive)
}

func TestStockService_CreateItem_RecordsInitialQuantityAsInMovement(t *testing.T) {
	items := &fakeStockItemRepo{
		createFn: func(_ context.Context, tenantID uuid.UUID, item *models.StockItem) error {
			item.ID = uuid.New()
			return nil
		},
	}
	var recordedMovement *models.StockMovement
	movements := &fakeStockMovementRepo{
		recordMovementFn: func(_ context.Context, _ uuid.UUID, m *models.StockMovement) (*models.StockItem, error) {
			recordedMovement = m
			return &models.StockItem{Name: "Luva", Unit: "un", QuantityOnHand: 20}, nil
		},
	}
	svc := service.NewStockService(items, movements)

	saved, err := svc.CreateItem(context.Background(), uuid.New(), &models.StockItem{Name: "Luva", Unit: "un"}, 20, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 20, saved.QuantityOnHand)
	require.NotNil(t, recordedMovement)
	assert.Equal(t, models.StockMovementIn, recordedMovement.Type)
	assert.Equal(t, 20, recordedMovement.Quantity)
}

func TestStockService_RecordMovement_Validation(t *testing.T) {
	cases := []struct {
		name     string
		movement *models.StockMovement
	}{
		{"missing item_id", &models.StockMovement{Type: models.StockMovementIn, Quantity: 1}},
		{"zero quantity", &models.StockMovement{ItemID: uuid.New(), Type: models.StockMovementIn, Quantity: 0}},
		{"negative quantity", &models.StockMovement{ItemID: uuid.New(), Type: models.StockMovementIn, Quantity: -5}},
		{"unknown type", &models.StockMovement{ItemID: uuid.New(), Type: "BOGUS", Quantity: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := service.NewStockService(&fakeStockItemRepo{}, &fakeStockMovementRepo{})
			_, err := svc.RecordMovement(context.Background(), uuid.New(), tc.movement)
			require.ErrorIs(t, err, service.ErrValidation)
		})
	}
}

func TestStockService_RecordMovement_DelegatesToRepository(t *testing.T) {
	itemID := uuid.New()
	called := false
	movements := &fakeStockMovementRepo{
		recordMovementFn: func(_ context.Context, _ uuid.UUID, m *models.StockMovement) (*models.StockItem, error) {
			called = true
			assert.Equal(t, itemID, m.ItemID)
			return &models.StockItem{QuantityOnHand: 5}, nil
		},
	}
	svc := service.NewStockService(&fakeStockItemRepo{}, movements)

	item, err := svc.RecordMovement(context.Background(), uuid.New(), &models.StockMovement{
		ItemID:   itemID,
		Type:     models.StockMovementOut,
		Quantity: 3,
	})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, 5, item.QuantityOnHand)
}

func intPtr(n int) *int { return &n }
