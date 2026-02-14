package updaters_test

//go:generate mockgen -destination=marketPrices_mocks_test.go -package=updaters_test github.com/annymsMthd/industry-tool/internal/updaters MarketPricesRepository,MarketPricesEsiClient

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_MarketPricesUpdaterShouldUpdateJitaMarket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Mock market orders from ESI
	mockOrders := []*client.MarketOrder{
		{
			TypeID:       34,
			Price:        5.45,
			IsBuyOrder:   true,
			VolumeRemain: 10000,
		},
		{
			TypeID:       34,
			Price:        5.50,
			IsBuyOrder:   false,
			VolumeRemain: 8000,
		},
		{
			TypeID:       35,
			Price:        10.20,
			IsBuyOrder:   true,
			VolumeRemain: 5000,
		},
		{
			TypeID:       35,
			Price:        10.30,
			IsBuyOrder:   false,
			VolumeRemain: 3000,
		},
	}

	// Expect check for last update time (return nil = no previous update)
	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	// Expect ESI client call
	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	// Expect delete old prices
	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	// Expect upsert new prices
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prices []models.MarketPrice) error {
			// Verify we got prices for both types
			assert.Equal(t, 2, len(prices))

			// Find and verify type 34
			var price34 *models.MarketPrice
			for _, p := range prices {
				if p.TypeID == 34 {
					price34 = &p
					break
				}
			}
			assert.NotNil(t, price34)
			assert.NotNil(t, price34.BuyPrice)
			assert.NotNil(t, price34.SellPrice)
			assert.Equal(t, 5.45, *price34.BuyPrice)  // Best bid
			assert.Equal(t, 5.50, *price34.SellPrice) // Best ask
			assert.Equal(t, int64(18000), *price34.DailyVolume)

			return nil
		}).
		Times(1)

	// Create updater
	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	// Execute
	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_SkipsRecentUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Set last update to 1 hour ago (within 6 hour window)
	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		DoAndReturn(func(ctx context.Context, regionID int64) (*time.Time, error) {
			// Return a time within the last 6 hours
			recent := time.Now().Add(-1 * time.Hour)
			return &recent, nil
		}).
		Times(1)

	// Should NOT call ESI or update database
	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), gomock.Any()).
		Times(0)
	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), gomock.Any()).
		Times(0)
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		Times(0)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	// Execute - should skip update
	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_ESIClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// No previous update
	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	// ESI client returns error
	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(nil, assert.AnError).
		Times(1)

	// Should NOT try to delete or upsert
	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), gomock.Any()).
		Times(0)
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		Times(0)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch market orders from ESI")
}

func Test_MarketPricesUpdater_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	mockOrders := []*client.MarketOrder{
		{TypeID: 34, Price: 5.50, IsBuyOrder: false, VolumeRemain: 1000},
	}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	// Delete fails
	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(assert.AnError).
		Times(1)

	// Should NOT try to upsert
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		Times(0)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete old market prices")
}

func Test_MarketPricesUpdater_UpsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	mockOrders := []*client.MarketOrder{
		{TypeID: 34, Price: 5.50, IsBuyOrder: false, VolumeRemain: 1000},
	}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	// Upsert fails
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		Return(assert.AnError).
		Times(1)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert market prices")
}

func Test_MarketPricesUpdater_EmptyOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Empty orders
	mockOrders := []*client.MarketOrder{}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	// Should upsert empty slice
	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prices []models.MarketPrice) error {
			assert.Equal(t, 0, len(prices))
			return nil
		}).
		Times(1)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_OnlyBuyOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Only buy orders for type 34
	mockOrders := []*client.MarketOrder{
		{TypeID: 34, Price: 5.40, IsBuyOrder: true, VolumeRemain: 1000},
		{TypeID: 34, Price: 5.45, IsBuyOrder: true, VolumeRemain: 2000},
	}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prices []models.MarketPrice) error {
			assert.Equal(t, 1, len(prices))
			assert.Equal(t, int64(34), prices[0].TypeID)
			assert.NotNil(t, prices[0].BuyPrice)
			assert.Nil(t, prices[0].SellPrice) // No sell orders
			assert.Equal(t, 5.45, *prices[0].BuyPrice) // Highest buy order
			return nil
		}).
		Times(1)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_OnlySellOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Only sell orders for type 34
	mockOrders := []*client.MarketOrder{
		{TypeID: 34, Price: 5.60, IsBuyOrder: false, VolumeRemain: 1000},
		{TypeID: 34, Price: 5.50, IsBuyOrder: false, VolumeRemain: 2000},
	}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prices []models.MarketPrice) error {
			assert.Equal(t, 1, len(prices))
			assert.Equal(t, int64(34), prices[0].TypeID)
			assert.Nil(t, prices[0].BuyPrice) // No buy orders
			assert.NotNil(t, prices[0].SellPrice)
			assert.Equal(t, 5.50, *prices[0].SellPrice) // Lowest sell order
			return nil
		}).
		Times(1)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_MultiplePricesPicksBest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// Multiple buy and sell orders - should pick best of each
	mockOrders := []*client.MarketOrder{
		{TypeID: 34, Price: 5.40, IsBuyOrder: true, VolumeRemain: 1000},
		{TypeID: 34, Price: 5.45, IsBuyOrder: true, VolumeRemain: 2000},  // Best buy
		{TypeID: 34, Price: 5.42, IsBuyOrder: true, VolumeRemain: 1500},
		{TypeID: 34, Price: 5.55, IsBuyOrder: false, VolumeRemain: 800},
		{TypeID: 34, Price: 5.50, IsBuyOrder: false, VolumeRemain: 1200}, // Best sell
		{TypeID: 34, Price: 5.60, IsBuyOrder: false, VolumeRemain: 900},
	}

	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, nil).
		Times(1)

	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), int64(10000002)).
		Return(mockOrders, nil).
		Times(1)

	mockRepo.EXPECT().
		DeleteAllForRegion(gomock.Any(), int64(10000002)).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		UpsertPrices(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prices []models.MarketPrice) error {
			assert.Equal(t, 1, len(prices))
			assert.Equal(t, int64(34), prices[0].TypeID)
			assert.NotNil(t, prices[0].BuyPrice)
			assert.NotNil(t, prices[0].SellPrice)
			assert.Equal(t, 5.45, *prices[0].BuyPrice)  // Highest buy
			assert.Equal(t, 5.50, *prices[0].SellPrice) // Lowest sell
			// Total volume should be sum of all orders
			assert.Equal(t, int64(7400), *prices[0].DailyVolume)
			return nil
		}).
		Times(1)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.NoError(t, err)
}

func Test_MarketPricesUpdater_GetLastUpdateTimeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockMarketPricesRepository(ctrl)
	mockESIClient := NewMockMarketPricesEsiClient(ctrl)

	// GetLastUpdateTime fails
	mockRepo.EXPECT().
		GetLastUpdateTime(gomock.Any(), int64(10000002)).
		Return(nil, assert.AnError).
		Times(1)

	// Should NOT proceed
	mockESIClient.EXPECT().
		GetMarketOrders(gomock.Any(), gomock.Any()).
		Times(0)

	updater := updaters.NewMarketPrices(mockRepo, mockESIClient)

	err := updater.UpdateJitaMarket(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get last market price update time")
}
