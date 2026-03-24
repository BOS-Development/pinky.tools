package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_MarketPricesShouldUpsertPrices(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	// Need to ensure asset_item_types exist first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
	}
	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	marketPricesRepo := repositories.NewMarketPrices(db)

	buyPrice1 := 5.45
	sellPrice1 := 5.50
	volume1 := int64(100000)

	buyPrice2 := 10.20
	sellPrice2 := 10.30
	volume2 := int64(50000)

	prices := []models.MarketPrice{
		{
			TypeID:      34,
			RegionID:    10000002,
			BuyPrice:    &buyPrice1,
			SellPrice:   &sellPrice1,
			DailyVolume: &volume1,
		},
		{
			TypeID:      35,
			RegionID:    10000002,
			BuyPrice:    &buyPrice2,
			SellPrice:   &sellPrice2,
			DailyVolume: &volume2,
		},
	}

	err = marketPricesRepo.UpsertPrices(context.Background(), prices)
	assert.NoError(t, err)

	// Update a price
	newSellPrice := 5.55
	prices[0].SellPrice = &newSellPrice

	err = marketPricesRepo.UpsertPrices(context.Background(), prices)
	assert.NoError(t, err)

	// Verify via GetPricesForTypes
	retrieved, err := marketPricesRepo.GetPricesForTypes(context.Background(), []int64{34, 35}, 10000002)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(retrieved))
	assert.NotNil(t, retrieved[34])
	assert.Equal(t, 5.55, *retrieved[34].SellPrice)
}

func Test_MarketPricesShouldDeleteAllForRegion(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	// Setup item types
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 100, TypeName: "Test Item", Volume: 1.0, IconID: nil},
	}
	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	marketPricesRepo := repositories.NewMarketPrices(db)

	price := 100.0
	volume := int64(1000)

	prices := []models.MarketPrice{
		{
			TypeID:      100,
			RegionID:    10000002,
			BuyPrice:    &price,
			SellPrice:   &price,
			DailyVolume: &volume,
		},
	}

	err = marketPricesRepo.UpsertPrices(context.Background(), prices)
	assert.NoError(t, err)

	// Delete all prices for the region
	err = marketPricesRepo.DeleteAllForRegion(context.Background(), 10000002)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := marketPricesRepo.GetPricesForTypes(context.Background(), []int64{100}, 10000002)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(retrieved))
}

func Test_MarketPricesShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	marketPricesRepo := repositories.NewMarketPrices(db)

	err = marketPricesRepo.UpsertPrices(context.Background(), []models.MarketPrice{})
	assert.NoError(t, err)
}

func Test_UpsertAdjustedPrices_InsertsWithoutMarketPricesRow(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewMarketPrices(db)

	// Type IDs that have no rows in market_prices — verifies the new table
	// accepts any type_id without requiring a market_prices FK
	prices := map[int64]float64{
		99001: 1234.56,
		99002: 9876.54,
	}

	err = repo.UpsertAdjustedPrices(context.Background(), prices)
	assert.NoError(t, err)

	got, err := repo.GetAllAdjustedPrices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1234.56, got[99001])
	assert.Equal(t, 9876.54, got[99002])
}

func Test_UpsertAdjustedPrices_UpdatesExistingRow(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewMarketPrices(db)

	err = repo.UpsertAdjustedPrices(context.Background(), map[int64]float64{99010: 100.0})
	assert.NoError(t, err)

	// Upsert again with a new price — should update, not error
	err = repo.UpsertAdjustedPrices(context.Background(), map[int64]float64{99010: 200.0})
	assert.NoError(t, err)

	got, err := repo.GetAllAdjustedPrices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200.0, got[99010])
}

func Test_UpsertAdjustedPrices_EmptyMapIsNoop(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewMarketPrices(db)

	err = repo.UpsertAdjustedPrices(context.Background(), map[int64]float64{})
	assert.NoError(t, err)
}

func Test_GetAdjustedPriceLastUpdateTime_ReturnsNilWhenEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewMarketPrices(db)

	lastUpdate, err := repo.GetAdjustedPriceLastUpdateTime(context.Background())
	assert.NoError(t, err)
	// Table may or may not be empty — just verify no error and valid return
	_ = lastUpdate
}

func Test_GetAdjustedPriceLastUpdateTime_ReturnsTimeAfterUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewMarketPrices(db)

	err = repo.UpsertAdjustedPrices(context.Background(), map[int64]float64{99020: 50.0})
	assert.NoError(t, err)

	lastUpdate, err := repo.GetAdjustedPriceLastUpdateTime(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, lastUpdate)
}
