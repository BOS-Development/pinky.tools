package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

// createRunForPnl is a shared helper to create a user and hauling run for P&L tests.
func createRunForPnl(t *testing.T, userID int64, userName string) (*models.HaulingRun, *repositories.HaulingRunPnl) {
	t.Helper()
	db, err := setupDatabase(t)
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: userName})
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "PnL Test Run",
		Status:       "SELLING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	pnlRepo := repositories.NewHaulingRunPnl(db)
	return run, pnlRepo
}

func Test_HaulingRunPnl_UpsertPnlEntry_Create(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9200), "PnL User 1")

	avgSell := 1500.0
	totalRevenue := 150000.0
	totalCost := 100000.0

	entry := &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(34),
		QuantitySold:    int64(100),
		AvgSellPriceISK: &avgSell,
		TotalRevenueISK: &totalRevenue,
		TotalCostISK:    &totalCost,
	}

	err := pnlRepo.UpsertPnlEntry(context.Background(), entry)
	assert.NoError(t, err)
	assert.NotZero(t, entry.ID)
	assert.NotEmpty(t, entry.CreatedAt)
	assert.NotEmpty(t, entry.UpdatedAt)
}

func Test_HaulingRunPnl_UpsertPnlEntry_Update(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9210), "PnL User 2")

	avgSell := 1500.0
	totalRevenue := 150000.0
	totalCost := 100000.0

	entry := &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(35),
		QuantitySold:    int64(100),
		AvgSellPriceISK: &avgSell,
		TotalRevenueISK: &totalRevenue,
		TotalCostISK:    &totalCost,
	}

	// First insert
	err := pnlRepo.UpsertPnlEntry(context.Background(), entry)
	assert.NoError(t, err)
	firstID := entry.ID

	// Update with different values
	newAvgSell := 1600.0
	newTotalRevenue := 160000.0
	updatedEntry := &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(35),
		QuantitySold:    int64(120),
		AvgSellPriceISK: &newAvgSell,
		TotalRevenueISK: &newTotalRevenue,
		TotalCostISK:    &totalCost,
	}

	err = pnlRepo.UpsertPnlEntry(context.Background(), updatedEntry)
	assert.NoError(t, err)
	// ON CONFLICT UPDATE returns same row ID
	assert.Equal(t, firstID, updatedEntry.ID)

	// Verify updated values
	entries, err := pnlRepo.GetPnlByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(120), entries[0].QuantitySold)
	assert.NotNil(t, entries[0].AvgSellPriceISK)
	assert.Equal(t, 1600.0, *entries[0].AvgSellPriceISK)
}

func Test_HaulingRunPnl_UpsertPnlEntry_NullCosts(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9220), "PnL User 3")

	entry := &models.HaulingRunPnlEntry{
		RunID:        run.ID,
		TypeID:       int64(36),
		QuantitySold: int64(50),
		// nil prices
	}

	err := pnlRepo.UpsertPnlEntry(context.Background(), entry)
	assert.NoError(t, err)
	assert.NotZero(t, entry.ID)

	entries, err := pnlRepo.GetPnlByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Nil(t, entries[0].AvgSellPriceISK)
	assert.Nil(t, entries[0].TotalRevenueISK)
	assert.Nil(t, entries[0].TotalCostISK)
	assert.Nil(t, entries[0].NetProfitISK)
}

func Test_HaulingRunPnl_GetPnlByRunID_Empty(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9230), "PnL User 4")

	entries, err := pnlRepo.GetPnlByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 0)
}

func Test_HaulingRunPnl_GetPnlByRunID_WithTypeName(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9240)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "PnL User 5"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "TypeName PnL Run",
		Status:       "SELLING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	// Add a run item with a type name (so the JOIN works)
	itemsRepo := repositories.NewHaulingRunItems(db)
	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run.ID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
	})
	assert.NoError(t, err)

	pnlRepo := repositories.NewHaulingRunPnl(db)
	totalRevenue := 150000.0
	totalCost := 100000.0
	entry := &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(34),
		QuantitySold:    int64(100),
		TotalRevenueISK: &totalRevenue,
		TotalCostISK:    &totalCost,
	}
	err = pnlRepo.UpsertPnlEntry(context.Background(), entry)
	assert.NoError(t, err)

	entries, err := pnlRepo.GetPnlByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Tritanium", entries[0].TypeName)
	// net_profit should be revenue - cost = 50000
	assert.NotNil(t, entries[0].NetProfitISK)
	assert.Equal(t, 50000.0, *entries[0].NetProfitISK)
}

func Test_HaulingRunPnl_GetPnlSummaryByRunID_Empty(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9250), "PnL User 6")

	summary, err := pnlRepo.GetPnlSummaryByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 0.0, summary.TotalRevenueISK)
	assert.Equal(t, 0.0, summary.TotalCostISK)
	assert.Equal(t, 0.0, summary.NetProfitISK)
	assert.Equal(t, 0.0, summary.MarginPct)
	assert.Equal(t, int64(0), summary.ItemsSold)
	assert.Equal(t, int64(0), summary.ItemsPending)
}

func Test_HaulingRunPnl_GetPnlSummaryByRunID_WithData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9260)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "PnL User 7"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	run, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Summary PnL Run",
		Status:       "SELLING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	// Add run items (one sold, one pending)
	itemsRepo := repositories.NewHaulingRunItems(db)
	soldItem, err := itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:            run.ID,
		TypeID:           int64(34),
		TypeName:         "Tritanium",
		QuantityPlanned:  int64(100),
		QuantityAcquired: int64(100),
	})
	assert.NoError(t, err)
	_ = soldItem

	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:            run.ID,
		TypeID:           int64(35),
		TypeName:         "Pyerite",
		QuantityPlanned:  int64(50),
		QuantityAcquired: int64(50),
	})
	assert.NoError(t, err)

	pnlRepo := repositories.NewHaulingRunPnl(db)

	// Only sell Tritanium, leaving Pyerite as pending
	totalRevenue := 150000.0
	totalCost := 100000.0
	err = pnlRepo.UpsertPnlEntry(context.Background(), &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(34),
		QuantitySold:    int64(100),
		TotalRevenueISK: &totalRevenue,
		TotalCostISK:    &totalCost,
	})
	assert.NoError(t, err)

	summary, err := pnlRepo.GetPnlSummaryByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 150000.0, summary.TotalRevenueISK)
	assert.Equal(t, 100000.0, summary.TotalCostISK)
	assert.Equal(t, 50000.0, summary.NetProfitISK)
	// Margin = 50000 / 150000 * 100 ≈ 33.33%
	assert.InDelta(t, 33.33, summary.MarginPct, 0.1)
	assert.Equal(t, int64(1), summary.ItemsSold)
	// Pyerite is pending (quantity_sold=0 < quantity_acquired=50)
	assert.Equal(t, int64(1), summary.ItemsPending)
}

func Test_HaulingRunPnl_MultipleEntries(t *testing.T) {
	run, pnlRepo := createRunForPnl(t, int64(9270), "PnL User 8")

	revenue1 := 50000.0
	cost1 := 40000.0
	revenue2 := 80000.0
	cost2 := 60000.0

	err := pnlRepo.UpsertPnlEntry(context.Background(), &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(37),
		QuantitySold:    int64(10),
		TotalRevenueISK: &revenue1,
		TotalCostISK:    &cost1,
	})
	assert.NoError(t, err)

	err = pnlRepo.UpsertPnlEntry(context.Background(), &models.HaulingRunPnlEntry{
		RunID:           run.ID,
		TypeID:          int64(38),
		QuantitySold:    int64(20),
		TotalRevenueISK: &revenue2,
		TotalCostISK:    &cost2,
	})
	assert.NoError(t, err)

	entries, err := pnlRepo.GetPnlByRunID(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
}
