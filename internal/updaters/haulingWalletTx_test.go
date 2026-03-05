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

// --- Mocks for haulingWalletTx ---

type MockHaulingWalletTxUserRepo struct {
	mock.Mock
}

func (m *MockHaulingWalletTxUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type MockHaulingWalletTxCharacterRepo struct {
	mock.Mock
}

func (m *MockHaulingWalletTxCharacterRepo) GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.Character), args.Error(1)
}

func (m *MockHaulingWalletTxCharacterRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	args := m.Called(ctx, id, userID, token, refreshToken, expiresOn)
	return args.Error(0)
}

type MockHaulingWalletTxRunsRepo struct {
	mock.Mock
}

func (m *MockHaulingWalletTxRunsRepo) ListSellingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRun), args.Error(1)
}

type MockHaulingWalletTxItemsRepo struct {
	mock.Mock
}

func (m *MockHaulingWalletTxItemsRepo) GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRunItem), args.Error(1)
}

func (m *MockHaulingWalletTxItemsRepo) UpdateItemActualSellPrice(ctx context.Context, itemID int64, runID int64, actualSellPriceISK float64) error {
	args := m.Called(ctx, itemID, runID, actualSellPriceISK)
	return args.Error(0)
}

type MockHaulingWalletTxEsiClient struct {
	mock.Mock
}

func (m *MockHaulingWalletTxEsiClient) GetCharacterWalletTransactions(ctx context.Context, characterID int64, token string) ([]*client.WalletTransaction, error) {
	args := m.Called(ctx, characterID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*client.WalletTransaction), args.Error(1)
}

func (m *MockHaulingWalletTxEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RefreshedToken), args.Error(1)
}

func setupWalletTxUpdater() (*updaters.HaulingWalletTxUpdater, *MockHaulingWalletTxUserRepo, *MockHaulingWalletTxCharacterRepo, *MockHaulingWalletTxRunsRepo, *MockHaulingWalletTxItemsRepo, *MockHaulingWalletTxEsiClient) {
	userRepo := new(MockHaulingWalletTxUserRepo)
	charRepo := new(MockHaulingWalletTxCharacterRepo)
	runsRepo := new(MockHaulingWalletTxRunsRepo)
	itemsRepo := new(MockHaulingWalletTxItemsRepo)
	esiClient := new(MockHaulingWalletTxEsiClient)

	updater := updaters.NewHaulingWalletTx(userRepo, charRepo, runsRepo, itemsRepo, esiClient)
	return updater, userRepo, charRepo, runsRepo, itemsRepo, esiClient
}

func makeCharWithWalletScope(id int64, userID int64) *repositories.Character {
	return &repositories.Character{
		ID:                id,
		UserID:            userID,
		Name:              "Test Wallet Character",
		EsiToken:          "valid-token",
		EsiScopes:         "esi-wallet.read_character_wallet.v1",
		EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
	}
}

// --- Tests ---

func Test_HaulingWalletTx_UpdateAllUsers_NoUsers(t *testing.T) {
	updater, userRepo, _, _, _, _ := setupWalletTxUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{}, nil)

	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingWalletTx_UpdateAllUsers_GetIDsError(t *testing.T) {
	updater, userRepo, _, _, _, _ := setupWalletTxUpdater()

	userRepo.On("GetAllIDs", mock.Anything).Return(nil, errors.New("db error"))

	err := updater.UpdateAllUsers(context.Background())
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func Test_HaulingWalletTx_UpdateUserTransactions_NoRuns(t *testing.T) {
	updater, _, _, runsRepo, _, _ := setupWalletTxUpdater()
	userID := int64(100)

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)
	runsRepo.AssertExpectations(t)
}

func Test_HaulingWalletTx_UpdateUserTransactions_NoCharacters(t *testing.T) {
	updater, _, charRepo, runsRepo, _, _ := setupWalletTxUpdater()
	userID := int64(100)

	run := makeSellingRun(int64(1), userID, int64(10000043), time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)
	charRepo.AssertExpectations(t)
}

func Test_HaulingWalletTx_UpdateUserTransactions_CharMissingScope(t *testing.T) {
	updater, _, charRepo, runsRepo, _, esiClient := setupWalletTxUpdater()
	userID := int64(100)

	run := makeSellingRun(int64(1), userID, int64(10000043), time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{
		{
			ID:                int64(500),
			UserID:            userID,
			EsiScopes:         "esi-markets.read_character_orders.v1", // no wallet scope
			EsiToken:          "token",
			EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
		},
	}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)
	esiClient.AssertNotCalled(t, "GetCharacterWalletTransactions", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingWalletTx_UpdateUserTransactions_UpdatesActualSellPrice(t *testing.T) {
	updater, _, charRepo, runsRepo, itemsRepo, esiClient := setupWalletTxUpdater()
	userID := int64(100)
	runID := int64(1)
	itemID := int64(10)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              itemID,
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
	}
	char := makeCharWithWalletScope(int64(500), userID)

	// Sell transaction after run was created
	txDate := time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	tx := &client.WalletTransaction{
		TransactionID: int64(7001),
		TypeID:        int64(34),
		UnitPrice:     1500.0,
		Quantity:      int64(60),
		IsBuy:         false,
		Date:          txDate,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterWalletTransactions", mock.Anything, char.ID, char.EsiToken).Return([]*client.WalletTransaction{tx}, nil)
	itemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)
	itemsRepo.On("UpdateItemActualSellPrice", mock.Anything, itemID, runID, 1500.0).Return(nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)

	itemsRepo.AssertExpectations(t)
}

func Test_HaulingWalletTx_UpdateUserTransactions_SkipsBuyTransactions(t *testing.T) {
	updater, _, charRepo, runsRepo, itemsRepo, esiClient := setupWalletTxUpdater()
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
	}
	char := makeCharWithWalletScope(int64(500), userID)

	// Buy transaction — should be ignored
	txDate := time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	tx := &client.WalletTransaction{
		TransactionID: int64(7001),
		TypeID:        int64(34),
		UnitPrice:     1000.0,
		Quantity:      int64(100),
		IsBuy:         true, // buy, not sell
		Date:          txDate,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterWalletTransactions", mock.Anything, char.ID, char.EsiToken).Return([]*client.WalletTransaction{tx}, nil)
	itemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)

	itemsRepo.AssertNotCalled(t, "UpdateItemActualSellPrice", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingWalletTx_UpdateUserTransactions_SkipsTxBeforeRunCreated(t *testing.T) {
	updater, _, charRepo, runsRepo, itemsRepo, esiClient := setupWalletTxUpdater()
	userID := int64(100)
	runID := int64(1)

	// Run created now
	runCreatedAt := time.Now().Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt)

	item := &models.HaulingRunItem{
		ID:              int64(10),
		RunID:           runID,
		TypeID:          int64(34),
		TypeName:        "Tritanium",
		QuantityPlanned: int64(100),
	}
	char := makeCharWithWalletScope(int64(500), userID)

	// Transaction BEFORE run was created
	txBeforeRun := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	tx := &client.WalletTransaction{
		TransactionID: int64(7001),
		TypeID:        int64(34),
		UnitPrice:     1500.0,
		Quantity:      int64(60),
		IsBuy:         false,
		Date:          txBeforeRun,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterWalletTransactions", mock.Anything, char.ID, char.EsiToken).Return([]*client.WalletTransaction{tx}, nil)
	itemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)

	itemsRepo.AssertNotCalled(t, "UpdateItemActualSellPrice", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingWalletTx_UpdateUserTransactions_NoUpdateIfPriceUnchanged(t *testing.T) {
	updater, _, charRepo, runsRepo, itemsRepo, esiClient := setupWalletTxUpdater()
	userID := int64(100)
	runID := int64(1)
	itemID := int64(10)

	runCreatedAt := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	run := makeSellingRun(runID, userID, int64(10000043), runCreatedAt)

	existingPrice := 1500.0
	item := &models.HaulingRunItem{
		ID:                 itemID,
		RunID:              runID,
		TypeID:             int64(34),
		TypeName:           "Tritanium",
		QuantityPlanned:    int64(100),
		ActualSellPriceISK: &existingPrice, // already set to 1500
	}
	char := makeCharWithWalletScope(int64(500), userID)

	txDate := time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	tx := &client.WalletTransaction{
		TransactionID: int64(7001),
		TypeID:        int64(34),
		UnitPrice:     1500.0, // same price
		Quantity:      int64(60),
		IsBuy:         false,
		Date:          txDate,
	}

	runsRepo.On("ListSellingByUser", mock.Anything, userID).Return([]*models.HaulingRun{run}, nil)
	charRepo.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	esiClient.On("GetCharacterWalletTransactions", mock.Anything, char.ID, char.EsiToken).Return([]*client.WalletTransaction{tx}, nil)
	itemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)

	// Should NOT call UpdateItemActualSellPrice since price is unchanged
	itemsRepo.AssertNotCalled(t, "UpdateItemActualSellPrice", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingWalletTx_UpdateUserTransactions_RefreshesExpiredToken(t *testing.T) {
	updater, _, charRepo, runsRepo, itemsRepo, esiClient := setupWalletTxUpdater()
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
	}

	char := &repositories.Character{
		ID:                int64(500),
		UserID:            userID,
		EsiScopes:         "esi-wallet.read_character_wallet.v1",
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
	esiClient.On("GetCharacterWalletTransactions", mock.Anything, char.ID, "new-token").Return([]*client.WalletTransaction{}, nil)
	itemsRepo.On("GetItemsByRunID", mock.Anything, runID).Return([]*models.HaulingRunItem{item}, nil)

	err := updater.UpdateUserTransactions(context.Background(), userID)
	assert.NoError(t, err)

	esiClient.AssertExpectations(t)
	charRepo.AssertExpectations(t)
}
