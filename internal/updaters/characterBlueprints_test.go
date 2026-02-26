package updaters_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockBlueprintsRepo struct {
	replacedBlueprints map[string][]*models.CharacterBlueprint // keyed by "ownerID:ownerType"
	replaceErr         error
}

func (m *mockBlueprintsRepo) ReplaceBlueprints(ctx context.Context, ownerID int64, ownerType string, userID int64, blueprints []*models.CharacterBlueprint) error {
	if m.replaceErr != nil {
		return m.replaceErr
	}
	if m.replacedBlueprints == nil {
		m.replacedBlueprints = map[string][]*models.CharacterBlueprint{}
	}
	key := string(rune(ownerID)) + ":" + ownerType
	m.replacedBlueprints[key] = blueprints
	return nil
}

type mockBlueprintsEsiClient struct {
	characterBlueprintsByChar map[int64][]*client.EsiBlueprint
	corpBlueprintsByCorp      map[int64][]*client.EsiBlueprint
	charBlueprintsErr         error
	corpBlueprintsErr         error
	refreshedToken            *client.RefreshedToken
	refreshErr                error
}

func (m *mockBlueprintsEsiClient) GetCharacterBlueprints(ctx context.Context, characterID int64, token string) ([]*client.EsiBlueprint, error) {
	if m.charBlueprintsErr != nil {
		return nil, m.charBlueprintsErr
	}
	bps := m.characterBlueprintsByChar[characterID]
	if bps == nil {
		bps = []*client.EsiBlueprint{}
	}
	return bps, nil
}

func (m *mockBlueprintsEsiClient) GetCorporationBlueprints(ctx context.Context, corporationID int64, token string) ([]*client.EsiBlueprint, error) {
	if m.corpBlueprintsErr != nil {
		return nil, m.corpBlueprintsErr
	}
	bps := m.corpBlueprintsByCorp[corporationID]
	if bps == nil {
		bps = []*client.EsiBlueprint{}
	}
	return bps, nil
}

func (m *mockBlueprintsEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshedToken, nil
}

type mockBlueprintsUserRepo struct {
	userIDs []int64
	err     error
}

func (m *mockBlueprintsUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	return m.userIDs, m.err
}

type mockBlueprintsCharRepo struct {
	charactersByUser map[int64][]*repositories.Character
	getErr           error
	tokenUpdateErr   error
}

func (m *mockBlueprintsCharRepo) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.charactersByUser[baseUserID], nil
}

func (m *mockBlueprintsCharRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.tokenUpdateErr
}

type mockBlueprintsCorpRepo struct {
	corpsByUser    map[int64][]repositories.PlayerCorporation
	getErr         error
	tokenUpdateErr error
}

func (m *mockBlueprintsCorpRepo) Get(ctx context.Context, userID int64) ([]repositories.PlayerCorporation, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.corpsByUser[userID], nil
}

func (m *mockBlueprintsCorpRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.tokenUpdateErr
}

// --- Tests ---

func Test_CharacterBlueprintsUpdater_UpdateAllUsers_NoUsers(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{}}
	charRepo := &mockBlueprintsCharRepo{}
	corpRepo := &mockBlueprintsCorpRepo{}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, bpRepo.replacedBlueprints)
}

func Test_CharacterBlueprintsUpdater_UpdatesCharacterBlueprints(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		characterBlueprintsByChar: map[int64][]*client.EsiBlueprint{
			1001: {
				{ItemID: 90001, TypeID: 787, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -1, MaterialEfficiency: 10, TimeEfficiency: 20, Runs: -1},
				{ItemID: 90002, TypeID: 46166, LocationID: 60003760, LocationFlag: "Hangar", Quantity: -2, MaterialEfficiency: 8, TimeEfficiency: 16, Runs: 50},
			},
		},
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Test Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-characters.read_blueprints.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	// Should have stored blueprints for character 1001
	assert.Len(t, bpRepo.replacedBlueprints, 1)
	key := string(rune(1001)) + ":character"
	bps := bpRepo.replacedBlueprints[key]
	assert.Len(t, bps, 2)
	assert.Equal(t, int64(787), bps[0].TypeID)
	assert.Equal(t, 10, bps[0].MaterialEfficiency)
	assert.Equal(t, -1, bps[0].Quantity)
	assert.Equal(t, int64(46166), bps[1].TypeID)
	assert.Equal(t, 8, bps[1].MaterialEfficiency)
	assert.Equal(t, -2, bps[1].Quantity)
}

func Test_CharacterBlueprintsUpdater_SkipsWithoutScope(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		characterBlueprintsByChar: map[int64][]*client.EsiBlueprint{},
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "No Scope Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-assets.read_assets.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	assert.Empty(t, bpRepo.replacedBlueprints)
}

func Test_CharacterBlueprintsUpdater_RefreshesExpiredToken(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		characterBlueprintsByChar: map[int64][]*client.EsiBlueprint{
			1001: {
				{ItemID: 91001, TypeID: 787, LocationID: 60003760, Quantity: -1, MaterialEfficiency: 10, TimeEfficiency: 20, Runs: -1},
			},
		},
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-token",
			RefreshToken: "new-refresh",
			Expiry:       time.Now().Add(20 * time.Minute),
		},
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Expired Token Char", UserID: 100, EsiToken: "expired-token", EsiRefreshToken: "refresh-token", EsiScopes: "esi-characters.read_blueprints.v1", EsiTokenExpiresOn: time.Now().Add(-time.Hour)},
			},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	key := string(rune(1001)) + ":character"
	assert.Len(t, bpRepo.replacedBlueprints[key], 1)
}

func Test_CharacterBlueprintsUpdater_HandlesEsiError(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		charBlueprintsErr: assert.AnError,
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Error Char", UserID: 100, EsiToken: "valid-token", EsiScopes: "esi-characters.read_blueprints.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	// Should not propagate error — logs it and continues
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, bpRepo.replacedBlueprints)
}

func Test_CharacterBlueprintsUpdater_UpdatesCorporationBlueprints(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		corpBlueprintsByCorp: map[int64][]*client.EsiBlueprint{
			3001001: {
				{ItemID: 92001, TypeID: 787, LocationID: 60003760, LocationFlag: "CorpSAG1", Quantity: -1, MaterialEfficiency: 9, TimeEfficiency: 18, Runs: -1},
			},
		},
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 3001001, UserID: 100, Name: "Test Corp", EsiToken: "corp-token", EsiRefreshToken: "corp-refresh", EsiScopes: "esi-corporations.read_blueprints.v1", EsiExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	key := string(rune(3001001)) + ":corporation"
	bps := bpRepo.replacedBlueprints[key]
	assert.Len(t, bps, 1)
	assert.Equal(t, int64(787), bps[0].TypeID)
	assert.Equal(t, 9, bps[0].MaterialEfficiency)
}

func Test_CharacterBlueprintsUpdater_SkipsCorporationWithoutScope(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {},
		},
	}
	corpRepo := &mockBlueprintsCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 3001001, UserID: 100, Name: "Test Corp", EsiToken: "corp-token", EsiScopes: "esi-assets.read_corporation_assets.v1", EsiExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, bpRepo.replacedBlueprints)
}

func Test_CharacterBlueprintsUpdater_RefreshesCorporationExpiredToken(t *testing.T) {
	bpRepo := &mockBlueprintsRepo{}
	esiClient := &mockBlueprintsEsiClient{
		corpBlueprintsByCorp: map[int64][]*client.EsiBlueprint{
			3001001: {
				{ItemID: 93001, TypeID: 788, LocationID: 60003760, Quantity: -1, MaterialEfficiency: 7, TimeEfficiency: 14, Runs: -1},
			},
		},
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-corp-token",
			RefreshToken: "new-corp-refresh",
			Expiry:       time.Now().Add(20 * time.Minute),
		},
	}
	userRepo := &mockBlueprintsUserRepo{userIDs: []int64{100}}
	charRepo := &mockBlueprintsCharRepo{
		charactersByUser: map[int64][]*repositories.Character{100: {}},
	}
	corpRepo := &mockBlueprintsCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 3001001, UserID: 100, Name: "Test Corp", EsiToken: "expired-corp-token", EsiRefreshToken: "corp-refresh", EsiScopes: "esi-corporations.read_blueprints.v1", EsiExpiresOn: time.Now().Add(-time.Hour)},
			},
		},
	}

	updater := updaters.NewCharacterBlueprintsUpdater(userRepo, charRepo, corpRepo, bpRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	key := string(rune(3001001)) + ":corporation"
	assert.Len(t, bpRepo.replacedBlueprints[key], 1)
}
