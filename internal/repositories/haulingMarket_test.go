package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_HaulingMarket_UpsertSnapshots_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)
	err = repo.UpsertSnapshots(context.Background(), []*models.HaulingMarketSnapshot{})
	assert.NoError(t, err)
}

func Test_HaulingMarket_UpsertSnapshots(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	buyPrice := 1000.0
	sellPrice := 1100.0
	volAvail := int64(500)

	snapshots := []*models.HaulingMarketSnapshot{
		{
			TypeID:          int64(34),
			RegionID:        int64(10000002),
			SystemID:        int64(0),
			BuyPrice:        &buyPrice,
			SellPrice:       &sellPrice,
			VolumeAvailable: &volAvail,
		},
		{
			TypeID:   int64(35),
			RegionID: int64(10000002),
			SystemID: int64(0),
		},
	}

	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)
}

func Test_HaulingMarket_UpsertSnapshots_Idempotent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	buyPrice := 1000.0
	sellPrice := 1100.0

	snap := &models.HaulingMarketSnapshot{
		TypeID:    int64(340),
		RegionID:  int64(10000099),
		SystemID:  int64(0),
		BuyPrice:  &buyPrice,
		SellPrice: &sellPrice,
	}

	err = repo.UpsertSnapshots(context.Background(), []*models.HaulingMarketSnapshot{snap})
	assert.NoError(t, err)

	// Upsert again with updated price
	newSell := 1200.0
	snap.SellPrice = &newSell
	err = repo.UpsertSnapshots(context.Background(), []*models.HaulingMarketSnapshot{snap})
	assert.NoError(t, err)
}

func Test_HaulingMarket_GetSnapshotAge_NoData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	age, err := repo.GetSnapshotAge(context.Background(), int64(99999999), int64(0))
	assert.NoError(t, err)
	assert.Nil(t, age)
}

func Test_HaulingMarket_GetSnapshotAge_WithData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	buyPrice := 500.0
	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(3400), RegionID: int64(10000077), SystemID: int64(0), BuyPrice: &buyPrice},
		{TypeID: int64(3500), RegionID: int64(10000077), SystemID: int64(0), BuyPrice: &buyPrice},
	}
	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)

	age, err := repo.GetSnapshotAge(context.Background(), int64(10000077), int64(0))
	assert.NoError(t, err)
	assert.NotNil(t, age)
	// Age should be recent (within 10 seconds of now)
	assert.WithinDuration(t, time.Now(), *age, 10*time.Second)
}

func Test_HaulingMarket_GetScannerResults_NoData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	results, err := repo.GetScannerResults(context.Background(), int64(10000002), int64(0), int64(10000043))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingMarket_GetScannerResults_WithArbitrageOpportunity(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	// Source region: sell price = 800 (cheap)
	sourceSell := 800.0
	sourceVol := int64(1000)

	// Dest region: buy price = 1200 (profitable)
	destBuy := 1200.0

	snapshots := []*models.HaulingMarketSnapshot{
		// Source region snapshot (we buy sell orders at source)
		{TypeID: int64(34000), RegionID: int64(10000055), SystemID: int64(0), SellPrice: &sourceSell, VolumeAvailable: &sourceVol},
		// Destination region snapshot (we sell to buy orders at dest)
		{TypeID: int64(34000), RegionID: int64(10000066), SystemID: int64(0), BuyPrice: &destBuy},
	}
	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)

	results, err := repo.GetScannerResults(context.Background(), int64(10000055), int64(0), int64(10000066))
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	row := results[0]
	assert.Equal(t, int64(34000), row.TypeID)
	assert.NotNil(t, row.BuyPrice)
	assert.Equal(t, 800.0, *row.BuyPrice)
	assert.NotNil(t, row.SellPrice)
	assert.Equal(t, 1200.0, *row.SellPrice)
	assert.NotNil(t, row.NetProfitISK)
	assert.Equal(t, 400.0, *row.NetProfitISK)
	assert.NotNil(t, row.Spread)
	// Spread = 1200/800 - 1 = 0.5
	assert.InDelta(t, 0.5, *row.Spread, 0.001)
	assert.Equal(t, "gap", row.Indicator) // >15%
}

func Test_HaulingMarket_GetScannerResults_MarkupIndicator(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	sourceSell := 1000.0
	destBuy := 1080.0 // 8% markup

	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(34001), RegionID: int64(10000088), SystemID: int64(0), SellPrice: &sourceSell},
		{TypeID: int64(34001), RegionID: int64(10000089), SystemID: int64(0), BuyPrice: &destBuy},
	}
	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)

	results, err := repo.GetScannerResults(context.Background(), int64(10000088), int64(0), int64(10000089))
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "markup", results[0].Indicator) // 5-15%
}

func Test_HaulingMarket_GetScannerResults_ThinIndicator(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	sourceSell := 1000.0
	destBuy := 1030.0 // 3% spread

	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(34002), RegionID: int64(10000091), SystemID: int64(0), SellPrice: &sourceSell},
		{TypeID: int64(34002), RegionID: int64(10000092), SystemID: int64(0), BuyPrice: &destBuy},
	}
	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)

	results, err := repo.GetScannerResults(context.Background(), int64(10000091), int64(0), int64(10000092))
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "thin", results[0].Indicator) // <5%
}

func Test_HaulingMarket_GetScannerResults_NoArbitrage(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingMarket(db)

	// Source sell price > dest buy price: no arbitrage
	sourceSell := 1500.0
	destBuy := 1000.0

	snapshots := []*models.HaulingMarketSnapshot{
		{TypeID: int64(34003), RegionID: int64(10000093), SystemID: int64(0), SellPrice: &sourceSell},
		{TypeID: int64(34003), RegionID: int64(10000094), SystemID: int64(0), BuyPrice: &destBuy},
	}
	err = repo.UpsertSnapshots(context.Background(), snapshots)
	assert.NoError(t, err)

	results, err := repo.GetScannerResults(context.Background(), int64(10000093), int64(0), int64(10000094))
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}
