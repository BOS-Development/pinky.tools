package updaters_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockCharacterRepo struct {
	characters []*repositories.Character
	getAllErr  error
	updateErr  error
}

func (m *mockCharacterRepo) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	return m.characters, m.getAllErr
}

func (m *mockCharacterRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.updateErr
}

type mockCharacterAssetsRepo struct {
	updateErr           error
	containers          []int64
	containersErr       error
	upsertNamesErr      error
	stationIDs          []int64
	stationIDsErr       error
	updatedAssets       []*models.EveAsset
	updatedCharacterID  int64
	updatedUserID       int64
}

func (m *mockCharacterAssetsRepo) UpdateAssets(ctx context.Context, characterID, userID int64, assets []*models.EveAsset) error {
	m.updatedAssets = assets
	m.updatedCharacterID = characterID
	m.updatedUserID = userID
	return m.updateErr
}

func (m *mockCharacterAssetsRepo) GetAssembledContainers(ctx context.Context, character, user int64) ([]int64, error) {
	return m.containers, m.containersErr
}

func (m *mockCharacterAssetsRepo) UpsertContainerNames(ctx context.Context, characterID, userID int64, locationNames map[int64]string) error {
	return m.upsertNamesErr
}

func (m *mockCharacterAssetsRepo) GetPlayerOwnedStationIDs(ctx context.Context, character, user int64) ([]int64, error) {
	return m.stationIDs, m.stationIDsErr
}

type mockPlayerCorpRepo struct {
	corporations []repositories.PlayerCorporation
	getErr       error
	updateErr    error
	upsertDivErr error
}

func (m *mockPlayerCorpRepo) Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error) {
	return m.corporations, m.getErr
}

func (m *mockPlayerCorpRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return m.updateErr
}

func (m *mockPlayerCorpRepo) UpsertDivisions(ctx context.Context, corp, user int64, divisions *models.CorporationDivisions) error {
	return m.upsertDivErr
}

type mockCorpAssetsRepo struct {
	upsertErr      error
	containers     []int64
	containersErr  error
	upsertNamesErr error
	stationIDs     []int64
	stationIDsErr  error
}

func (m *mockCorpAssetsRepo) Upsert(ctx context.Context, corp, user int64, assets []*models.EveAsset) error {
	return m.upsertErr
}

func (m *mockCorpAssetsRepo) GetAssembledContainers(ctx context.Context, corp, user int64) ([]int64, error) {
	return m.containers, m.containersErr
}

func (m *mockCorpAssetsRepo) UpsertContainerNames(ctx context.Context, corp, user int64, locationNames map[int64]string) error {
	return m.upsertNamesErr
}

func (m *mockCorpAssetsRepo) GetPlayerOwnedStationIDs(ctx context.Context, corp, user int64) ([]int64, error) {
	return m.stationIDs, m.stationIDsErr
}

type mockAssetStationRepo struct {
	upsertErr      error
	filterStaleIDs []int64
	filterStaleErr error
}

func (m *mockAssetStationRepo) Upsert(ctx context.Context, stations []models.Station) error {
	return m.upsertErr
}

func (m *mockAssetStationRepo) FilterStaleStationIDs(ctx context.Context, ids []int64, knownMaxAge, unknownMaxAge time.Duration) ([]int64, error) {
	if m.filterStaleIDs != nil {
		return m.filterStaleIDs, m.filterStaleErr
	}
	// Default: all IDs are stale (preserves existing test behavior)
	return ids, m.filterStaleErr
}

type mockUserTimestampRepo struct {
	updateErr error
	called    bool
}

func (m *mockUserTimestampRepo) UpdateAssetsLastUpdated(ctx context.Context, userID int64) error {
	m.called = true
	return m.updateErr
}

type mockEsiClientForAssets struct {
	charAssets      []*models.EveAsset
	charAssetsErr   error
	charNames       map[int64]string
	charNamesErr    error
	stations        []models.Station
	stationsErr     error
	corpAssets      []*models.EveAsset
	corpAssetsErr   error
	corpNames       map[int64]string
	corpNamesErr    error
	divisions       *models.CorporationDivisions
	divisionsErr    error
	refreshedToken  *client.RefreshedToken
	refreshErr      error
}

func (m *mockEsiClientForAssets) GetCharacterAssets(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	return m.charAssets, m.charAssetsErr
}

func (m *mockEsiClientForAssets) GetCharacterLocationNames(ctx context.Context, characterID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	return m.charNames, m.charNamesErr
}

func (m *mockEsiClientForAssets) GetPlayerOwnedStationInformation(ctx context.Context, token, refresh string, expire time.Time, ids []int64) ([]models.Station, error) {
	return m.stations, m.stationsErr
}

func (m *mockEsiClientForAssets) GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	return m.corpAssets, m.corpAssetsErr
}

func (m *mockEsiClientForAssets) GetCorporationLocationNames(ctx context.Context, corpID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	return m.corpNames, m.corpNamesErr
}

func (m *mockEsiClientForAssets) GetCorporationDivisions(ctx context.Context, corpID int64, token, refresh string, expire time.Time) (*models.CorporationDivisions, error) {
	return m.divisions, m.divisionsErr
}

func (m *mockEsiClientForAssets) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	return m.refreshedToken, m.refreshErr
}

// --- Helper ---

func newTestUpdater(
	charRepo *mockCharacterRepo,
	charAssetsRepo *mockCharacterAssetsRepo,
	stationRepo *mockAssetStationRepo,
	corpRepo *mockPlayerCorpRepo,
	corpAssetsRepo *mockCorpAssetsRepo,
	esiClient *mockEsiClientForAssets,
	timestampRepo *mockUserTimestampRepo,
	concurrency int,
) *updaters.Assets {
	return updaters.NewAssets(charAssetsRepo, charRepo, stationRepo, corpRepo, corpAssetsRepo, esiClient, timestampRepo, concurrency)
}

// --- UpdateCharacterAssets Tests ---

func Test_Assets_UpdateCharacterAssets_Success(t *testing.T) {
	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{100, 200},
		stationIDs: []int64{60003760},
	}
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{{TypeID: 34, Quantity: 1000}},
		charNames:  map[int64]string{100: "Box A", 200: "Box B"},
		stations:   []models.Station{{ID: 60003760, Name: "Jita"}},
	}
	stationRepo := &mockAssetStationRepo{}

	u := newTestUpdater(
		&mockCharacterRepo{},
		charAssetsRepo,
		stationRepo,
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)

	assert.NoError(t, err)
	assert.Equal(t, int64(12345), charAssetsRepo.updatedCharacterID)
	assert.Equal(t, int64(42), charAssetsRepo.updatedUserID)
	assert.Len(t, charAssetsRepo.updatedAssets, 1)
}

func Test_Assets_UpdateCharacterAssets_RefreshesExpiredToken(t *testing.T) {
	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
		charNames:  map[int64]string{},
		stations:   []models.Station{},
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-token",
			RefreshToken: "new-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
	}
	charRepo := &mockCharacterRepo{}

	u := newTestUpdater(
		charRepo,
		charAssetsRepo,
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "expired-token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(-time.Hour), // expired
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.NoError(t, err)
}

func Test_Assets_UpdateCharacterAssets_TokenRefreshError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		refreshErr: fmt.Errorf("refresh failed"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "expired-token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(-time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to refresh token for character 12345")
}

func Test_Assets_UpdateCharacterAssets_GetAssetsError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		charAssetsErr: fmt.Errorf("ESI error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get assets from the esi client")
}

func Test_Assets_UpdateCharacterAssets_UpdateAssetsRepoError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
	}
	charAssetsRepo := &mockCharacterAssetsRepo{
		updateErr: fmt.Errorf("db error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		charAssetsRepo,
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update assets in repository")
}

func Test_Assets_UpdateCharacterAssets_GetContainersError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
	}
	charAssetsRepo := &mockCharacterAssetsRepo{
		containersErr: fmt.Errorf("containers error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		charAssetsRepo,
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get character containers")
}

func Test_Assets_UpdateCharacterAssets_UpsertStationsError(t *testing.T) {
	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{60003760},
	}
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
		charNames:  map[int64]string{},
		stations:   []models.Station{{ID: 60003760, Name: "Jita"}},
	}
	stationRepo := &mockAssetStationRepo{
		upsertErr: fmt.Errorf("station error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		charAssetsRepo,
		stationRepo,
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	char := &repositories.Character{
		ID:                12345,
		UserID:            42,
		EsiToken:          "token",
		EsiRefreshToken:   "refresh",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	err := u.UpdateCharacterAssets(context.Background(), char, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert player owned stations")
}

// --- UpdateCorporationAssets Tests ---

func Test_Assets_UpdateCorporationAssets_Success(t *testing.T) {
	corpAssetsRepo := &mockCorpAssetsRepo{
		containers: []int64{300},
		stationIDs: []int64{60003760},
	}
	esiClient := &mockEsiClientForAssets{
		corpAssets: []*models.EveAsset{{TypeID: 35, Quantity: 500}},
		corpNames:  map[int64]string{300: "Corp Box"},
		stations:   []models.Station{{ID: 60003760, Name: "Jita"}},
		divisions:  &models.CorporationDivisions{},
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		corpAssetsRepo,
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	corp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err := u.UpdateCorporationAssets(context.Background(), corp, 42)
	assert.NoError(t, err)
}

func Test_Assets_UpdateCorporationAssets_RefreshesExpiredToken(t *testing.T) {
	corpAssetsRepo := &mockCorpAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	esiClient := &mockEsiClientForAssets{
		corpAssets: []*models.EveAsset{},
		corpNames:  map[int64]string{},
		stations:   []models.Station{},
		divisions:  &models.CorporationDivisions{},
		refreshedToken: &client.RefreshedToken{
			AccessToken:  "new-token",
			RefreshToken: "new-refresh",
			Expiry:       time.Now().Add(time.Hour),
		},
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		corpAssetsRepo,
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	corp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		EsiToken:        "expired-token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(-time.Hour),
	}

	err := u.UpdateCorporationAssets(context.Background(), corp, 42)
	assert.NoError(t, err)
}

func Test_Assets_UpdateCorporationAssets_TokenRefreshError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		refreshErr: fmt.Errorf("refresh failed"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	corp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		EsiToken:        "expired-token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(-time.Hour),
	}

	err := u.UpdateCorporationAssets(context.Background(), corp, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to refresh token for corporation 2001")
}

func Test_Assets_UpdateCorporationAssets_GetAssetsError(t *testing.T) {
	esiClient := &mockEsiClientForAssets{
		corpAssetsErr: fmt.Errorf("ESI error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	corp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err := u.UpdateCorporationAssets(context.Background(), corp, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get corp assets")
}

func Test_Assets_UpdateCorporationAssets_UpsertDivisionsError(t *testing.T) {
	corpAssetsRepo := &mockCorpAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	esiClient := &mockEsiClientForAssets{
		corpAssets: []*models.EveAsset{},
		corpNames:  map[int64]string{},
		stations:   []models.Station{},
		divisions:  &models.CorporationDivisions{},
	}
	corpRepo := &mockPlayerCorpRepo{
		upsertDivErr: fmt.Errorf("divisions error"),
	}

	u := newTestUpdater(
		&mockCharacterRepo{},
		&mockCharacterAssetsRepo{},
		&mockAssetStationRepo{},
		corpRepo,
		corpAssetsRepo,
		esiClient,
		&mockUserTimestampRepo{},
		5,
	)

	corp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err := u.UpdateCorporationAssets(context.Background(), corp, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert corporation divisions")
}

// --- UpdateUserAssets Tests ---

func Test_Assets_UpdateUserAssets_Success(t *testing.T) {
	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	corpAssetsRepo := &mockCorpAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	charRepo := &mockCharacterRepo{
		characters: []*repositories.Character{
			{ID: 1, UserID: 42, EsiToken: "t", EsiRefreshToken: "r", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
		},
	}
	corpRepo := &mockPlayerCorpRepo{
		corporations: []repositories.PlayerCorporation{
			{ID: 2001, UserID: 42, EsiToken: "t", EsiRefreshToken: "r", EsiExpiresOn: time.Now().Add(time.Hour)},
		},
	}
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
		charNames:  map[int64]string{},
		corpAssets: []*models.EveAsset{},
		corpNames:  map[int64]string{},
		stations:   []models.Station{},
		divisions:  &models.CorporationDivisions{},
	}
	timestampRepo := &mockUserTimestampRepo{}

	u := newTestUpdater(charRepo, charAssetsRepo, &mockAssetStationRepo{}, corpRepo, corpAssetsRepo, esiClient, timestampRepo, 5)

	err := u.UpdateUserAssets(context.Background(), 42)

	assert.NoError(t, err)
	assert.True(t, timestampRepo.called, "should update assets_last_updated_at timestamp")
}

func Test_Assets_UpdateUserAssets_GetCharactersError(t *testing.T) {
	charRepo := &mockCharacterRepo{
		getAllErr: fmt.Errorf("db error"),
	}

	u := newTestUpdater(charRepo, &mockCharacterAssetsRepo{}, &mockAssetStationRepo{}, &mockPlayerCorpRepo{}, &mockCorpAssetsRepo{}, &mockEsiClientForAssets{}, &mockUserTimestampRepo{}, 5)

	err := u.UpdateUserAssets(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user chars from repository")
}

func Test_Assets_UpdateUserAssets_GetCorporationsError(t *testing.T) {
	charRepo := &mockCharacterRepo{
		characters: []*repositories.Character{},
	}
	corpRepo := &mockPlayerCorpRepo{
		getErr: fmt.Errorf("db error"),
	}

	u := newTestUpdater(charRepo, &mockCharacterAssetsRepo{}, &mockAssetStationRepo{}, corpRepo, &mockCorpAssetsRepo{}, &mockEsiClientForAssets{}, &mockUserTimestampRepo{}, 5)

	err := u.UpdateUserAssets(context.Background(), 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user player corporations")
}

func Test_Assets_UpdateUserAssets_IndividualFailureDoesNotStopOthers(t *testing.T) {
	// Character update will fail, but corporation update should still succeed
	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	corpAssetsRepo := &mockCorpAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	charRepo := &mockCharacterRepo{
		characters: []*repositories.Character{
			{ID: 1, UserID: 42, EsiToken: "t", EsiRefreshToken: "r", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
		},
	}
	corpRepo := &mockPlayerCorpRepo{
		corporations: []repositories.PlayerCorporation{
			{ID: 2001, UserID: 42, EsiToken: "t", EsiRefreshToken: "r", EsiExpiresOn: time.Now().Add(time.Hour)},
		},
	}
	// Character assets fail from ESI, corporation assets succeed
	esiClient := &mockEsiClientForAssets{
		charAssetsErr: fmt.Errorf("ESI error for character"),
		corpAssets:    []*models.EveAsset{},
		corpNames:     map[int64]string{},
		stations:      []models.Station{},
		divisions:     &models.CorporationDivisions{},
	}
	timestampRepo := &mockUserTimestampRepo{}

	u := newTestUpdater(charRepo, charAssetsRepo, &mockAssetStationRepo{}, corpRepo, corpAssetsRepo, esiClient, timestampRepo, 5)

	err := u.UpdateUserAssets(context.Background(), 42)

	// UpdateUserAssets should NOT return error — individual failures are logged
	assert.NoError(t, err)
	assert.True(t, timestampRepo.called, "timestamp should still be updated even if some entities fail")
}

func Test_Assets_UpdateUserAssets_NoCharactersOrCorps(t *testing.T) {
	charRepo := &mockCharacterRepo{characters: []*repositories.Character{}}
	corpRepo := &mockPlayerCorpRepo{corporations: []repositories.PlayerCorporation{}}
	timestampRepo := &mockUserTimestampRepo{}

	u := newTestUpdater(charRepo, &mockCharacterAssetsRepo{}, &mockAssetStationRepo{}, corpRepo, &mockCorpAssetsRepo{}, &mockEsiClientForAssets{}, timestampRepo, 5)

	err := u.UpdateUserAssets(context.Background(), 42)

	assert.NoError(t, err)
	assert.True(t, timestampRepo.called, "timestamp should be updated even with no characters/corps")
}

func Test_Assets_UpdateUserAssets_TimestampErrorIsLogged(t *testing.T) {
	charRepo := &mockCharacterRepo{characters: []*repositories.Character{}}
	corpRepo := &mockPlayerCorpRepo{corporations: []repositories.PlayerCorporation{}}
	timestampRepo := &mockUserTimestampRepo{updateErr: fmt.Errorf("timestamp error")}

	u := newTestUpdater(charRepo, &mockCharacterAssetsRepo{}, &mockAssetStationRepo{}, corpRepo, &mockCorpAssetsRepo{}, &mockEsiClientForAssets{}, timestampRepo, 5)

	// Should not return error even if timestamp update fails
	err := u.UpdateUserAssets(context.Background(), 42)
	assert.NoError(t, err)
}

func Test_Assets_UpdateUserAssets_ConcurrencyLimit(t *testing.T) {
	// Verify the semaphore limits concurrent goroutines to the configured value
	var maxConcurrent int64
	var current int64

	charAssetsRepo := &mockCharacterAssetsRepo{
		containers: []int64{},
		stationIDs: []int64{},
	}
	esiClient := &mockEsiClientForAssets{
		charAssets: []*models.EveAsset{},
		charNames:  map[int64]string{},
		stations:   []models.Station{},
	}

	// Create 10 characters to test concurrency
	characters := []*repositories.Character{}
	for i := 0; i < 10; i++ {
		characters = append(characters, &repositories.Character{
			ID:                int64(i + 1),
			UserID:            42,
			EsiToken:          "t",
			EsiRefreshToken:   "r",
			EsiTokenExpiresOn: time.Now().Add(time.Hour),
		})
	}

	// Use a custom mock that tracks concurrency
	trackingEsiClient := &concurrencyTrackingEsiClient{
		inner:         esiClient,
		maxConcurrent: &maxConcurrent,
		current:       &current,
	}

	charRepo := &mockCharacterRepo{characters: characters}
	corpRepo := &mockPlayerCorpRepo{corporations: []repositories.PlayerCorporation{}}

	u := updaters.NewAssets(charAssetsRepo, charRepo, &mockAssetStationRepo{}, corpRepo, &mockCorpAssetsRepo{}, trackingEsiClient, &mockUserTimestampRepo{}, 2)

	err := u.UpdateUserAssets(context.Background(), 42)
	assert.NoError(t, err)

	// With concurrency=2, max concurrent should never exceed 2
	assert.LessOrEqual(t, maxConcurrent, int64(2))
}

// Helper mock that tracks concurrency for GetCharacterAssets
type concurrencyTrackingEsiClient struct {
	inner         *mockEsiClientForAssets
	maxConcurrent *int64
	current       *int64
}

func (m *concurrencyTrackingEsiClient) GetCharacterAssets(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	cur := atomic.AddInt64(m.current, 1)
	// Update max if this is higher
	for {
		old := atomic.LoadInt64(m.maxConcurrent)
		if cur <= old || atomic.CompareAndSwapInt64(m.maxConcurrent, old, cur) {
			break
		}
	}
	time.Sleep(5 * time.Millisecond) // simulate work
	atomic.AddInt64(m.current, -1)
	return m.inner.charAssets, m.inner.charAssetsErr
}

func (m *concurrencyTrackingEsiClient) GetCharacterLocationNames(ctx context.Context, characterID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	return m.inner.charNames, m.inner.charNamesErr
}

func (m *concurrencyTrackingEsiClient) GetPlayerOwnedStationInformation(ctx context.Context, token, refresh string, expire time.Time, ids []int64) ([]models.Station, error) {
	return m.inner.stations, m.inner.stationsErr
}

func (m *concurrencyTrackingEsiClient) GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error) {
	return m.inner.corpAssets, m.inner.corpAssetsErr
}

func (m *concurrencyTrackingEsiClient) GetCorporationLocationNames(ctx context.Context, corpID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error) {
	return m.inner.corpNames, m.inner.corpNamesErr
}

func (m *concurrencyTrackingEsiClient) GetCorporationDivisions(ctx context.Context, corpID int64, token, refresh string, expire time.Time) (*models.CorporationDivisions, error) {
	return m.inner.divisions, m.inner.divisionsErr
}

func (m *concurrencyTrackingEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	return m.inner.refreshedToken, m.inner.refreshErr
}

func Test_Assets_Constructor(t *testing.T) {
	u := updaters.NewAssets(
		&mockCharacterAssetsRepo{},
		&mockCharacterRepo{},
		&mockAssetStationRepo{},
		&mockPlayerCorpRepo{},
		&mockCorpAssetsRepo{},
		&mockEsiClientForAssets{},
		&mockUserTimestampRepo{},
		5,
	)
	assert.NotNil(t, u)
}
