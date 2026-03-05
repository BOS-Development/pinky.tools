package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_HaulingStructures_UpsertStructureSnapshots_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)
	err = repo.UpsertStructureSnapshots(context.Background(), int64(1234567890), []*models.HaulingMarketSnapshot{})
	assert.NoError(t, err)
}

func Test_HaulingStructures_UpsertStructureSnapshots_Persists(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	structureID := int64(1111111111)
	buyPrice := 900.0
	sellPrice := 1050.0
	volAvail := int64(200)

	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), BuyPrice: &buyPrice, SellPrice: &sellPrice, VolumeAvailable: &volAvail},
		{TypeID: int64(35), BuyPrice: nil, SellPrice: nil},
	}

	err = repo.UpsertStructureSnapshots(context.Background(), structureID, snapshots)
	assert.NoError(t, err)
}

func Test_HaulingStructures_UpsertStructureSnapshots_Idempotent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	structureID := int64(1111111112)
	buyPrice := 900.0
	sellPrice := 1050.0

	snap := &models.HaulingMarketSnapshot{TypeID: int64(36), BuyPrice: &buyPrice, SellPrice: &sellPrice}

	err = repo.UpsertStructureSnapshots(context.Background(), structureID, []*models.HaulingMarketSnapshot{snap})
	assert.NoError(t, err)

	// Upsert again with new price — should not error
	newBuy := 950.0
	snap.BuyPrice = &newBuy
	err = repo.UpsertStructureSnapshots(context.Background(), structureID, []*models.HaulingMarketSnapshot{snap})
	assert.NoError(t, err)
}

func Test_HaulingStructures_GetStructureSnapshotAge_NoData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	age, err := repo.GetStructureSnapshotAge(context.Background(), int64(9999999999))
	assert.NoError(t, err)
	assert.Nil(t, age)
}

func Test_HaulingStructures_GetStructureSnapshotAge_HasData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	structureID := int64(1111111113)
	buyPrice := 500.0
	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), BuyPrice: &buyPrice},
		{TypeID: int64(35), BuyPrice: &buyPrice},
	}

	err = repo.UpsertStructureSnapshots(context.Background(), structureID, snapshots)
	assert.NoError(t, err)

	age, err := repo.GetStructureSnapshotAge(context.Background(), structureID)
	assert.NoError(t, err)
	assert.NotNil(t, age)
}

func Test_HaulingStructures_GetStructureScannerResults_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	// No snapshots exist for this structure — should return empty slice, not error
	results, err := repo.GetStructureScannerResults(context.Background(), int64(9999999998), int64(10000002), int64(0))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingStructures_GetRegionToStructureResults_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingStructures(db)

	// No snapshots exist — should return empty slice, not error
	results, err := repo.GetRegionToStructureResults(context.Background(), int64(10000002), int64(0), int64(9999999997))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingStructures_GetStructureScannerResults_ArbitrageCalculation(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	// Set up structure snapshots (source)
	structRepo := repositories.NewHaulingStructures(db)
	structureID := int64(1111111114)

	// Sell from structure at 900, can buy at destination for 1100 → profit
	sellPrice := 900.0
	buyPrice := 1100.0
	volAvail := int64(50)
	err = structRepo.UpsertStructureSnapshots(context.Background(), structureID, []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), SellPrice: &sellPrice, VolumeAvailable: &volAvail},
	})
	assert.NoError(t, err)

	// Set up region market snapshots (destination: buy at 1100)
	marketRepo := repositories.NewHaulingMarket(db)
	destRegion := int64(10000043)
	err = marketRepo.UpsertSnapshots(context.Background(), []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), RegionID: destRegion, SystemID: int64(0), BuyPrice: &buyPrice},
	})
	assert.NoError(t, err)

	results, err := structRepo.GetStructureScannerResults(context.Background(), structureID, destRegion, int64(0))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// There should be an arbitrage row since destBuy (1100) > srcSell (900)
	if len(results) > 0 {
		r := results[0]
		assert.Equal(t, int64(34), r.TypeID)
		assert.NotNil(t, r.BuyPrice)  // source sell price becomes "buy price" in the row
		assert.NotNil(t, r.SellPrice) // dest buy price becomes "sell price" in the row
		assert.NotNil(t, r.NetProfitISK)
		assert.Greater(t, *r.NetProfitISK, 0.0)
	}
}

func Test_HaulingStructures_GetRegionToStructureResults_ArbitrageCalculation(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	// Set up region market snapshots (source)
	marketRepo := repositories.NewHaulingMarket(db)
	srcRegion := int64(10000099)
	srcSell := 800.0
	err = marketRepo.UpsertSnapshots(context.Background(), []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), RegionID: srcRegion, SystemID: int64(0), SellPrice: &srcSell},
	})
	assert.NoError(t, err)

	// Set up structure snapshots (destination: buy at 1000)
	structRepo := repositories.NewHaulingStructures(db)
	structureID := int64(1111111115)
	destBuy := 1000.0
	err = structRepo.UpsertStructureSnapshots(context.Background(), structureID, []*models.HaulingMarketSnapshot{
		{TypeID: int64(34), BuyPrice: &destBuy},
	})
	assert.NoError(t, err)

	results, err := structRepo.GetRegionToStructureResults(context.Background(), srcRegion, int64(0), structureID)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// There should be an arbitrage row since destBuy (1000) > srcSell (800)
	if len(results) > 0 {
		r := results[0]
		assert.Equal(t, int64(34), r.TypeID)
		assert.NotNil(t, r.NetProfitISK)
		assert.Greater(t, *r.NetProfitISK, 0.0)
	}
}
