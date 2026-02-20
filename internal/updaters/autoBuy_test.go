package updaters_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockAutoBuyConfigsRepo struct {
	byUserConfigs []*models.AutoBuyConfig
	byUserErr     error
	allActive     []*models.AutoBuyConfig
	allActiveErr  error
	deficits      []*models.StockpileDeficitItem
	deficitsErr   error
}

func (m *mockAutoBuyConfigsRepo) GetByUser(ctx context.Context, userID int64) ([]*models.AutoBuyConfig, error) {
	return m.byUserConfigs, m.byUserErr
}

func (m *mockAutoBuyConfigsRepo) GetAllActive(ctx context.Context) ([]*models.AutoBuyConfig, error) {
	return m.allActive, m.allActiveErr
}

func (m *mockAutoBuyConfigsRepo) GetStockpileDeficitsForConfig(ctx context.Context, config *models.AutoBuyConfig) ([]*models.StockpileDeficitItem, error) {
	return m.deficits, m.deficitsErr
}

type mockAutoBuyOrdersRepo struct {
	activeOrders       []*models.BuyOrder
	activeOrdersErr    error
	upsertErr          error
	deactivateAllErr   error
	deactivateOneErr   error
	upsertedOrders     []*models.BuyOrder
	deactivatedIDs     []int64
	deactivatedAllIDs  []int64
}

func (m *mockAutoBuyOrdersRepo) GetActiveAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) ([]*models.BuyOrder, error) {
	return m.activeOrders, m.activeOrdersErr
}

func (m *mockAutoBuyOrdersRepo) UpsertAutoBuy(ctx context.Context, order *models.BuyOrder) error {
	// Make a copy to capture the state at the time of the call
	copy := *order
	m.upsertedOrders = append(m.upsertedOrders, &copy)
	return m.upsertErr
}

func (m *mockAutoBuyOrdersRepo) DeactivateAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) error {
	m.deactivatedAllIDs = append(m.deactivatedAllIDs, autoBuyConfigID)
	return m.deactivateAllErr
}

func (m *mockAutoBuyOrdersRepo) DeactivateAutoBuyOrder(ctx context.Context, orderID int64) error {
	m.deactivatedIDs = append(m.deactivatedIDs, orderID)
	return m.deactivateOneErr
}

// --- Helper ---

func newAutoBuyUpdater(
	configRepo *mockAutoBuyConfigsRepo,
	buyOrderRepo *mockAutoBuyOrdersRepo,
	marketRepo *mockAutoSellMarketRepo,
) *updaters.AutoBuy {
	return updaters.NewAutoBuy(configRepo, buyOrderRepo, marketRepo)
}

// --- SyncForUser Tests ---

func Test_AutoBuy_SyncForUser_Success(t *testing.T) {
	configID := int64(1)
	buyPrice := 100.0

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 110.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 200, Deficit: 800},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, buyOrderRepo.upsertedOrders, 1)

	upserted := buyOrderRepo.upsertedOrders[0]
	assert.Equal(t, int64(42), upserted.BuyerUserID)
	assert.Equal(t, int64(34), upserted.TypeID)
	assert.Equal(t, int64(60003760), upserted.LocationID)
	assert.Equal(t, int64(800), upserted.QuantityDesired)
	assert.Equal(t, 110.0, upserted.MaxPricePerUnit) // 100 * 110 / 100
	assert.Equal(t, &configID, upserted.AutoBuyConfigID)
	assert.True(t, upserted.IsActive)
	assert.Len(t, buyOrderRepo.deactivatedIDs, 0)
}

func Test_AutoBuy_SyncForUser_NoConfigs(t *testing.T) {
	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{},
	}
	buyOrderRepo := &mockAutoBuyOrdersRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, buyOrderRepo.upsertedOrders, 0)
}

func Test_AutoBuy_SyncForUser_GetConfigsError(t *testing.T) {
	configRepo := &mockAutoBuyConfigsRepo{
		byUserErr: fmt.Errorf("db error"),
	}
	buyOrderRepo := &mockAutoBuyOrdersRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get auto-buy configs for user")
}

func Test_AutoBuy_SyncForUser_NoDeficit_DeactivatesExisting(t *testing.T) {
	configID := int64(1)
	existingOrderID := int64(99)

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 110.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 1000, Deficit: 0},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{
			{
				ID:              existingOrderID,
				TypeID:          34,
				AutoBuyConfigID: &configID,
				IsActive:        true,
			},
		},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// No upserts since deficit is 0
	assert.Len(t, buyOrderRepo.upsertedOrders, 0)
	// Existing order should be deactivated (not in activeTypes)
	assert.Contains(t, buyOrderRepo.deactivatedIDs, existingOrderID)
}

func Test_AutoBuy_SyncForUser_NoPrice_DeactivatesExisting(t *testing.T) {
	configID := int64(1)
	existingOrderID := int64(99)

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 110.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 200, Deficit: 800},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{
			{
				ID:              existingOrderID,
				TypeID:          34,
				AutoBuyConfigID: &configID,
				IsActive:        true,
			},
		},
	}

	// No prices available
	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// No upserts since there's no price
	assert.Len(t, buyOrderRepo.upsertedOrders, 0)
	// Deactivated from "no price" path and from "not in activeTypes" path
	assert.Contains(t, buyOrderRepo.deactivatedIDs, existingOrderID)
}

func Test_AutoBuy_SyncForUser_PerItemOverride(t *testing.T) {
	configID := int64(1)
	sellPrice := 200.0

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 110.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{
				TypeID:          34,
				DesiredQuantity: 1000,
				CurrentQuantity: 200,
				Deficit:         800,
				PriceSource:     stringPtr("jita_sell"),
				PricePercentage: float64Ptr(95.0),
			},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, SellPrice: &sellPrice},
		},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, buyOrderRepo.upsertedOrders, 1)

	upserted := buyOrderRepo.upsertedOrders[0]
	// Per-item override: jita_sell at 95%, not config's jita_buy at 110%
	assert.Equal(t, 190.0, upserted.MaxPricePerUnit) // 200 * 95 / 100
	assert.True(t, upserted.IsActive)
}

func Test_AutoBuy_SyncForUser_JitaSellPricing(t *testing.T) {
	configID := int64(1)
	sellPrice := 55.0

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 90.0,
				PriceSource:     "jita_sell",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 0, Deficit: 1000},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, SellPrice: &sellPrice},
		},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, buyOrderRepo.upsertedOrders, 1)

	upserted := buyOrderRepo.upsertedOrders[0]
	assert.Equal(t, 49.5, upserted.MaxPricePerUnit) // 55 * 90 / 100
	assert.True(t, upserted.IsActive)
}

func Test_AutoBuy_SyncForUser_JitaSplitPricing(t *testing.T) {
	configID := int64(1)
	buyPrice := 50.0
	sellPrice := 55.0

	configRepo := &mockAutoBuyConfigsRepo{
		byUserConfigs: []*models.AutoBuyConfig{
			{
				ID:              configID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 90.0,
				PriceSource:     "jita_split",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 0, Deficit: 1000},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice, SellPrice: &sellPrice},
		},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, buyOrderRepo.upsertedOrders, 1)

	upserted := buyOrderRepo.upsertedOrders[0]
	assert.Equal(t, 47.25, upserted.MaxPricePerUnit) // (50+55)/2 * 90 / 100 = 52.5 * 0.9
	assert.True(t, upserted.IsActive)
}

// --- SyncForAllUsers Tests ---

func Test_AutoBuy_SyncForAllUsers_Success(t *testing.T) {
	buyPrice := 100.0

	configRepo := &mockAutoBuyConfigsRepo{
		allActive: []*models.AutoBuyConfig{
			{
				ID:              1,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				MaxPricePercentage: 110.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
			{
				ID:              2,
				UserID:          99,
				OwnerType:       "character",
				OwnerID:         67890,
				LocationID:      60003760,
				MaxPricePercentage: 105.0,
				PriceSource:     "jita_buy",
				IsActive:        true,
			},
		},
		deficits: []*models.StockpileDeficitItem{
			{TypeID: 34, DesiredQuantity: 1000, CurrentQuantity: 200, Deficit: 800},
		},
	}

	buyOrderRepo := &mockAutoBuyOrdersRepo{
		activeOrders: []*models.BuyOrder{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.NoError(t, err)
	// Should have upserted for both configs
	assert.Len(t, buyOrderRepo.upsertedOrders, 2)
}

func Test_AutoBuy_SyncForAllUsers_NoConfigs(t *testing.T) {
	configRepo := &mockAutoBuyConfigsRepo{
		allActive: []*models.AutoBuyConfig{},
	}
	buyOrderRepo := &mockAutoBuyOrdersRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.NoError(t, err)
}

func Test_AutoBuy_SyncForAllUsers_Error(t *testing.T) {
	configRepo := &mockAutoBuyConfigsRepo{
		allActiveErr: fmt.Errorf("db error"),
	}
	buyOrderRepo := &mockAutoBuyOrdersRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoBuyUpdater(configRepo, buyOrderRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all active auto-buy configs")
}

// --- Constructor Test ---

func Test_AutoBuy_Constructor(t *testing.T) {
	u := updaters.NewAutoBuy(
		&mockAutoBuyConfigsRepo{},
		&mockAutoBuyOrdersRepo{},
		&mockAutoSellMarketRepo{},
	)
	assert.NotNil(t, u)
}

// --- Helper (local to this file, not redefining shared ones) ---

func stringPtr(v string) *string {
	return &v
}
