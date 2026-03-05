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

// --- Mocks for haulingCharOrders ---

type MockHaulingCharOrdersUserRepo struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type MockHaulingCharOrdersCharacterRepo struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersCharacterRepo) GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.Character), args.Error(1)
}

func (m *MockHaulingCharOrdersCharacterRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	args := m.Called(ctx, id, userID, token, refreshToken, expiresOn)
	return args.Error(0)
}

type MockHaulingCharOrdersRunsRepo struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersRunsRepo) ListSellingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRun), args.Error(1)
}

func (m *MockHaulingCharOrdersRunsRepo) UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error {
	args := m.Called(ctx, id, userID, status)
	return args.Error(0)
}

type MockHaulingCharOrdersRunItemsRepo struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersRunItemsRepo) GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRunItem), args.Error(1)
}

func (m *MockHaulingCharOrdersRunItemsRepo) UpdateItemSold(ctx context.Context, itemID int64, runID int64, qtySold int64, sellOrderID *int64) error {
	args := m.Called(ctx, itemID, runID, qtySold, sellOrderID)
	return args.Error(0)
}

type MockHaulingCharOrdersNotifier struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersNotifier) NotifyHaulingItemSold(ctx context.Context, userID int64, run *models.HaulingRun, item *models.HaulingRunItem) {
	m.Called(ctx, userID, run, item)
}

func (m *MockHaulingCharOrdersNotifier) NotifyHaulingComplete(ctx context.Context, userID int64, run *models.HaulingRun, summary *models.HaulingRunPnlSummary) {
	m.Called(ctx, userID, run, summary)
}

type MockHaulingCharOrdersEsiClient struct {
	mock.Mock
}

func (m *MockHaulingCharOrdersEsiClient) GetCharacterOrders(ctx context.Context, characterID int64, token string) ([]*client.CharacterOrder, error) {
	args := m.Called(ctx, characterID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*client.CharacterOrder), args.Error(1)
}

func (m *MockHaulingCharOrdersEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RefreshedToken), args.Error(1)
}

func setupCharOrdersUpdater() (*updaters.HaulingCharOrdersUpdater, *MockHaulingCharOrdersUserRepo, *MockHaulingCharOrdersCharacterRepo, *MockHaulingCharOrdersRunsRepo, *MockHaulingCharOrdersRunItemsRepo, *MockHaulingCharOrdersNotifier, *MockHaulingCharOrdersEsiClient) {
	userRepo := new(MockHaulingCharOrdersUserRepo)
	charRepo := new(MockHaulingCharOrdersCharacterRepo)
	runsRepo := new(MockHaulingCharOrdersRunsRepo)
	runItemsRepo := new(MockHaulingCharOrdersRunItemsRepo)
	notifier := new(MockHaulingCharOrdersNotifier)
	esiClient := new(MockHaulingCharOrdersEsiClient)

	updater := updaters.NewHaulingCharOrders(userRepo, charRepo, runsRepo, runItemsRepo, notifier, esiClient)
	return updater, userRepo, charRepo, runsRepo, runItemsRepo, notifier, esiClient
}

func makeCharWithOrdersScope(id int64, userID int64) *repositories.Character {
	return &repositories.Character{
		ID:                id,
		UserID:            userID,
		Name:              "Test Character",
		EsiToken:          "valid-token",
		EsiScopes:         "esi-markets.read_character_orders.v1",
		EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
	}
}

func makeSellingRun(id int64, userID int64, toRegionID int64, createdAt string) *models.HaulingRun {
	return &models.HaulingRun{
		ID:           id,
		UserID:       userID,
		Status:       "SELLING",
		ToRegionID:   toRegionID,
		CreatedAt:    createdAt,
		NotifyTier3:  true,
	}
}

// --- Tests ---

func Test_HaulingCharOrders_UpdateAllUsers_NoUsers(t *testing.T) {
	updater, userRepo, _, _, _, _, _ := setupCharOrdersUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{}, nil)

	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingCharOrders_UpdateAllUsers_GetIDsError(t *testing.T) {
	updater, userRepo, _, _, _, _, _ := setupCharOrdersUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return(nil, errors.New("db error"))

	err := updater.UpdateAllUsers(context.Background())
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingCharOrders_UpdateUserOrders_NoRuns(t *testing.T) {
	updater, _, _, runsRepo, _, _, _ := setupCharOrdersUpdater()
	userID := int64(100)

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	runsRepo.AssertExpectations(t)
}

func Test_HaulingCharOrders_UpdateUserOrders_NoCharacters(t *testing.T) {
	updater, _, charRepo, runsRepo, _, _, _ := setupCharOrdersUpdater()
	userID := int64(100)

	run := makeSellingRun(int64(1), userID, int64(10000043), time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	runsRepo.AssertExpectations(t)
	charRepo.AssertExpectations(t)
}

func Test_HaulingCharOrders_UpdateUserOrders_CharMissingScope(t *testing.T) {
	updater, _, charRepo, runsRepo, _, _, esiClient := setupCharOrdersUpdater()
	userID := int64(100)

	run := makeSellingRun(int64(1), userID, int64(10000043), time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{
		{
			ID:                int64(500),
			UserID:            userID,
			EsiScopes:         "esi-industry.read_character_jobs.v1", // no orders scope
			EsiToken:          "token",
			EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
		},
	}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)
	esiClient.AssertNotCalled(t, "GetCharacterOrders", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCharOrders_UpdateUserOrders_MatchesSellOrderToItem(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, notifier, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)
	itemID := int64(10)
	toRegionID := int64(10000043)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, toRegionID, runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              itemID,
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(0),
	}
	char := makeCharWithOrdersScope(int64(500), userID)

	issuedAfterRun := time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	order := &client.CharacterOrder{
		OrderID:      int64(9001),
		TypeID:       int64(34),
		RegionID:     toRegionID,
		VolumeTotal:  int64(100),
		VolumeRemain: int64(40), // 60 sold
		IsBuyOrder:   false,
		Issued:       issuedAfterRun,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, char.EsiToken).Return([]*client.CharacterOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	orderID := int64(9001)
	runItemsRepo.On("UpdateItemSold", mock.Anything, itemID, runID, int64(60), &orderID).Return(nil)
	// Item not fully sold (60 < 100), so no run auto-complete
	runsRepo.AssertNotCalled(t, "UpdateRunStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runsRepo.AssertExpectations(t)
	charRepo.AssertExpectations(t)
	esiClient.AssertExpectations(t)
	runItemsRepo.AssertExpectations(t)
	notifier.AssertNotCalled(t, "NotifyHaulingItemSold", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCharOrders_UpdateUserOrders_SkipsBuyOrders(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, _, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(0),
	}
	char := makeCharWithOrdersScope(int64(500), userID)

	// Buy order — should be ignored
	order := &client.CharacterOrder{
		OrderID:      int64(9001),
		TypeID:       int64(34),
		RegionID:     int64(10000043),
		VolumeTotal:  int64(100),
		VolumeRemain: int64(50),
		IsBuyOrder:   true,
		Issued:       time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, char.EsiToken).Return([]*client.CharacterOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runItemsRepo.AssertNotCalled(t, "UpdateItemSold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCharOrders_UpdateUserOrders_SkipsOrderBeforeRunCreated(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, _, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	// Run created NOW
	runCreatedAt := time.Now().Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(0),
	}
	char := makeCharWithOrdersScope(int64(500), userID)

	// Order issued BEFORE run was created
	issuedBeforeRun := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	order := &client.CharacterOrder{
		OrderID:      int64(9001),
		TypeID:       int64(34),
		RegionID:     int64(10000043),
		VolumeTotal:  int64(100),
		VolumeRemain: int64(50),
		IsBuyOrder:   false,
		Issued:       issuedBeforeRun,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, char.EsiToken).Return([]*client.CharacterOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runItemsRepo.AssertNotCalled(t, "UpdateItemSold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCharOrders_UpdateUserOrders_SkipsWrongRegion(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, _, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt) // destination: Domain

	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(0),
	}
	char := makeCharWithOrdersScope(int64(500), userID)

	// Order in wrong region
	order := &client.CharacterOrder{
		OrderID:      int64(9001),
		TypeID:       int64(34),
		RegionID:     int64(10000002), // The Forge, not Domain
		VolumeTotal:  int64(100),
		VolumeRemain: int64(50),
		IsBuyOrder:   false,
		Issued:       time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, char.EsiToken).Return([]*client.CharacterOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runItemsRepo.AssertNotCalled(t, "UpdateItemSold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingCharOrders_UpdateUserOrders_AutoCompletesWhenAllSold(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, notifier, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)
	toRegionID := int64(10000043)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, toRegionID, runCreatedAt)

	// Item fully sold: qty_sold == quantity_planned
	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(100), // already fully sold
	}
	char := makeCharWithOrdersScope(int64(500), userID)

	issuedAfterRun := time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	order := &client.CharacterOrder{
		OrderID:      int64(9001),
		TypeID:       int64(34),
		RegionID:     toRegionID,
		VolumeTotal:  int64(100),
		VolumeRemain: int64(0), // 100 sold
		IsBuyOrder:   false,
		Issued:       issuedAfterRun,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, char.EsiToken).Return([]*client.CharacterOrder{order}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)
	// qty_sold already matches — no UpdateItemSold call
	// All sold => auto-complete
	runsRepo.On("UpdateRunStatus", mock.Anything, runID, userID, "COMPLETE").Return(nil)
	notifier.On("NotifyHaulingComplete", mock.Anything, userID, run, (*models.HaulingRunPnlSummary)(nil)).Return().Maybe()

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	runsRepo.AssertExpectations(t)
}

func Test_HaulingCharOrders_UpdateUserOrders_RefreshesExpiredToken(t *testing.T) {
	updater, _, charRepo, runsRepo, runItemsRepo, _, esiClient := setupCharOrdersUpdater()
	userID := int64(100)
	runID := int64(1)
	toRegionID := int64(10000043)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, toRegionID, runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
		QtySold:         int64(0),
	}

	char := &repositories.Character{
		ID:                int64(500),
		UserID:            userID,
		EsiScopes:         "esi-markets.read_character_orders.v1",
		EsiToken:          "old-token",
		EsiRefreshToken:   "refresh-token",
		EsiTokenExpiresOn: time.Now().Add(-1 * time.Hour), // expired
	}
	refreshed := &client.RefreshedToken{
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("RefreshAccessToken", mock.Anything, "refresh-token").Return(refreshed, nil)
	charRepo.On("UpdateTokens", mock.Anything, char.ID, char.UserID, "new-token", "new-refresh", refreshed.Expiry).Return(nil)
	esiClient.On("GetCharacterOrders", mock.Anything, char.ID, "new-token").Return([]*client.CharacterOrder{}, nil)
	runItemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserOrders(context.Background(), userID)
	assert.NoError(t, err)

	esiClient.AssertExpectations(t)
	charRepo.AssertExpectations(t)
}
