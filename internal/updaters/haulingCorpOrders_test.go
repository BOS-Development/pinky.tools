package updaters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks for haulingCorpOrders ---

type MockHaulingCorpUserRepo struct {
	mock.Mock
}

func (m *MockHaulingCorpUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type MockHaulingCorpCorpRepo struct {
	mock.Mock
}

func (m *MockHaulingCorpCorpRepo) Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.PlayerCorporation), args.Error(1)
}

func (m *MockHaulingCorpCorpRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	args := m.Called(ctx, id, userID, token, refreshToken, expiresOn)
	return args.Error(0)
}

type MockHaulingCorpRunsRepo struct {
	mock.Mock
}

func (m *MockHaulingCorpRunsRepo) ListAccumulatingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRun), args.Error(1)
}

type MockHaulingCorpRunItemsRepo struct {
	mock.Mock
}

func (m *MockHaulingCorpRunItemsRepo) GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRunItem), args.Error(1)
}

type MockHaulingCorpItemsRepo struct {
	mock.Mock
}

func (m *MockHaulingCorpItemsRepo) UpdateItemAcquired(ctx context.Context, itemID int64, runID int64, quantityAcquired int64) error {
	args := m.Called(ctx, itemID, runID, quantityAcquired)
	return args.Error(0)
}

type MockHaulingCorpEsiClient struct {
	mock.Mock
}

func (m *MockHaulingCorpEsiClient) GetCorporationOrders(ctx context.Context, corporationID int64, token string) ([]*client.CorpOrder, error) {
	args := m.Called(ctx, corporationID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*client.CorpOrder), args.Error(1)
}

func (m *MockHaulingCorpEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RefreshedToken), args.Error(1)
}

func setupCorpOrdersUpdater() (*updaters.HaulingCorpOrdersUpdater, *MockHaulingCorpUserRepo, *MockHaulingCorpCorpRepo, *MockHaulingCorpRunsRepo, *MockHaulingCorpRunItemsRepo, *MockHaulingCorpItemsRepo, *MockHaulingCorpEsiClient) {
	userRepo := new(MockHaulingCorpUserRepo)
	corpRepo := new(MockHaulingCorpCorpRepo)
	runsRepo := new(MockHaulingCorpRunsRepo)
	runItemsRepo := new(MockHaulingCorpRunItemsRepo)
	itemsRepo := new(MockHaulingCorpItemsRepo)
	esiClient := new(MockHaulingCorpEsiClient)

	updater := updaters.NewHaulingCorpOrders(userRepo, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient)
	return updater, userRepo, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient
}

func makeCorpWithOrdersScope(id int64, userID int64) repositories.PlayerCorporation {
	return repositories.PlayerCorporation{
		ID:        id,
		UserID:    userID,
		Name:      "Test Corp",
		EsiToken:  "valid-token",
		EsiScopes: "esi-corporations.read_orders.v1",
		EsiExpiresOn: time.Now().Add(1 * time.Hour),
	}
}

// --- Tests ---

func Test_HaulingCorpOrders_UpdateAllUsers_NoUsers(t *testing.T) {
	updater, userRepo, _, _, _, _, _ := setupCorpOrdersUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{}, nil)

	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingCorpOrders_UpdateAllUsers_GetIDsError(t *testing.T) {
	updater, userRepo, _, _, _, _, _ := setupCorpOrdersUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return(nil, errors.New("db error"))

	err := updater.UpdateAllUsers(context.Background())
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingCorpOrders_UpdateUserOrders_NoRuns(t *testing.T) {
	updater, _, _, runsRepo, _, _, _ := setupCorpOrdersUpdater()
	userID := int64(100)

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	runsRepo.AssertExpectations(t)
}

func Test_HaulingCorpOrders_UpdateUserOrders_NoCorporations(t *testing.T) {
	updater, _, corpRepo, runsRepo, _, _, _ := setupCorpOrdersUpdater()
	userID := int64(100)

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{
		{ID: int64(1), UserID: userID, Status: "ACCUMULATING"},
	}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	runsRepo.AssertExpectations(t)
	corpRepo.AssertExpectations(t)
}

func Test_HaulingCorpOrders_UpdateUserOrders_CorpMissingScope(t *testing.T) {
	updater, _, corpRepo, runsRepo, _, _, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{
		{ID: int64(1), UserID: userID, Status: "ACCUMULATING"},
	}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{
		{
			ID:        int64(1000),
			UserID:    userID,
			EsiScopes: "esi-industry.read_corporation_jobs.v1", // no orders scope
			EsiToken:  "token",
			EsiExpiresOn: time.Now().Add(1 * time.Hour),
		},
	}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	// ESI should not be called since scope is missing
	esiClient.AssertNotCalled(t, "GetCorporationOrders", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCorpOrders_UpdateUserOrders_MatchesOrderToItem(t *testing.T) {
	updater, _, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)
	runID := int64(1)
	itemID := int64(10)

	run := &models.HaulingRun{ID: runID, UserID: userID, Status: "ACCUMULATING"}
	item := &models.HaulingRunItem{
		ID:               itemID,
		RunID:            runID,
		TypeID:           int64(34),
		TypeName:         "Tritanium",
		QuantityPlanned:  int64(100),
		QuantityAcquired: int64(0),
	}
	corp := makeCorpWithOrdersScope(int64(1000), userID)
	order := &client.CorpOrder{
		OrderID:      int64(555),
		TypeID:       int64(34),
		VolumeTotal:  int64(100),
		VolumeRemain: int64(40), // 60 filled
		IsBuyOrder:   true,
		IssuedBy:     int64(999),
	}

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{corp}, nil)
	esiClient.On("GetCorporationOrders", mock.Anything, corp.ID, corp.EsiToken).Return([]*client.CorpOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)
	// quantity_acquired should be updated to 60 (100-40)
	itemsRepo.On("UpdateItemAcquired", mock.Anything, itemID, runID, int64(60)).Return(nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runsRepo.AssertExpectations(t)
	corpRepo.AssertExpectations(t)
	esiClient.AssertExpectations(t)
	runItemsRepo.AssertExpectations(t)
	itemsRepo.AssertExpectations(t)
}

func Test_HaulingCorpOrders_UpdateUserOrders_NoMatchingOrder(t *testing.T) {
	updater, _, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	run := &models.HaulingRun{ID: runID, UserID: userID, Status: "ACCUMULATING"}
	item := &models.HaulingRunItem{
		ID:               int64(10),
		RunID:            runID,
		TypeID:           int64(35), // Pyerite - no order for this type
		TypeName:         "Pyerite",
		QuantityPlanned:  int64(100),
		QuantityAcquired: int64(0),
	}
	corp := makeCorpWithOrdersScope(int64(1000), userID)
	order := &client.CorpOrder{
		OrderID:      int64(555),
		TypeID:       int64(34), // Tritanium - different type
		VolumeTotal:  int64(100),
		VolumeRemain: int64(40),
		IsBuyOrder:   true,
	}

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{corp}, nil)
	esiClient.On("GetCorporationOrders", mock.Anything, corp.ID, corp.EsiToken).Return([]*client.CorpOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	// UpdateItemAcquired should NOT be called since types don't match
	itemsRepo.AssertNotCalled(t, "UpdateItemAcquired", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCorpOrders_UpdateUserOrders_SkipsSellOrders(t *testing.T) {
	updater, _, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	run := &models.HaulingRun{ID: runID, UserID: userID, Status: "ACCUMULATING"}
	item := &models.HaulingRunItem{
		ID:               int64(10),
		RunID:            runID,
		TypeID:           int64(34),
		TypeName:         "Tritanium",
		QuantityPlanned:  int64(100),
		QuantityAcquired: int64(0),
	}
	corp := makeCorpWithOrdersScope(int64(1000), userID)
	// Sell order — should be ignored
	order := &client.CorpOrder{
		OrderID:      int64(555),
		TypeID:       int64(34),
		VolumeTotal:  int64(100),
		VolumeRemain: int64(40),
		IsBuyOrder:   false, // sell order
	}

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{corp}, nil)
	esiClient.On("GetCorporationOrders", mock.Anything, corp.ID, corp.EsiToken).Return([]*client.CorpOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	itemsRepo.AssertNotCalled(t, "UpdateItemAcquired", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCorpOrders_UpdateUserOrders_AlreadyMatchingQuantity(t *testing.T) {
	updater, _, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	run := &models.HaulingRun{ID: runID, UserID: userID, Status: "ACCUMULATING"}
	item := &models.HaulingRunItem{
		ID:               int64(10),
		RunID:            runID,
		TypeID:           int64(34),
		TypeName:         "Tritanium",
		QuantityPlanned:  int64(100),
		QuantityAcquired: int64(60), // already at 60
	}
	corp := makeCorpWithOrdersScope(int64(1000), userID)
	order := &client.CorpOrder{
		OrderID:      int64(555),
		TypeID:       int64(34),
		VolumeTotal:  int64(100),
		VolumeRemain: int64(40), // 60 filled — same as current
		IsBuyOrder:   true,
	}

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{corp}, nil)
	esiClient.On("GetCorporationOrders", mock.Anything, corp.ID, corp.EsiToken).Return([]*client.CorpOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	// Should NOT update since quantity is already correct
	itemsRepo.AssertNotCalled(t, "UpdateItemAcquired", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCorpOrders_UpdateUserOrders_RefreshesExpiredToken(t *testing.T) {
	updater, _, corpRepo, runsRepo, runItemsRepo, itemsRepo, esiClient := setupCorpOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	run := &models.HaulingRun{ID: runID, UserID: userID, Status: "ACCUMULATING"}
	item := &models.HaulingRunItem{
		ID: int64(10), RunID: runID, TypeID: int64(34), TypeName: "Tritanium",
		QuantityPlanned: int64(100), QuantityAcquired: int64(0),
	}
	corp := repositories.PlayerCorporation{
		ID:              int64(1000),
		UserID:          userID,
		EsiScopes:       "esi-corporations.read_orders.v1",
		EsiToken:        "old-token",
		EsiRefreshToken: "refresh-token",
		EsiExpiresOn:    time.Now().Add(-1 * time.Hour), // expired
	}
	refreshed := &client.RefreshedToken{
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	order := &client.CorpOrder{
		TypeID: int64(34), VolumeTotal: int64(100), VolumeRemain: int64(50), IsBuyOrder: true,
	}

	runsRepo.On("ListAccumulatingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{corp}, nil)
	esiClient.On("RefreshAccessToken", mock.Anything, "refresh-token").Return(refreshed, nil)
	corpRepo.On("UpdateTokens", mock.Anything, corp.ID, corp.UserID, "new-token", "new-refresh", refreshed.Expiry).Return(nil)
	esiClient.On("GetCorporationOrders", mock.Anything, corp.ID, "new-token").Return([]*client.CorpOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)
	itemsRepo.On("UpdateItemAcquired", mock.Anything, int64(10), runID, int64(50)).Return(nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	esiClient.AssertExpectations(t)
	corpRepo.AssertExpectations(t)
	itemsRepo.AssertExpectations(t)
}
