package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

// createAnalyticsTestData creates two completed runs with P&L data for analytics tests.
// Returns the db, the two runs, and a pnlRepo.
func createAnalyticsTestData(t *testing.T, userID int64, userName string) (*repositories.HaulingAnalytics, *repositories.HaulingRunPnl, *models.HaulingRun, *models.HaulingRun) {
	t.Helper()
	db, err := setupDatabase(t)
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	if err := userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: userName}); err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	runsRepo := repositories.NewHaulingRuns(db)
	maxVol := 350000.0

	// Run 1: Jita → Amarr, COMPLETE
	run1, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Analytics Run 1",
		Status:       "COMPLETE",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		MaxVolumeM3:  &maxVol,
	})
	if err != nil {
		t.Fatalf("failed to create run 1: %v", err)
	}
	// Set completed_at via UpdateRunStatus
	if err := runsRepo.UpdateRunStatus(context.Background(), run1.ID, userID, "COMPLETE"); err != nil {
		t.Fatalf("failed to set run1 status: %v", err)
	}

	// Run 2: Jita → Dodixie, COMPLETE
	run2, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Analytics Run 2",
		Status:       "COMPLETE",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000068),
		MaxVolumeM3:  &maxVol,
	})
	if err != nil {
		t.Fatalf("failed to create run 2: %v", err)
	}
	if err := runsRepo.UpdateRunStatus(context.Background(), run2.ID, userID, "COMPLETE"); err != nil {
		t.Fatalf("failed to set run2 status: %v", err)
	}

	itemsRepo := repositories.NewHaulingRunItems(db)
	pnlRepo := repositories.NewHaulingRunPnl(db)
	analyticsRepo := repositories.NewHaulingAnalytics(db)

	// Add run items for type name resolution
	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run1.ID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(1000),
	})
	if err != nil {
		t.Fatalf("failed to add item to run1: %v", err)
	}

	_, err = itemsRepo.AddItem(context.Background(), &models.HaulingRunItem{
		RunID:           run2.ID,
		TypeID:          int64(35),
		TypeName:        "Pyerite",
		QuantityPlanned: int64(500),
	})
	if err != nil {
		t.Fatalf("failed to add item to run2: %v", err)
	}

	// P&L for run1: revenue=200000, cost=100000, profit=100000
	rev1 := 200000.0
	cost1 := 100000.0
	if err := pnlRepo.UpsertPnlEntry(context.Background(), &models.HaulingRunPnlEntry{
		RunID:           run1.ID,
		TypeID:          int64(34),
		QuantitySold:    int64(100),
		TotalRevenueISK: &rev1,
		TotalCostISK:    &cost1,
	}); err != nil {
		t.Fatalf("failed to upsert pnl run1: %v", err)
	}

	// P&L for run2: revenue=80000, cost=60000, profit=20000
	rev2 := 80000.0
	cost2 := 60000.0
	if err := pnlRepo.UpsertPnlEntry(context.Background(), &models.HaulingRunPnlEntry{
		RunID:           run2.ID,
		TypeID:          int64(35),
		QuantitySold:    int64(50),
		TotalRevenueISK: &rev2,
		TotalCostISK:    &cost2,
	}); err != nil {
		t.Fatalf("failed to upsert pnl run2: %v", err)
	}

	return analyticsRepo, pnlRepo, run1, run2
}

// --- GetRouteAnalytics ---

func Test_HaulingAnalytics_GetRouteAnalytics_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9500)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Empty User"})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	results, err := analyticsRepo.GetRouteAnalytics(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingAnalytics_GetRouteAnalytics_WithData(t *testing.T) {
	analyticsRepo, _, run1, _ := createAnalyticsTestData(t, int64(9510), "Analytics Route User")

	results, err := analyticsRepo.GetRouteAnalytics(context.Background(), int64(9510))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Two different routes (run1: 10000002→10000043, run2: 10000002→10000068)
	assert.Len(t, results, 2)

	// First result ordered by total_profit DESC: run1 has profit=100000
	r := results[0]
	assert.Equal(t, run1.FromRegionID, r.FromRegionID)
	assert.Equal(t, run1.ToRegionID, r.ToRegionID)
	assert.Equal(t, int64(1), r.TotalRuns)
	assert.Equal(t, 100000.0, r.TotalProfitISK)
	assert.Equal(t, 100000.0, r.AvgProfitISK)
	// Margin = 100000/200000*100 = 50%
	assert.InDelta(t, 50.0, r.AvgMarginPct, 0.1)
	assert.Equal(t, 100000.0, r.BestRunProfitISK)
	assert.Equal(t, 100000.0, r.WorstRunProfitISK)
}

func Test_HaulingAnalytics_GetRouteAnalytics_ExcludesNonComplete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9520)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Non-Complete User"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)
	// Create a PLANNING run (not COMPLETE)
	_, err = runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Not Complete",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	results, err := analyticsRepo.GetRouteAnalytics(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

// --- GetItemAnalytics ---

func Test_HaulingAnalytics_GetItemAnalytics_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9530)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Item Empty"})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	results, err := analyticsRepo.GetItemAnalytics(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingAnalytics_GetItemAnalytics_WithData(t *testing.T) {
	analyticsRepo, _, _, _ := createAnalyticsTestData(t, int64(9540), "Analytics Item User")

	results, err := analyticsRepo.GetItemAnalytics(context.Background(), int64(9540))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	// Ordered by total_profit DESC: Tritanium (type 34) has profit=100000
	item := results[0]
	assert.Equal(t, int64(34), item.TypeID)
	assert.Equal(t, "Tritanium", item.TypeName)
	assert.Equal(t, int64(1), item.TotalRuns)
	assert.Equal(t, int64(100), item.TotalQtySold)
	assert.Equal(t, 100000.0, item.TotalProfitISK)
	// Margin = 100000/200000*100 = 50%
	assert.InDelta(t, 50.0, item.AvgMarginPct, 0.1)
}

// --- GetProfitTimeSeries ---

func Test_HaulingAnalytics_GetProfitTimeSeries_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9550)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics TimeSeries Empty"})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	results, err := analyticsRepo.GetProfitTimeSeries(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func Test_HaulingAnalytics_GetProfitTimeSeries_WithData(t *testing.T) {
	analyticsRepo, _, _, _ := createAnalyticsTestData(t, int64(9560), "Analytics TimeSeries User")

	results, err := analyticsRepo.GetProfitTimeSeries(context.Background(), int64(9560))
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Two routes completed today
	assert.GreaterOrEqual(t, len(results), 2)

	// All results should have a valid date format
	for _, dp := range results {
		assert.Len(t, dp.Date, 10) // YYYY-MM-DD
		assert.Greater(t, dp.RunCount, int64(0))
	}
}

// --- GetRunDurationSummary ---

func Test_HaulingAnalytics_GetRunDurationSummary_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9570)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Duration Empty"})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	summary, err := analyticsRepo.GetRunDurationSummary(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(0), summary.TotalCompletedRuns)
	assert.Equal(t, 0.0, summary.AvgDurationDays)
	assert.Equal(t, 0.0, summary.TotalProfitISK)
}

func Test_HaulingAnalytics_GetRunDurationSummary_WithData(t *testing.T) {
	analyticsRepo, _, _, _ := createAnalyticsTestData(t, int64(9580), "Analytics Duration User")

	summary, err := analyticsRepo.GetRunDurationSummary(context.Background(), int64(9580))
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(2), summary.TotalCompletedRuns)
	// Total profit: 100000 + 20000 = 120000
	assert.Equal(t, 120000.0, summary.TotalProfitISK)
	// Duration is close to 0 since runs were just created and completed
	assert.GreaterOrEqual(t, summary.AvgDurationDays, 0.0)
	assert.GreaterOrEqual(t, summary.MinDurationDays, 0.0)
	assert.GreaterOrEqual(t, summary.MaxDurationDays, 0.0)
}

// --- GetCompletedRuns ---

func Test_HaulingAnalytics_GetCompletedRuns_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9590)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Completed Empty"})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	runs, total, err := analyticsRepo.GetCompletedRuns(context.Background(), userID, 20, 0)
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Equal(t, int64(0), total)
	assert.Len(t, runs, 0)
}

func Test_HaulingAnalytics_GetCompletedRuns_WithData(t *testing.T) {
	analyticsRepo, _, _, _ := createAnalyticsTestData(t, int64(9600), "Analytics Completed User")

	runs, total, err := analyticsRepo.GetCompletedRuns(context.Background(), int64(9600), 20, 0)
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Equal(t, int64(2), total)
	assert.Len(t, runs, 2)

	// All returned runs should be COMPLETE
	for _, r := range runs {
		assert.Equal(t, "COMPLETE", r.Status)
		assert.NotNil(t, r.CompletedAt)
		assert.NotNil(t, r.Items)
	}
}

func Test_HaulingAnalytics_GetCompletedRuns_Pagination(t *testing.T) {
	analyticsRepo, _, _, _ := createAnalyticsTestData(t, int64(9610), "Analytics Pagination User")

	// Limit 1 — should return 1 result but total=2
	runs, total, err := analyticsRepo.GetCompletedRuns(context.Background(), int64(9610), 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, runs, 1)

	// Offset 2 — should return 0 results
	runs2, total2, err := analyticsRepo.GetCompletedRuns(context.Background(), int64(9610), 20, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, runs2, 0)
}

func Test_HaulingAnalytics_GetCompletedRuns_IncludesCancelled(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userID := int64(9620)
	userRepo := repositories.NewUserRepository(db)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Analytics Cancelled User"})
	assert.NoError(t, err)

	runsRepo := repositories.NewHaulingRuns(db)

	// Create a CANCELLED run
	run1, err := runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Cancelled Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)
	err = runsRepo.UpdateRunStatus(context.Background(), run1.ID, userID, "CANCELLED")
	assert.NoError(t, err)

	// Create a PLANNING run (should not appear)
	_, err = runsRepo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Planning Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	analyticsRepo := repositories.NewHaulingAnalytics(db)
	runs, total, err := analyticsRepo.GetCompletedRuns(context.Background(), userID, 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, runs, 1)
	assert.Equal(t, "CANCELLED", runs[0].Status)
}
