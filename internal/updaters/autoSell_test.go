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

type mockAutoSellContainersRepo struct {
	byUserContainers []*models.AutoSellContainer
	byUserErr        error
	allActive        []*models.AutoSellContainer
	allActiveErr     error
	containerItems   []*models.ContainerItem
	containerItemErr error
}

func (m *mockAutoSellContainersRepo) GetByUser(ctx context.Context, userID int64) ([]*models.AutoSellContainer, error) {
	return m.byUserContainers, m.byUserErr
}

func (m *mockAutoSellContainersRepo) GetAllActive(ctx context.Context) ([]*models.AutoSellContainer, error) {
	return m.allActive, m.allActiveErr
}

func (m *mockAutoSellContainersRepo) GetItemsInContainer(ctx context.Context, ownerType string, ownerID int64, containerID int64) ([]*models.ContainerItem, error) {
	return m.containerItems, m.containerItemErr
}

type mockAutoSellForSaleRepo struct {
	activeListings    []*models.ForSaleItem
	activeListingsErr error
	upsertErr         error
	deactivateErr     error
	upsertedItems     []*models.ForSaleItem
	deactivatedIDs    []int64
}

func (m *mockAutoSellForSaleRepo) GetActiveAutoSellListings(ctx context.Context, autoSellContainerID int64) ([]*models.ForSaleItem, error) {
	return m.activeListings, m.activeListingsErr
}

func (m *mockAutoSellForSaleRepo) Upsert(ctx context.Context, item *models.ForSaleItem) error {
	// Make a copy to capture the state at the time of the call
	copy := *item
	m.upsertedItems = append(m.upsertedItems, &copy)
	return m.upsertErr
}

func (m *mockAutoSellForSaleRepo) DeactivateAutoSellListings(ctx context.Context, autoSellContainerID int64) error {
	m.deactivatedIDs = append(m.deactivatedIDs, autoSellContainerID)
	return m.deactivateErr
}

type mockAutoSellMarketRepo struct {
	prices    map[int64]*models.MarketPrice
	pricesErr error
}

func (m *mockAutoSellMarketRepo) GetPricesForTypes(ctx context.Context, typeIDs []int64, regionID int64) (map[int64]*models.MarketPrice, error) {
	return m.prices, m.pricesErr
}

// --- Helper ---

func newAutoSellUpdater(
	autoSellRepo *mockAutoSellContainersRepo,
	forSaleRepo *mockAutoSellForSaleRepo,
	marketRepo *mockAutoSellMarketRepo,
) *updaters.AutoSell {
	return updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo)
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

// --- SyncForUser Tests ---

func Test_AutoSell_SyncForUser_Success(t *testing.T) {
	containerID := int64(1)
	buyPrice := 100.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 1)

	upserted := forSaleRepo.upsertedItems[0]
	assert.Equal(t, int64(42), upserted.UserID)
	assert.Equal(t, int64(34), upserted.TypeID)
	assert.Equal(t, "character", upserted.OwnerType)
	assert.Equal(t, int64(12345), upserted.OwnerID)
	assert.Equal(t, int64(60003760), upserted.LocationID)
	assert.Equal(t, int64(1000), upserted.QuantityAvailable)
	assert.Equal(t, 90.0, upserted.PricePerUnit) // 100 * 90 / 100
	assert.Equal(t, &containerID, upserted.AutoSellContainerID)
	assert.True(t, upserted.IsActive)
}

func Test_AutoSell_SyncForUser_NoContainers(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{},
	}
	forSaleRepo := &mockAutoSellForSaleRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 0)
}

func Test_AutoSell_SyncForUser_GetContainersError(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		byUserErr: fmt.Errorf("db error"),
	}
	forSaleRepo := &mockAutoSellForSaleRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get auto-sell containers for user")
}

func Test_AutoSell_SyncForUser_NoJitaPrice_DeactivatesExisting(t *testing.T) {
	containerID := int64(1)
	existingListingID := int64(99)

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{
			{
				ID:                  existingListingID,
				TypeID:              34,
				AutoSellContainerID: &containerID,
				IsActive:            true,
			},
		},
	}

	// No prices available
	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// Deactivated twice: once from "no price" path, once from "not in activeTypes" path
	assert.Len(t, forSaleRepo.upsertedItems, 2)
	assert.False(t, forSaleRepo.upsertedItems[0].IsActive)
	assert.False(t, forSaleRepo.upsertedItems[1].IsActive)
}

func Test_AutoSell_SyncForUser_ZeroBuyPrice_DeactivatesExisting(t *testing.T) {
	containerID := int64(1)
	zeroBuy := 0.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{
			{
				ID:                  99,
				TypeID:              34,
				AutoSellContainerID: &containerID,
				IsActive:            true,
			},
		},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &zeroBuy},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// Deactivated twice: once from "zero price" path, once from "not in activeTypes" path
	assert.Len(t, forSaleRepo.upsertedItems, 2)
	assert.False(t, forSaleRepo.upsertedItems[0].IsActive)
	assert.False(t, forSaleRepo.upsertedItems[1].IsActive)
}

func Test_AutoSell_SyncForUser_ItemRemovedFromContainer_Deactivated(t *testing.T) {
	containerID := int64(1)
	buyPrice := 100.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		// Container is now empty
		containerItems: []*models.ContainerItem{},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{
			{
				ID:                  99,
				TypeID:              34,
				AutoSellContainerID: &containerID,
				IsActive:            true,
			},
		},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// The existing listing should be deactivated
	assert.Len(t, forSaleRepo.upsertedItems, 1)
	assert.False(t, forSaleRepo.upsertedItems[0].IsActive)
	assert.Equal(t, int64(34), forSaleRepo.upsertedItems[0].TypeID)
}

func Test_AutoSell_SyncForUser_MultipleItems(t *testing.T) {
	containerID := int64(1)
	buyPriceTrit := 5.0
	buyPricePyer := 10.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 80.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
			{TypeID: 35, Quantity: 500},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPriceTrit},
			35: {TypeID: 35, RegionID: 10000002, BuyPrice: &buyPricePyer},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 2)

	// Check both items were upserted with correct prices (80% of buy price)
	pricesByType := make(map[int64]float64)
	for _, item := range forSaleRepo.upsertedItems {
		pricesByType[item.TypeID] = item.PricePerUnit
	}
	assert.Equal(t, 4.0, pricesByType[34])  // 5 * 80 / 100
	assert.Equal(t, 8.0, pricesByType[35])  // 10 * 80 / 100
}

func Test_AutoSell_SyncForUser_GetItemsError_ContinuesOtherContainers(t *testing.T) {
	buyPrice := 100.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              1,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		// GetItemsInContainer will fail
		containerItemErr: fmt.Errorf("db error"),
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	// SyncForUser logs errors per-container but returns nil
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 0)
}

func Test_AutoSell_SyncForUser_GetPricesError_ContinuesOtherContainers(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              1,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		pricesErr: fmt.Errorf("price lookup error"),
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 0)
}

func Test_AutoSell_SyncForUser_WithDivisionNumber(t *testing.T) {
	containerID := int64(1)
	divNum := 3
	buyPrice := 100.0

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "corporation",
				OwnerID:         2001,
				LocationID:      60003760,
				ContainerID:     9000,
				DivisionNumber:  &divNum,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 500},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 1)

	upserted := forSaleRepo.upsertedItems[0]
	assert.Equal(t, "corporation", upserted.OwnerType)
	assert.Equal(t, int64(2001), upserted.OwnerID)
	assert.NotNil(t, upserted.DivisionNumber)
	assert.Equal(t, 3, *upserted.DivisionNumber)
}

// --- SyncForAllUsers Tests ---

func Test_AutoSell_SyncForAllUsers_Success(t *testing.T) {
	buyPrice := 100.0

	autoSellRepo := &mockAutoSellContainersRepo{
		allActive: []*models.AutoSellContainer{
			{
				ID:              1,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
			{
				ID:              2,
				UserID:          99,
				OwnerType:       "character",
				OwnerID:         67890,
				LocationID:      60003760,
				ContainerID:     9001,
				PricePercentage: 85.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: &buyPrice},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.NoError(t, err)
	// Should have upserted for both containers
	assert.Len(t, forSaleRepo.upsertedItems, 2)
}

func Test_AutoSell_SyncForAllUsers_NoContainers(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		allActive: []*models.AutoSellContainer{},
	}
	forSaleRepo := &mockAutoSellForSaleRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.NoError(t, err)
}

func Test_AutoSell_SyncForAllUsers_GetAllActiveError(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		allActiveErr: fmt.Errorf("db error"),
	}
	forSaleRepo := &mockAutoSellForSaleRepo{}
	marketRepo := &mockAutoSellMarketRepo{}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all active auto-sell containers")
}

func Test_AutoSell_SyncForUser_NilBuyPrice_DeactivatesExisting(t *testing.T) {
	containerID := int64(1)

	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              containerID,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{
			{TypeID: 34, Quantity: 1000},
		},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{
			{
				ID:                  99,
				TypeID:              34,
				AutoSellContainerID: &containerID,
				IsActive:            true,
			},
		},
	}

	// Price exists but BuyPrice is nil
	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{
			34: {TypeID: 34, RegionID: 10000002, BuyPrice: nil},
		},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	// Deactivated twice: once from "nil price" path, once from "not in activeTypes" path
	assert.Len(t, forSaleRepo.upsertedItems, 2)
	assert.False(t, forSaleRepo.upsertedItems[0].IsActive)
	assert.False(t, forSaleRepo.upsertedItems[1].IsActive)
}

func Test_AutoSell_SyncForUser_EmptyContainer_NoItems(t *testing.T) {
	autoSellRepo := &mockAutoSellContainersRepo{
		byUserContainers: []*models.AutoSellContainer{
			{
				ID:              1,
				UserID:          42,
				OwnerType:       "character",
				OwnerID:         12345,
				LocationID:      60003760,
				ContainerID:     9000,
				PricePercentage: 90.0,
			},
		},
		containerItems: []*models.ContainerItem{},
	}

	forSaleRepo := &mockAutoSellForSaleRepo{
		activeListings: []*models.ForSaleItem{},
	}

	marketRepo := &mockAutoSellMarketRepo{
		prices: map[int64]*models.MarketPrice{},
	}

	u := newAutoSellUpdater(autoSellRepo, forSaleRepo, marketRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, forSaleRepo.upsertedItems, 0)
}

func Test_AutoSell_Constructor(t *testing.T) {
	u := updaters.NewAutoSell(
		&mockAutoSellContainersRepo{},
		&mockAutoSellForSaleRepo{},
		&mockAutoSellMarketRepo{},
	)
	assert.NotNil(t, u)
}
