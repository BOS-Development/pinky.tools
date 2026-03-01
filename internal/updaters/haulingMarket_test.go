package updaters_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// Mock ESI client for hauling market

type mockHaulingMarketEsiClient struct {
	orders  []*client.MarketOrder
	ordersErr error
	history []*client.MarketHistoryEntry
	historyErr error
}

func (m *mockHaulingMarketEsiClient) GetMarketOrdersFiltered(ctx context.Context, regionID int64, systemID int64) ([]*client.MarketOrder, error) {
	return m.orders, m.ordersErr
}

func (m *mockHaulingMarketEsiClient) GetMarketHistory(ctx context.Context, regionID int64, typeID int64) ([]*client.MarketHistoryEntry, error) {
	return m.history, m.historyErr
}

// Mock repository for hauling market

type mockHaulingMarketRepo struct {
	snapshotAge    *time.Time
	snapshotAgeErr error
	upsertErr      error
	upsertedSnaps  []*models.HaulingMarketSnapshot
}

func (m *mockHaulingMarketRepo) UpsertSnapshots(ctx context.Context, snapshots []*models.HaulingMarketSnapshot) error {
	m.upsertedSnaps = append(m.upsertedSnaps, snapshots...)
	return m.upsertErr
}

func (m *mockHaulingMarketRepo) GetSnapshotAge(ctx context.Context, regionID int64, systemID int64) (*time.Time, error) {
	return m.snapshotAge, m.snapshotAgeErr
}

func Test_HaulingMarket_ScanRegion_FreshCacheSkip(t *testing.T) {
	// When snapshot is fresh, skip the scan
	recentTime := time.Now().Add(-5 * time.Minute)
	repo := &mockHaulingMarketRepo{snapshotAge: &recentTime}
	esi := &mockHaulingMarketEsiClient{}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.NoError(t, err)
	// No snapshots should have been upserted
	assert.Len(t, repo.upsertedSnaps, 0)
}

func Test_HaulingMarket_ScanRegion_StaleCacheScan(t *testing.T) {
	// When snapshot is stale (>30m), run the scan
	staleTime := time.Now().Add(-1 * time.Hour)
	repo := &mockHaulingMarketRepo{snapshotAge: &staleTime}
	esi := &mockHaulingMarketEsiClient{
		orders: []*client.MarketOrder{
			{TypeID: int64(34), Price: 1000.0, IsBuyOrder: false, VolumeRemain: int64(500)},
			{TypeID: int64(34), Price: 1200.0, IsBuyOrder: true, VolumeRemain: int64(200)},
			{TypeID: int64(35), Price: 500.0, IsBuyOrder: false, VolumeRemain: int64(100)},
		},
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.NoError(t, err)
	assert.NotEmpty(t, repo.upsertedSnaps)
	assert.Len(t, repo.upsertedSnaps, 2) // 2 types
}

func Test_HaulingMarket_ScanRegion_NoExistingSnapshot(t *testing.T) {
	// When no existing snapshot (nil age), always scan
	repo := &mockHaulingMarketRepo{snapshotAge: nil}
	esi := &mockHaulingMarketEsiClient{
		orders: []*client.MarketOrder{
			{TypeID: int64(34), Price: 1000.0, IsBuyOrder: false, VolumeRemain: int64(500)},
		},
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 1)
}

func Test_HaulingMarket_ScanRegion_ESIError(t *testing.T) {
	repo := &mockHaulingMarketRepo{snapshotAge: nil}
	esi := &mockHaulingMarketEsiClient{
		ordersErr: fmt.Errorf("ESI timeout"),
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch market orders")
}

func Test_HaulingMarket_ScanRegion_SnapshotAgeError(t *testing.T) {
	repo := &mockHaulingMarketRepo{snapshotAgeErr: fmt.Errorf("db error")}
	esi := &mockHaulingMarketEsiClient{}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check snapshot age")
}

func Test_HaulingMarket_ScanRegion_UpsertError(t *testing.T) {
	repo := &mockHaulingMarketRepo{
		snapshotAge: nil,
		upsertErr:   fmt.Errorf("db write failed"),
	}
	esi := &mockHaulingMarketEsiClient{
		orders: []*client.MarketOrder{
			{TypeID: int64(34), Price: 1000.0, IsBuyOrder: false, VolumeRemain: int64(500)},
		},
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert snapshots")
}

func Test_HaulingMarket_ScanRegion_EmptyOrders(t *testing.T) {
	repo := &mockHaulingMarketRepo{snapshotAge: nil}
	esi := &mockHaulingMarketEsiClient{orders: []*client.MarketOrder{}}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.NoError(t, err)
	// With no orders, no snapshots upserted
	assert.Len(t, repo.upsertedSnaps, 0)
}

func Test_HaulingMarket_ScanRegion_BestPriceSelection(t *testing.T) {
	// Verify that min sell price and max buy price are correctly selected
	repo := &mockHaulingMarketRepo{snapshotAge: nil}
	esi := &mockHaulingMarketEsiClient{
		orders: []*client.MarketOrder{
			{TypeID: int64(34), Price: 1200.0, IsBuyOrder: false, VolumeRemain: int64(100)},
			{TypeID: int64(34), Price: 1000.0, IsBuyOrder: false, VolumeRemain: int64(200)},
			{TypeID: int64(34), Price: 800.0, IsBuyOrder: true, VolumeRemain: int64(50)},
			{TypeID: int64(34), Price: 900.0, IsBuyOrder: true, VolumeRemain: int64(75)},
		},
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanRegion(context.Background(), int64(10000002), int64(0))

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 1)

	snap := repo.upsertedSnaps[0]
	// Min sell price should be 1000
	assert.NotNil(t, snap.SellPrice)
	assert.Equal(t, 1000.0, *snap.SellPrice)
	// Max buy price should be 900
	assert.NotNil(t, snap.BuyPrice)
	assert.Equal(t, 900.0, *snap.BuyPrice)
	// Total sell volume = 100 + 200 = 300
	assert.NotNil(t, snap.VolumeAvailable)
	assert.Equal(t, int64(300), *snap.VolumeAvailable)
}

func Test_HaulingMarket_ScanForHistory_Success(t *testing.T) {
	repo := &mockHaulingMarketRepo{}
	esi := &mockHaulingMarketEsiClient{
		history: []*client.MarketHistoryEntry{
			{Date: "2026-02-01", Volume: int64(1000)},
			{Date: "2026-02-02", Volume: int64(2000)},
			{Date: "2026-02-03", Volume: int64(1500)},
		},
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanForHistory(context.Background(), int64(10000002), []int64{34, 35})

	assert.NoError(t, err)
	// Should have upserted 2 snapshots (one per type)
	assert.Len(t, repo.upsertedSnaps, 2)
	// avg daily volume = (1000 + 2000 + 1500) / 3 = 1500
	for _, s := range repo.upsertedSnaps {
		assert.NotNil(t, s.AvgDailyVolume)
		assert.Equal(t, 1500.0, *s.AvgDailyVolume)
	}
}

func Test_HaulingMarket_ScanForHistory_ESIError_Continues(t *testing.T) {
	// If ESI fails for one type, continue with others
	repo := &mockHaulingMarketRepo{}
	esi := &mockHaulingMarketEsiClient{
		historyErr: fmt.Errorf("ESI error"),
	}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanForHistory(context.Background(), int64(10000002), []int64{34})

	// ScanForHistory logs error but continues, returning nil
	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 0)
}

func Test_HaulingMarket_ScanForHistory_EmptyHistory(t *testing.T) {
	repo := &mockHaulingMarketRepo{}
	esi := &mockHaulingMarketEsiClient{history: []*client.MarketHistoryEntry{}}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanForHistory(context.Background(), int64(10000002), []int64{34})

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 0)
}

func Test_HaulingMarket_ScanForHistory_Over30Days(t *testing.T) {
	// Should only use last 30 days
	repo := &mockHaulingMarketRepo{}
	history := make([]*client.MarketHistoryEntry, 40)
	for i := 0; i < 40; i++ {
		history[i] = &client.MarketHistoryEntry{Volume: int64(100)}
	}
	esi := &mockHaulingMarketEsiClient{history: history}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanForHistory(context.Background(), int64(10000002), []int64{34})

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 1)
	// avg daily = 100 (all same)
	assert.Equal(t, 100.0, *repo.upsertedSnaps[0].AvgDailyVolume)
}

func Test_HaulingMarket_ScanForHistory_EmptyTypeIDs(t *testing.T) {
	repo := &mockHaulingMarketRepo{}
	esi := &mockHaulingMarketEsiClient{}

	u := updaters.NewHaulingMarket(repo, esi)
	err := u.ScanForHistory(context.Background(), int64(10000002), []int64{})

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedSnaps, 0)
}
