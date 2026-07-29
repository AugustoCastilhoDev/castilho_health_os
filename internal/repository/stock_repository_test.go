package repository_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/castilho/health-os/internal/domain/models"
	"github.com/castilho/health-os/internal/repository"
	"github.com/castilho/health-os/internal/testutil"
)

func TestStockRepository_RecordMovement_UpdatesBalanceAndLogs(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	admin := testutil.NewUser(t, gdb, tenant.ID, models.RoleTenantAdmin)
	items := repository.NewStockItemRepository(gdb)
	movements := repository.NewStockMovementRepository(gdb)
	ctx := context.Background()

	item := &models.StockItem{Name: "Luva Cirúrgica", Unit: "par"}
	require.NoError(t, items.Create(ctx, tenant.ID, item))

	updated, err := movements.RecordMovement(ctx, tenant.ID, &models.StockMovement{
		ItemID:      item.ID,
		Type:        models.StockMovementIn,
		Quantity:    50,
		CreatedByID: admin.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, 50, updated.QuantityOnHand)

	updated, err = movements.RecordMovement(ctx, tenant.ID, &models.StockMovement{
		ItemID:      item.ID,
		Type:        models.StockMovementOut,
		Quantity:    20,
		CreatedByID: admin.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, 30, updated.QuantityOnHand)

	found, err := items.FindByID(ctx, tenant.ID, item.ID)
	require.NoError(t, err)
	assert.Equal(t, 30, found.QuantityOnHand)

	log, err := movements.ListByItem(ctx, tenant.ID, item.ID)
	require.NoError(t, err)
	require.Len(t, log, 2)
}

func TestStockRepository_RecordMovement_RejectsOutBeyondBalance(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	admin := testutil.NewUser(t, gdb, tenant.ID, models.RoleTenantAdmin)
	items := repository.NewStockItemRepository(gdb)
	movements := repository.NewStockMovementRepository(gdb)
	ctx := context.Background()

	item := &models.StockItem{Name: "Anestésico", Unit: "frasco"}
	require.NoError(t, items.Create(ctx, tenant.ID, item))

	_, err := movements.RecordMovement(ctx, tenant.ID, &models.StockMovement{
		ItemID:      item.ID,
		Type:        models.StockMovementOut,
		Quantity:    1,
		CreatedByID: admin.ID,
	})
	require.ErrorIs(t, err, repository.ErrInsufficientStock)

	found, err := items.FindByID(ctx, tenant.ID, item.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, found.QuantityOnHand, "a rejected OUT movement must not leave any partial effect")
}

// Regression guard for the row-lock in RecordMovement: without
// `clause.Locking{Strength: "UPDATE"}`, two concurrent OUT movements could
// both read the same starting balance and each subtract from it, losing an
// update. Firing them concurrently and asserting the final balance is exact
// is the only way to catch that class of bug — a sequential test would pass
// even with the lock removed.
func TestStockRepository_RecordMovement_SerializesConcurrentMovements(t *testing.T) {
	gdb := testutil.ConnectDB(t)
	tenant := testutil.NewTenant(t, gdb)
	admin := testutil.NewUser(t, gdb, tenant.ID, models.RoleTenantAdmin)
	items := repository.NewStockItemRepository(gdb)
	movements := repository.NewStockMovementRepository(gdb)
	ctx := context.Background()

	item := &models.StockItem{Name: "Máscara N95", Unit: "un"}
	require.NoError(t, items.Create(ctx, tenant.ID, item))
	_, err := movements.RecordMovement(ctx, tenant.ID, &models.StockMovement{
		ItemID:      item.ID,
		Type:        models.StockMovementIn,
		Quantity:    100,
		CreatedByID: admin.ID,
	})
	require.NoError(t, err)

	const concurrentMovements = 10
	var wg sync.WaitGroup
	wg.Add(concurrentMovements)
	for range concurrentMovements {
		go func() {
			defer wg.Done()
			_, err := movements.RecordMovement(ctx, tenant.ID, &models.StockMovement{
				ItemID:      item.ID,
				Type:        models.StockMovementOut,
				Quantity:    1,
				CreatedByID: admin.ID,
			})
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	found, err := items.FindByID(ctx, tenant.ID, item.ID)
	require.NoError(t, err)
	assert.Equal(t, 90, found.QuantityOnHand)
}
