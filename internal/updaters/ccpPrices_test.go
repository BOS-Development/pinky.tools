package updaters_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// Mock ESI client for CCP prices

type mockCcpPricesEsiClient struct {
	prices []*client.CcpMarketPrice
	err    error
}

func (m *mockCcpPricesEsiClient) GetCcpMarketPrices(ctx context.Context) ([]*client.CcpMarketPrice, error) {
	return m.prices, m.err
}

// Mock market repo for CCP prices

type mockCcpPricesMarketRepo struct {
	lastUpdateTime *time.Time
	lastUpdateErr  error
	upsertErr      error
	upsertedPrices map[int64]float64
}

func (m *mockCcpPricesMarketRepo) GetAdjustedPriceLastUpdateTime(ctx context.Context) (*time.Time, error) {
	return m.lastUpdateTime, m.lastUpdateErr
}

func (m *mockCcpPricesMarketRepo) UpsertAdjustedPrices(ctx context.Context, prices map[int64]float64) error {
	m.upsertedPrices = prices
	return m.upsertErr
}

func ptrFloat64(f float64) *float64 { return &f }

func Test_CcpPrices_FullUpdate(t *testing.T) {
	esiClient := &mockCcpPricesEsiClient{
		prices: []*client.CcpMarketPrice{
			{TypeID: 34, AdjustedPrice: ptrFloat64(10.5)},
			{TypeID: 35, AdjustedPrice: ptrFloat64(20.0)},
		},
	}
	repo := &mockCcpPricesMarketRepo{}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, map[int64]float64{34: 10.5, 35: 20.0}, repo.upsertedPrices)
}

func Test_CcpPrices_SkipsRecentUpdate(t *testing.T) {
	recentTime := time.Now().Add(-10 * time.Minute)
	repo := &mockCcpPricesMarketRepo{
		lastUpdateTime: &recentTime,
	}
	esiClient := &mockCcpPricesEsiClient{}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	// No upsert should have been called
	assert.Nil(t, repo.upsertedPrices)
}

func Test_CcpPrices_UpdatesWhenStale(t *testing.T) {
	staleTime := time.Now().Add(-2 * time.Hour)
	esiClient := &mockCcpPricesEsiClient{
		prices: []*client.CcpMarketPrice{
			{TypeID: 34, AdjustedPrice: ptrFloat64(10.5)},
		},
	}
	repo := &mockCcpPricesMarketRepo{
		lastUpdateTime: &staleTime,
	}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, map[int64]float64{34: 10.5}, repo.upsertedPrices)
}

func Test_CcpPrices_ErrorGettingLastUpdate(t *testing.T) {
	repo := &mockCcpPricesMarketRepo{
		lastUpdateErr: fmt.Errorf("db error"),
	}
	esiClient := &mockCcpPricesEsiClient{}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get last adjusted price update time")
}

func Test_CcpPrices_ErrorFetchingPrices(t *testing.T) {
	esiClient := &mockCcpPricesEsiClient{
		err: fmt.Errorf("network error"),
	}
	repo := &mockCcpPricesMarketRepo{}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch CCP market prices")
}

func Test_CcpPrices_ErrorUpserting(t *testing.T) {
	esiClient := &mockCcpPricesEsiClient{
		prices: []*client.CcpMarketPrice{
			{TypeID: 34, AdjustedPrice: ptrFloat64(10.5)},
		},
	}
	repo := &mockCcpPricesMarketRepo{
		upsertErr: fmt.Errorf("db error"),
	}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert adjusted prices")
}

func Test_CcpPrices_FiltersNilAdjustedPrices(t *testing.T) {
	esiClient := &mockCcpPricesEsiClient{
		prices: []*client.CcpMarketPrice{
			{TypeID: 34, AdjustedPrice: ptrFloat64(10.5)},
			{TypeID: 35, AdjustedPrice: nil},           // nil adjusted price
			{TypeID: 36, AdjustedPrice: ptrFloat64(0)},  // zero is valid
		},
	}
	repo := &mockCcpPricesMarketRepo{}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	// Only type 34 and 36 should be included (nil filtered out)
	assert.Equal(t, map[int64]float64{34: 10.5, 36: 0}, repo.upsertedPrices)
}

func Test_CcpPrices_FirstRunWithNoLastUpdate(t *testing.T) {
	esiClient := &mockCcpPricesEsiClient{
		prices: []*client.CcpMarketPrice{
			{TypeID: 34, AdjustedPrice: ptrFloat64(5.0)},
		},
	}
	repo := &mockCcpPricesMarketRepo{
		lastUpdateTime: nil, // No previous update
	}

	u := updaters.NewCcpPrices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, map[int64]float64{34: 5.0}, repo.upsertedPrices)
}
