package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock TradingStationsRepository ---

type MockTradingStationsRepository struct {
	mock.Mock
}

func (m *MockTradingStationsRepository) ListStations(ctx context.Context) ([]*models.TradingStation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TradingStation), args.Error(1)
}

// --- Mock UserTradingStructuresRepository ---

type MockUserTradingStructuresRepository struct {
	mock.Mock
}

func (m *MockUserTradingStructuresRepository) List(ctx context.Context, userID int64) ([]*models.UserTradingStructure, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserTradingStructure), args.Error(1)
}

func (m *MockUserTradingStructuresRepository) Upsert(ctx context.Context, s *models.UserTradingStructure) (*models.UserTradingStructure, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserTradingStructure), args.Error(1)
}

func (m *MockUserTradingStructuresRepository) Delete(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockUserTradingStructuresRepository) UpdateAccessStatus(ctx context.Context, userID int64, structureID int64, accessOK bool) error {
	args := m.Called(ctx, userID, structureID, accessOK)
	return args.Error(0)
}

// --- Mock TradingStructureMarketUpdater ---

type MockTradingStructureMarketUpdater struct {
	mock.Mock
}

func (m *MockTradingStructureMarketUpdater) ScanStructure(ctx context.Context, structureID int64, token string) (bool, error) {
	args := m.Called(ctx, structureID, token)
	return args.Bool(0), args.Error(1)
}

// --- Mock TradingStructureCharacterRepository ---

type MockTradingStructureCharacterRepository struct {
	mock.Mock
}

func (m *MockTradingStructureCharacterRepository) GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.Character), args.Error(1)
}

// --- Mock TradingStructureEsiClient ---

type MockTradingStructureEsiClient struct {
	mock.Mock
}

func (m *MockTradingStructureEsiClient) GetStructureInfo(ctx context.Context, structureID int64, token string) (*client.StructureInfo, error) {
	args := m.Called(ctx, structureID, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StructureInfo), args.Error(1)
}

func (m *MockTradingStructureEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RefreshedToken), args.Error(1)
}

// --- Mock TradingStructureSolarSystemRepository ---

type MockTradingStructureSolarSystemRepository struct {
	mock.Mock
}

func (m *MockTradingStructureSolarSystemRepository) GetRegionIDBySystemID(ctx context.Context, systemID int64) (int64, error) {
	args := m.Called(ctx, systemID)
	return args.Get(0).(int64), args.Error(1)
}

// --- Mock TradingStructureAssetRepository ---

type MockTradingStructureAssetRepository struct {
	mock.Mock
}

func (m *MockTradingStructureAssetRepository) GetPlayerOwnedStationIDs(ctx context.Context, characterID, userID int64) ([]int64, error) {
	args := m.Called(ctx, characterID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

// --- Setup helper ---

type tradingStructuresMocks struct {
	stations   *MockTradingStationsRepository
	structures *MockUserTradingStructuresRepository
	scanner    *MockTradingStructureMarketUpdater
	characters *MockTradingStructureCharacterRepository
	esi        *MockTradingStructureEsiClient
	systems    *MockTradingStructureSolarSystemRepository
	assets     *MockTradingStructureAssetRepository
}

func setupTradingStructuresController() (*controllers.TradingStructuresController, tradingStructuresMocks) {
	mocks := tradingStructuresMocks{
		stations:   new(MockTradingStationsRepository),
		structures: new(MockUserTradingStructuresRepository),
		scanner:    new(MockTradingStructureMarketUpdater),
		characters: new(MockTradingStructureCharacterRepository),
		esi:        new(MockTradingStructureEsiClient),
		systems:    new(MockTradingStructureSolarSystemRepository),
		assets:     new(MockTradingStructureAssetRepository),
	}
	router := &MockRouter{}
	controller := controllers.NewTradingStructures(router, mocks.stations, mocks.structures, mocks.scanner, mocks.characters, mocks.esi, mocks.systems, mocks.assets)
	return controller, mocks
}

// --- Tests: ListStations ---

func Test_TradingStructures_ListStations_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	expected := []*models.TradingStation{
		{ID: int64(1), StationID: int64(60003760), Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant", IsPreset: true},
	}
	mocks.stations.On("ListStations", mock.Anything).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/stations", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListStations(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	stations := result.([]*models.TradingStation)
	assert.Len(t, stations, 1)
	assert.True(t, stations[0].IsPreset)
	mocks.stations.AssertExpectations(t)
}

func Test_TradingStructures_ListStations_Error(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.stations.On("ListStations", mock.Anything).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/stations", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListStations(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: ListStructures ---

func Test_TradingStructures_ListStructures_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	expected := []*models.UserTradingStructure{
		{ID: int64(1), UserID: userID, StructureID: int64(1234567890), Name: "My Structure", AccessOK: true},
	}
	mocks.structures.On("List", mock.Anything, userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListStructures(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	structures := result.([]*models.UserTradingStructure)
	assert.Len(t, structures, 1)
	mocks.structures.AssertExpectations(t)
}

func Test_TradingStructures_ListStructures_Error(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.structures.On("List", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: AddStructure ---

func Test_TradingStructures_AddStructure_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001001)
	structureID := int64(1234567890)

	char := &repositories.Character{ID: charID, EsiRefreshToken: "refresh-token-123"}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.esi.On("RefreshAccessToken", mock.Anything, "refresh-token-123").Return(
		&client.RefreshedToken{AccessToken: "access-token-456"}, nil)
	mocks.esi.On("GetStructureInfo", mock.Anything, structureID, "access-token-456").Return(
		&client.StructureInfo{Name: "New Structure", SolarSystemID: int64(30000142)}, nil)
	mocks.systems.On("GetRegionIDBySystemID", mock.Anything, int64(30000142)).Return(int64(10000002), nil)
	mocks.structures.On("Upsert", mock.Anything, mock.AnythingOfType("*models.UserTradingStructure")).Return(
		&models.UserTradingStructure{
			ID: int64(1), UserID: userID, StructureID: structureID,
			Name: "New Structure", SystemID: int64(30000142), RegionID: int64(10000002),
			CharacterID: charID, AccessOK: true, CreatedAt: "2026-03-05T00:00:00Z",
		}, nil)
	// Async scan goroutine — use Maybe()
	mocks.scanner.On("ScanStructure", mock.Anything, structureID, "access-token-456").Return(true, nil).Maybe()

	body, _ := json.Marshal(map[string]int64{"structureId": structureID, "characterId": charID})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	created := result.(*models.UserTradingStructure)
	assert.Equal(t, structureID, created.StructureID)
	assert.Equal(t, "New Structure", created.Name)
	mocks.characters.AssertExpectations(t)
	mocks.esi.AssertExpectations(t)
	mocks.structures.AssertExpectations(t)
}

func Test_TradingStructures_AddStructure_AccessDenied(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001002)
	structureID := int64(9876543210)

	char := &repositories.Character{ID: charID, EsiRefreshToken: "refresh-token-789"}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.esi.On("RefreshAccessToken", mock.Anything, "refresh-token-789").Return(
		&client.RefreshedToken{AccessToken: "access-token-bad"}, nil)
	// ESI returns nil (403 — no access)
	mocks.esi.On("GetStructureInfo", mock.Anything, structureID, "access-token-bad").Return(nil, nil)

	body, _ := json.Marshal(map[string]int64{"structureId": structureID, "characterId": charID})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, false, resp["accessOk"])
	mocks.esi.AssertExpectations(t)
}

func Test_TradingStructures_AddStructure_MissingStructureID(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]int64{"characterId": int64(200001001)})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TradingStructures_AddStructure_MissingCharacterID(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]int64{"structureId": int64(1234567890)})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TradingStructures_AddStructure_InvalidBody(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader([]byte("not json")))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TradingStructures_AddStructure_CharacterNotFound(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{}, nil)

	body, _ := json.Marshal(map[string]int64{"structureId": int64(1234567890), "characterId": int64(200001003)})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_TradingStructures_AddStructure_GetCharactersError(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.characters.On("GetAll", mock.Anything, userID).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(map[string]int64{"structureId": int64(1234567890), "characterId": int64(200001004)})
	req := httptest.NewRequest("POST", "/v1/hauling/structures", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.AddStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: DeleteStructure ---

func Test_TradingStructures_DeleteStructure_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.structures.On("Delete", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/hauling/structures/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.DeleteStructure(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.structures.AssertExpectations(t)
}

func Test_TradingStructures_DeleteStructure_NotFound(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.structures.On("Delete", mock.Anything, int64(5), userID).Return(errors.New("user trading structure not found"))

	req := httptest.NewRequest("DELETE", "/v1/hauling/structures/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.DeleteStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TradingStructures_DeleteStructure_InvalidID(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	req := httptest.NewRequest("DELETE", "/v1/hauling/structures/abc", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.DeleteStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Tests: ScanStructure ---

func Test_TradingStructures_ScanStructure_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001010)
	structureID := int64(1234567890)

	structure := &models.UserTradingStructure{ID: int64(5), UserID: userID, StructureID: structureID, CharacterID: charID, AccessOK: true}
	char := &repositories.Character{ID: charID, EsiRefreshToken: "refresh-abc"}

	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{structure}, nil)
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.esi.On("RefreshAccessToken", mock.Anything, "refresh-abc").Return(
		&client.RefreshedToken{AccessToken: "access-xyz"}, nil)
	mocks.scanner.On("ScanStructure", mock.Anything, structureID, "access-xyz").Return(true, nil)

	req := httptest.NewRequest("POST", "/v1/hauling/structures/5/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, true, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

func Test_TradingStructures_ScanStructure_AccessDenied(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001011)
	structureID := int64(9876543210)

	structure := &models.UserTradingStructure{ID: int64(5), UserID: userID, StructureID: structureID, CharacterID: charID, AccessOK: true}
	char := &repositories.Character{ID: charID, EsiRefreshToken: "refresh-def"}

	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{structure}, nil)
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.esi.On("RefreshAccessToken", mock.Anything, "refresh-def").Return(
		&client.RefreshedToken{AccessToken: "access-denied"}, nil)
	mocks.scanner.On("ScanStructure", mock.Anything, structureID, "access-denied").Return(false, nil)
	// UpdateAccessStatus is called when access is denied
	mocks.structures.On("UpdateAccessStatus", mock.Anything, userID, structureID, false).Return(nil)

	req := httptest.NewRequest("POST", "/v1/hauling/structures/5/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, false, resp["accessOk"])
}

func Test_TradingStructures_ScanStructure_StructureNotFound(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{}, nil)

	req := httptest.NewRequest("POST", "/v1/hauling/structures/99/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "99"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_TradingStructures_ScanStructure_CharacterNotFound(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001012)
	structureID := int64(1111111111)

	structure := &models.UserTradingStructure{ID: int64(5), UserID: userID, StructureID: structureID, CharacterID: charID}
	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{structure}, nil)
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{}, nil)

	req := httptest.NewRequest("POST", "/v1/hauling/structures/5/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_TradingStructures_ScanStructure_InvalidID(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	req := httptest.NewRequest("POST", "/v1/hauling/structures/abc/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TradingStructures_ScanStructure_ScanError(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001013)
	structureID := int64(2222222222)

	structure := &models.UserTradingStructure{ID: int64(5), UserID: userID, StructureID: structureID, CharacterID: charID}
	char := &repositories.Character{ID: charID, EsiRefreshToken: "refresh-ghi"}

	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{structure}, nil)
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.esi.On("RefreshAccessToken", mock.Anything, "refresh-ghi").Return(
		&client.RefreshedToken{AccessToken: "access-error"}, nil)
	mocks.scanner.On("ScanStructure", mock.Anything, structureID, "access-error").Return(false, errors.New("ESI error"))

	req := httptest.NewRequest("POST", "/v1/hauling/structures/5/scan", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.ScanStructure(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: ListCharacterAssetStructures ---

func Test_TradingStructures_ListCharacterAssetStructures_Success(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001020)
	structureID1 := int64(1000000001)
	structureID2 := int64(1000000002)

	char := &repositories.Character{ID: charID}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.assets.On("GetPlayerOwnedStationIDs", mock.Anything, charID, userID).Return([]int64{structureID1, structureID2}, nil)
	// structureID1 is already saved with a name; structureID2 is unknown
	savedStructures := []*models.UserTradingStructure{
		{ID: int64(10), UserID: userID, StructureID: structureID1, Name: "Known Structure"},
	}
	mocks.structures.On("List", mock.Anything, userID).Return(savedStructures, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001020/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001020"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	type assetStructure struct {
		StructureID int64  `json:"structureId"`
		Name        string `json:"name"`
	}
	// Re-encode and decode to inspect fields
	raw, _ := json.Marshal(result)
	var items []assetStructure
	_ = json.Unmarshal(raw, &items)
	assert.Len(t, items, 2)

	found1, found2 := false, false
	for _, item := range items {
		if item.StructureID == structureID1 {
			assert.Equal(t, "Known Structure", item.Name)
			found1 = true
		}
		if item.StructureID == structureID2 {
			assert.Equal(t, "", item.Name)
			found2 = true
		}
	}
	assert.True(t, found1, "structureID1 should be in results")
	assert.True(t, found2, "structureID2 should be in results")

	mocks.characters.AssertExpectations(t)
	mocks.assets.AssertExpectations(t)
	mocks.structures.AssertExpectations(t)
}

func Test_TradingStructures_ListCharacterAssetStructures_NoAssets(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001021)

	char := &repositories.Character{ID: charID}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.assets.On("GetPlayerOwnedStationIDs", mock.Anything, charID, userID).Return([]int64{}, nil)
	mocks.structures.On("List", mock.Anything, userID).Return([]*models.UserTradingStructure{}, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001021/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001021"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	raw, _ := json.Marshal(result)
	assert.Equal(t, "[]", string(raw))
}

func Test_TradingStructures_ListCharacterAssetStructures_CharacterNotFound(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{}, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001022/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001022"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	mocks.characters.AssertExpectations(t)
}

func Test_TradingStructures_ListCharacterAssetStructures_InvalidID(t *testing.T) {
	controller, _ := setupTradingStructuresController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/characters/abc/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TradingStructures_ListCharacterAssetStructures_GetCharactersError(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)

	mocks.characters.On("GetAll", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001023/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001023"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TradingStructures_ListCharacterAssetStructures_AssetsError(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001024)

	char := &repositories.Character{ID: charID}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.assets.On("GetPlayerOwnedStationIDs", mock.Anything, charID, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001024/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001024"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TradingStructures_ListCharacterAssetStructures_ListStructuresError(t *testing.T) {
	controller, mocks := setupTradingStructuresController()
	userID := int64(100)
	charID := int64(200001025)

	char := &repositories.Character{ID: charID}
	mocks.characters.On("GetAll", mock.Anything, userID).Return([]*repositories.Character{char}, nil)
	mocks.assets.On("GetPlayerOwnedStationIDs", mock.Anything, charID, userID).Return([]int64{int64(1000000010)}, nil)
	mocks.structures.On("List", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/characters/200001025/asset-structures", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "200001025"}}

	result, httpErr := controller.ListCharacterAssetStructures(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}
