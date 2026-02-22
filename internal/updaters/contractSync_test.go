package updaters_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockContractSyncPurchaseRepo struct {
	contractCreated    []*models.PurchaseTransaction
	contractCreatedErr error
	completedCalls     []completeCall
	completeErr        error
}

type completeCall struct {
	PurchaseID    int64
	EveContractID int64
}

func (m *mockContractSyncPurchaseRepo) GetContractCreatedWithKeys(ctx context.Context) ([]*models.PurchaseTransaction, error) {
	return m.contractCreated, m.contractCreatedErr
}

func (m *mockContractSyncPurchaseRepo) CompleteWithContractID(ctx context.Context, purchaseID int64, eveContractID int64) error {
	m.completedCalls = append(m.completedCalls, completeCall{PurchaseID: purchaseID, EveContractID: eveContractID})
	return m.completeErr
}

type mockContractSyncCharRepo struct {
	charactersByUser map[int64][]*repositories.Character
	getErr           error
	tokenUpdateErr   error
}

func (m *mockContractSyncCharRepo) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.charactersByUser[baseUserID], nil
}

func (m *mockContractSyncCharRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.tokenUpdateErr
}

type mockContractSyncCorpRepo struct {
	corpsByUser    map[int64][]repositories.PlayerCorporation
	getErr         error
	tokenUpdateErr error
}

func (m *mockContractSyncCorpRepo) Get(ctx context.Context, userID int64) ([]repositories.PlayerCorporation, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.corpsByUser[userID], nil
}

func (m *mockContractSyncCorpRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.tokenUpdateErr
}

type mockContractSyncEsiClient struct {
	contractsByChar map[int64][]*client.EsiContract
	contractsByCorp map[int64][]*client.EsiContract
	contractsErr    error
	corpContractErr error
	refreshedToken  *client.RefreshedToken
	refreshErr      error
}

func (m *mockContractSyncEsiClient) GetCharacterContracts(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*client.EsiContract, error) {
	if m.contractsErr != nil {
		return nil, m.contractsErr
	}
	return m.contractsByChar[characterID], nil
}

func (m *mockContractSyncEsiClient) GetCorporationContracts(ctx context.Context, corporationID int64, token, refresh string, expire time.Time) ([]*client.EsiContract, error) {
	if m.corpContractErr != nil {
		return nil, m.corpContractErr
	}
	return m.contractsByCorp[corporationID], nil
}

func (m *mockContractSyncEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshedToken, nil
}

// --- Helpers ---

func strPtr(s string) *string { return &s }

func emptyCorpRepo() *mockContractSyncCorpRepo {
	return &mockContractSyncCorpRepo{corpsByUser: map[int64][]repositories.PlayerCorporation{}}
}

// --- Character Contract Tests ---

func Test_ContractSync_AutoCompletesMatchingContract(t *testing.T) {
	key := "PT-42"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 42, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 99999, Type: "item_exchange", Status: "finished", Title: "Items for PT-42"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 1)
	assert.Equal(t, int64(42), purchaseRepo.completedCalls[0].PurchaseID)
	assert.Equal(t, int64(99999), purchaseRepo.completedCalls[0].EveContractID)
}

func Test_ContractSync_NoActionWhenNoPurchases(t *testing.T) {
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{},
	}

	charRepo := &mockContractSyncCharRepo{}
	esiClient := &mockContractSyncEsiClient{}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_SkipsCharacterWithoutScope(t *testing.T) {
	key := "PT-10"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 10, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-assets.read_assets.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 99999, Type: "item_exchange", Status: "finished", Title: "PT-10 delivery"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_NoMatchWhenTitleDoesNotContainKey(t *testing.T) {
	key := "PT-42"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 42, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 88888, Type: "item_exchange", Status: "finished", Title: "Random contract"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_MultiplePurchasesSameKeyAllCompleted(t *testing.T) {
	key := "BATCH-1"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 10, BuyerUserID: 100, ContractKey: &key},
			{ID: 11, BuyerUserID: 100, ContractKey: &key},
			{ID: 12, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 55555, Type: "item_exchange", Status: "finished", Title: "Corp delivery BATCH-1"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 3)
	assert.Equal(t, int64(10), purchaseRepo.completedCalls[0].PurchaseID)
	assert.Equal(t, int64(11), purchaseRepo.completedCalls[1].PurchaseID)
	assert.Equal(t, int64(12), purchaseRepo.completedCalls[2].PurchaseID)
	for _, call := range purchaseRepo.completedCalls {
		assert.Equal(t, int64(55555), call.EveContractID)
	}
}

func Test_ContractSync_IgnoresNonFinishedContracts(t *testing.T) {
	key := "PT-1"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 1, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 111, Type: "item_exchange", Status: "outstanding", Title: "PT-1"},
				{ContractID: 222, Type: "courier", Status: "finished", Title: "PT-1"},
				{ContractID: 333, Type: "item_exchange", Status: "cancelled", Title: "PT-1"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_ESIErrorLogsAndContinues(t *testing.T) {
	key1 := "PT-1"
	key2 := "PT-2"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 1, BuyerUserID: 100, ContractKey: &key1},
			{ID: 2, BuyerUserID: 200, ContractKey: &key2},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
			200: {
				{ID: 3001, UserID: 200, EsiToken: "tok2", EsiRefreshToken: "ref2",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	// ESI returns error for all characters
	esiClient := &mockContractSyncEsiClient{
		contractsErr: errors.New("ESI unavailable"),
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	// SyncAll should not return an error — it logs per-user errors
	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_RefreshesExpiredToken(t *testing.T) {
	key := "PT-5"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 5, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "expired-tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(-1 * time.Hour), // expired
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-tok",
			RefreshToken: "new-ref",
			Expiry:       time.Now().Add(20 * time.Minute),
		},
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 77777, Type: "item_exchange", Status: "finished", Title: "PT-5 minerals"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 1)
	assert.Equal(t, int64(5), purchaseRepo.completedCalls[0].PurchaseID)
}

func Test_ContractSync_GetPurchasesError(t *testing.T) {
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreatedErr: errors.New("db down"),
	}

	charRepo := &mockContractSyncCharRepo{}
	esiClient := &mockContractSyncEsiClient{}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contract_created purchases")
}

// --- Corporation Contract Tests ---

func Test_ContractSync_CorpContractMatchesAndCompletes(t *testing.T) {
	key := "PT-50"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 50, BuyerUserID: 100, ContractKey: &key},
		},
	}

	// No character has the contract
	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	corpRepo := &mockContractSyncCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 5001, UserID: 100, EsiToken: "corp-tok", EsiRefreshToken: "corp-ref",
					EsiExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:    "esi-contracts.read_corporation_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {}, // character has no contracts
		},
		contractsByCorp: map[int64][]*client.EsiContract{
			5001: {
				{ContractID: 88888, Type: "item_exchange", Status: "finished", Title: "Corp delivery PT-50"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, corpRepo, esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 1)
	assert.Equal(t, int64(50), purchaseRepo.completedCalls[0].PurchaseID)
	assert.Equal(t, int64(88888), purchaseRepo.completedCalls[0].EveContractID)
}

func Test_ContractSync_SkipsCorporationWithoutScope(t *testing.T) {
	key := "PT-60"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 60, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{},
	}

	corpRepo := &mockContractSyncCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 5001, UserID: 100, EsiToken: "corp-tok", EsiRefreshToken: "corp-ref",
					EsiExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:    "esi-assets.read_corporation_assets.v1"}, // wrong scope
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByCorp: map[int64][]*client.EsiContract{
			5001: {
				{ContractID: 99999, Type: "item_exchange", Status: "finished", Title: "PT-60 delivery"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, corpRepo, esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_CorpTokenRefresh(t *testing.T) {
	key := "PT-70"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 70, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{},
	}

	corpRepo := &mockContractSyncCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 5001, UserID: 100, EsiToken: "expired-tok", EsiRefreshToken: "corp-ref",
					EsiExpiresOn: time.Now().Add(-1 * time.Hour), // expired
					EsiScopes:    "esi-contracts.read_corporation_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-corp-tok",
			RefreshToken: "new-corp-ref",
			Expiry:       time.Now().Add(20 * time.Minute),
		},
		contractsByCorp: map[int64][]*client.EsiContract{
			5001: {
				{ContractID: 77777, Type: "item_exchange", Status: "finished", Title: "PT-70 materials"},
			},
		},
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, corpRepo, esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 1)
	assert.Equal(t, int64(70), purchaseRepo.completedCalls[0].PurchaseID)
}

func Test_ContractSync_CorpESIErrorLogsAndContinues(t *testing.T) {
	key := "PT-80"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 80, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{},
	}

	corpRepo := &mockContractSyncCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 5001, UserID: 100, EsiToken: "corp-tok", EsiRefreshToken: "corp-ref",
					EsiExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:    "esi-contracts.read_corporation_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		corpContractErr: errors.New("ESI unavailable"),
	}

	syncer := updaters.NewContractSync(purchaseRepo, charRepo, corpRepo, esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 0)
}

func Test_ContractSync_NoCorporationsForBuyer(t *testing.T) {
	key := "PT-90"
	purchaseRepo := &mockContractSyncPurchaseRepo{
		contractCreated: []*models.PurchaseTransaction{
			{ID: 90, BuyerUserID: 100, ContractKey: &key},
		},
	}

	charRepo := &mockContractSyncCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 2001, UserID: 100, EsiToken: "tok", EsiRefreshToken: "ref",
					EsiTokenExpiresOn: time.Now().Add(1 * time.Hour),
					EsiScopes:         "esi-contracts.read_character_contracts.v1"},
			},
		},
	}

	esiClient := &mockContractSyncEsiClient{
		contractsByChar: map[int64][]*client.EsiContract{
			2001: {
				{ContractID: 11111, Type: "item_exchange", Status: "finished", Title: "PT-90 delivery"},
			},
		},
	}

	// No corps for this buyer
	syncer := updaters.NewContractSync(purchaseRepo, charRepo, emptyCorpRepo(), esiClient)
	err := syncer.SyncAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.completedCalls, 1)
	assert.Equal(t, int64(90), purchaseRepo.completedCalls[0].PurchaseID)
}
