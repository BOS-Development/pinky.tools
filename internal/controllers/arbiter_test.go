package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/services"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock implementations ---

type MockArbiterSettingsRepository struct {
	mock.Mock
}

func (m *MockArbiterSettingsRepository) GetArbiterSettings(ctx context.Context, userID int64) (*models.ArbiterSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArbiterSettings), args.Error(1)
}

func (m *MockArbiterSettingsRepository) UpsertArbiterSettings(ctx context.Context, settings *models.ArbiterSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *MockArbiterSettingsRepository) GetArbiterEnabled(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type MockArbiterScanRepository struct {
	mock.Mock
}

func (m *MockArbiterScanRepository) GetT2BlueprintsForScan(ctx context.Context) ([]*models.T2BlueprintScanItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.T2BlueprintScanItem), args.Error(1)
}

func (m *MockArbiterScanRepository) GetDecryptors(ctx context.Context) ([]*models.Decryptor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Decryptor), args.Error(1)
}

func (m *MockArbiterScanRepository) GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.BlueprintMaterial), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BlueprintProduct), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBestInventionCharacter(ctx context.Context, userID int64, blueprintTypeID int64) (*models.InventionCharacter, error) {
	args := m.Called(ctx, userID, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventionCharacter), args.Error(1)
}

func (m *MockArbiterScanRepository) GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error) {
	args := m.Called(ctx, systemID, activity)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetDemandStats(ctx context.Context, typeIDs []int64) (map[int64]*models.DemandStats, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return map[int64]*models.DemandStats{}, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.DemandStats), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetMarketPricesLastUpdated(ctx context.Context) (*time.Time, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

// Ensure MockArbiterScanRepository satisfies services.ArbiterScanRepository
var _ services.ArbiterScanRepository = &MockArbiterScanRepository{}

// --- Mock for scopes ---

type MockArbiterScopesRepository struct {
	mock.Mock
}

func (m *MockArbiterScopesRepository) GetScopes(ctx context.Context, userID int64) ([]*models.ArbiterScope, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ArbiterScope), args.Error(1)
}

func (m *MockArbiterScopesRepository) GetScope(ctx context.Context, scopeID, userID int64) (*models.ArbiterScope, error) {
	args := m.Called(ctx, scopeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArbiterScope), args.Error(1)
}

func (m *MockArbiterScopesRepository) CreateScope(ctx context.Context, scope *models.ArbiterScope) (int64, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScopesRepository) UpdateScope(ctx context.Context, scope *models.ArbiterScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockArbiterScopesRepository) DeleteScope(ctx context.Context, scopeID, userID int64) error {
	args := m.Called(ctx, scopeID, userID)
	return args.Error(0)
}

func (m *MockArbiterScopesRepository) GetScopeMembers(ctx context.Context, scopeID int64) ([]*models.ArbiterScopeMember, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ArbiterScopeMember), args.Error(1)
}

func (m *MockArbiterScopesRepository) AddScopeMember(ctx context.Context, member *models.ArbiterScopeMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockArbiterScopesRepository) RemoveScopeMember(ctx context.Context, memberID, scopeID int64) error {
	args := m.Called(ctx, memberID, scopeID)
	return args.Error(0)
}

// --- Mock for tax profile ---

type MockArbiterTaxProfileRepository struct {
	mock.Mock
}

func (m *MockArbiterTaxProfileRepository) GetTaxProfile(ctx context.Context, userID int64) (*models.ArbiterTaxProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArbiterTaxProfile), args.Error(1)
}

func (m *MockArbiterTaxProfileRepository) UpsertTaxProfile(ctx context.Context, profile *models.ArbiterTaxProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

// --- Mock for lists ---

type MockArbiterListsRepository struct {
	mock.Mock
}

func (m *MockArbiterListsRepository) GetBlacklist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ArbiterListItem), args.Error(1)
}

func (m *MockArbiterListsRepository) AddToBlacklist(ctx context.Context, userID, typeID int64) error {
	args := m.Called(ctx, userID, typeID)
	return args.Error(0)
}

func (m *MockArbiterListsRepository) RemoveFromBlacklist(ctx context.Context, userID, typeID int64) error {
	args := m.Called(ctx, userID, typeID)
	return args.Error(0)
}

func (m *MockArbiterListsRepository) GetWhitelist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ArbiterListItem), args.Error(1)
}

func (m *MockArbiterListsRepository) AddToWhitelist(ctx context.Context, userID, typeID int64) error {
	args := m.Called(ctx, userID, typeID)
	return args.Error(0)
}

func (m *MockArbiterListsRepository) RemoveFromWhitelist(ctx context.Context, userID, typeID int64) error {
	args := m.Called(ctx, userID, typeID)
	return args.Error(0)
}

// --- Mock for BOM ---

type MockArbiterBOMRepository struct {
	mock.Mock
}

func (m *MockArbiterBOMRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.BlueprintMaterial), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BlueprintProduct), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error) {
	args := m.Called(ctx, systemID, activity)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetScopeAssets(ctx context.Context, scopeID, userID int64) (map[int64]int64, error) {
	args := m.Called(ctx, scopeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]int64), args.Error(1)
}

// --- Mock for solar systems ---

type MockArbiterSolarSystemRepository struct {
	mock.Mock
}

func (m *MockArbiterSolarSystemRepository) SearchSolarSystems(ctx context.Context, query string, limit int) ([]*models.SolarSystemSearchResult, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SolarSystemSearchResult), args.Error(1)
}

// --- Mock for decryptors ---

type MockArbiterDecryptorRepository struct {
	mock.Mock
}

func (m *MockArbiterDecryptorRepository) GetDecryptors(ctx context.Context) ([]*models.Decryptor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Decryptor), args.Error(1)
}

// --- Setup helpers ---

type arbiterMocks struct {
	settings   *MockArbiterSettingsRepository
	scan       *MockArbiterScanRepository
	scopes     *MockArbiterScopesRepository
	tax        *MockArbiterTaxProfileRepository
	lists      *MockArbiterListsRepository
	bom        *MockArbiterBOMRepository
	solarSys   *MockArbiterSolarSystemRepository
	decryptors *MockArbiterDecryptorRepository
}

func setupArbiterController() (*controllers.Arbiter, *arbiterMocks) {
	mocks := &arbiterMocks{
		settings: &MockArbiterSettingsRepository{},
		scan:     &MockArbiterScanRepository{},
	}
	c := controllers.NewArbiter(&MockRouter{}, mocks.settings, mocks.scan)
	return c, mocks
}

func setupArbiterControllerFull() (*controllers.Arbiter, *arbiterMocks) {
	mocks := &arbiterMocks{
		settings:   &MockArbiterSettingsRepository{},
		scan:       &MockArbiterScanRepository{},
		scopes:     &MockArbiterScopesRepository{},
		tax:        &MockArbiterTaxProfileRepository{},
		lists:      &MockArbiterListsRepository{},
		bom:        &MockArbiterBOMRepository{},
		solarSys:   &MockArbiterSolarSystemRepository{},
		decryptors: &MockArbiterDecryptorRepository{},
	}
	c := controllers.NewArbiterFull(
		&MockRouter{},
		mocks.settings,
		mocks.scan,
		mocks.scopes,
		mocks.tax,
		mocks.lists,
		mocks.bom,
		mocks.solarSys,
		mocks.decryptors,
	)
	return c, mocks
}

// --- Tests ---

func Test_ArbiterGetSettings_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/settings", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterSettings(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
}

func Test_ArbiterGetSettings_Returns500_WhenFeatureCheckFails(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/arbiter/settings", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterSettings(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_ArbiterGetSettings_Returns200_WithDefaults_WhenEnabled(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	defaultSettings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		UseWhitelist:       true,
		UseBlacklist:       true,
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(defaultSettings, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/settings", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterSettings(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	s, ok := result.(*models.ArbiterSettings)
	assert.True(t, ok)
	assert.Equal(t, "athanor", s.ReactionStructure)
	assert.Equal(t, "raitaru", s.FinalStructure)

	mocks.settings.AssertExpectations(t)
}

func Test_ArbiterUpdateSettings_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	body, _ := json.Marshal(map[string]any{
		"reaction_structure":  "athanor",
		"reaction_rig":        "t1",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/settings", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateArbiterSettings(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterUpdateSettings_Returns400_WhenInvalidStructure(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]any{
		"reaction_structure":  "invalid_structure",
		"reaction_rig":        "t1",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/settings", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateArbiterSettings(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ArbiterUpdateSettings_Returns400_WhenInvalidRig(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]any{
		"reaction_structure":  "athanor",
		"reaction_rig":        "bad_rig",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/settings", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateArbiterSettings(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ArbiterUpdateSettings_Returns200_WhenValid(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("UpsertArbiterSettings", mock.Anything, mock.AnythingOfType("*models.ArbiterSettings")).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"reaction_structure":  "tatara",
		"reaction_rig":        "t2",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"component_structure": "azbel",
		"component_rig":       "t2",
		"final_structure":     "sotiyo",
		"final_rig":           "t2",
		"use_whitelist":       true,
		"use_blacklist":       false,
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/settings", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateArbiterSettings(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	s, ok := result.(*models.ArbiterSettings)
	assert.True(t, ok)
	assert.Equal(t, "tatara", s.ReactionStructure)
	assert.Equal(t, "sotiyo", s.FinalStructure)

	mocks.settings.AssertExpectations(t)
}

func Test_ArbiterUpdateSettings_AssignsFacilityTaxFields(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("UpsertArbiterSettings", mock.Anything, mock.AnythingOfType("*models.ArbiterSettings")).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"reaction_structure":    "tatara",
		"reaction_rig":          "t2",
		"invention_structure":   "raitaru",
		"invention_rig":         "t1",
		"component_structure":   "azbel",
		"component_rig":         "t2",
		"final_structure":       "sotiyo",
		"final_rig":             "t2",
		"final_facility_tax":    1.5,
		"component_facility_tax": 0.5,
		"reaction_facility_tax": 2.0,
		"invention_facility_tax": 0.25,
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/settings", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateArbiterSettings(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	s, ok := result.(*models.ArbiterSettings)
	assert.True(t, ok)
	assert.Equal(t, 1.5, s.FinalFacilityTax)
	assert.Equal(t, 0.5, s.ComponentFacilityTax)
	assert.Equal(t, 2.0, s.ReactionFacilityTax)
	assert.Equal(t, 0.25, s.InventionFacilityTax)

	mocks.settings.AssertExpectations(t)
}

func Test_ArbiterGetOpportunities_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/opportunities", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterOpportunities(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterGetOpportunities_Returns200_WithEmptyResults(t *testing.T) {
	c, mocks := setupArbiterController()
	userID := int64(100)

	defaultSettings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(defaultSettings, nil)
	mocks.scan.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{}, nil)
	mocks.scan.On("GetMarketPricesLastUpdated", mock.Anything).Return((*time.Time)(nil), nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/opportunities", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterOpportunities(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	scanResult, ok := result.(*models.ArbiterScanResult)
	assert.True(t, ok)
	assert.Equal(t, 0, scanResult.TotalScanned)
	assert.NotNil(t, scanResult.Opportunities)
	assert.True(t, scanResult.GeneratedAt.Before(time.Now().Add(time.Second)))

	mocks.settings.AssertExpectations(t)
	mocks.scan.AssertExpectations(t)
}

func Test_ArbiterGetOpportunities_LoadsTaxProfile_WhenTaxRepoSet(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	defaultSettings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
	}
	taxProfile := &models.ArbiterTaxProfile{
		UserID:          userID,
		SalesTaxRate:    3.6,
		BrokerFeeRate:   1.0,
		InputPriceType:  "sell",
		OutputPriceType: "sell",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(defaultSettings, nil)
	mocks.tax.On("GetTaxProfile", mock.Anything, userID).Return(taxProfile, nil)
	mocks.scan.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{}, nil)
	mocks.scan.On("GetMarketPricesLastUpdated", mock.Anything).Return((*time.Time)(nil), nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/opportunities", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterOpportunities(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	scanResult, ok := result.(*models.ArbiterScanResult)
	assert.True(t, ok)
	assert.Equal(t, 0, scanResult.TotalScanned)

	mocks.settings.AssertExpectations(t)
	mocks.tax.AssertExpectations(t)
	mocks.scan.AssertExpectations(t)
}

func Test_ArbiterGetOpportunities_Returns500_WhenTaxProfileFails(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	defaultSettings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(defaultSettings, nil)
	mocks.tax.On("GetTaxProfile", mock.Anything, userID).Return((*models.ArbiterTaxProfile)(nil), errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/arbiter/opportunities", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetArbiterOpportunities(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
	mocks.tax.AssertExpectations(t)
}

// --- Scopes ---

func Test_ArbiterGetScopes_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/scopes", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetScopes(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterGetScopes_Returns200_WithScopes(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	scopes := []*models.ArbiterScope{
		{ID: 1, UserID: userID, Name: "Main Scope", IsDefault: true},
	}
	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.scopes.On("GetScopes", mock.Anything, userID).Return(scopes, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/scopes", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetScopes(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	ss, ok := result.([]*models.ArbiterScope)
	assert.True(t, ok)
	assert.Len(t, ss, 1)
	assert.Equal(t, "Main Scope", ss[0].Name)

	mocks.settings.AssertExpectations(t)
	mocks.scopes.AssertExpectations(t)
}

func Test_ArbiterCreateScope_Returns400_WhenNameEmpty(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]any{"name": ""})
	req := httptest.NewRequest("POST", "/v1/arbiter/scopes", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.CreateScope(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ArbiterCreateScope_Returns200_WhenValid(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.scopes.On("CreateScope", mock.Anything, mock.AnythingOfType("*models.ArbiterScope")).Return(int64(42), nil)

	body, _ := json.Marshal(map[string]any{"name": "My Scope", "is_default": true})
	req := httptest.NewRequest("POST", "/v1/arbiter/scopes", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.CreateScope(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	s, ok := result.(*models.ArbiterScope)
	assert.True(t, ok)
	assert.Equal(t, int64(42), s.ID)
	assert.Equal(t, "My Scope", s.Name)

	mocks.settings.AssertExpectations(t)
	mocks.scopes.AssertExpectations(t)
}

func Test_ArbiterDeleteScope_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/scopes/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := c.DeleteScope(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterDeleteScope_Returns200_WhenValid(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.scopes.On("DeleteScope", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/scopes/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := c.DeleteScope(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mocks.settings.AssertExpectations(t)
	mocks.scopes.AssertExpectations(t)
}

// --- Tax Profile ---

func Test_ArbiterGetTaxProfile_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/tax-profile", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetTaxProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterGetTaxProfile_Returns200(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	profile := &models.ArbiterTaxProfile{
		UserID:          userID,
		SalesTaxRate:    0.036,
		BrokerFeeRate:   0.03,
		InputPriceType:  "sell",
		OutputPriceType: "buy",
	}
	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.tax.On("GetTaxProfile", mock.Anything, userID).Return(profile, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/tax-profile", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetTaxProfile(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	p, ok := result.(*models.ArbiterTaxProfile)
	assert.True(t, ok)
	assert.InDelta(t, 0.036, p.SalesTaxRate, 0.001)

	mocks.settings.AssertExpectations(t)
	mocks.tax.AssertExpectations(t)
}

func Test_ArbiterUpdateTaxProfile_Returns400_WhenInvalidPriceType(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]any{
		"sales_tax_rate":   0.036,
		"broker_fee_rate":  0.03,
		"input_price_type": "invalid",
		"output_price_type": "buy",
	})
	req := httptest.NewRequest("PUT", "/v1/arbiter/tax-profile", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.UpdateTaxProfile(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Blacklist ---

func Test_ArbiterGetBlacklist_Returns403_WhenNotEnabled(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(false, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/blacklist", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetBlacklist(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)
}

func Test_ArbiterGetBlacklist_Returns200(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	items := []*models.ArbiterListItem{
		{UserID: userID, TypeID: 101, Name: "Bad Item", AddedAt: time.Now()},
	}
	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.lists.On("GetBlacklist", mock.Anything, userID).Return(items, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/blacklist", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetBlacklist(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	bl, ok := result.([]*models.ArbiterListItem)
	assert.True(t, ok)
	assert.Len(t, bl, 1)
	assert.Equal(t, int64(101), bl[0].TypeID)

	mocks.settings.AssertExpectations(t)
	mocks.lists.AssertExpectations(t)
}

func Test_ArbiterAddToBlacklist_Returns400_WhenMissingTypeID(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	body, _ := json.Marshal(map[string]any{"type_id": 0})
	req := httptest.NewRequest("POST", "/v1/arbiter/blacklist", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.AddToBlacklist(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ArbiterAddToBlacklist_Returns200_WhenValid(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.lists.On("AddToBlacklist", mock.Anything, userID, int64(555)).Return(nil)

	body, _ := json.Marshal(map[string]any{"type_id": 555})
	req := httptest.NewRequest("POST", "/v1/arbiter/blacklist", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.AddToBlacklist(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mocks.settings.AssertExpectations(t)
	mocks.lists.AssertExpectations(t)
}

func Test_ArbiterRemoveFromBlacklist_Returns200_WhenValid(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.lists.On("RemoveFromBlacklist", mock.Anything, userID, int64(555)).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/arbiter/blacklist/555", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"typeID": "555"}}

	result, httpErr := c.RemoveFromBlacklist(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mocks.settings.AssertExpectations(t)
	mocks.lists.AssertExpectations(t)
}

// --- Whitelist ---

func Test_ArbiterGetWhitelist_Returns200(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	items := []*models.ArbiterListItem{
		{UserID: userID, TypeID: 202, Name: "Good Item", AddedAt: time.Now()},
	}
	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.lists.On("GetWhitelist", mock.Anything, userID).Return(items, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/whitelist", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetWhitelist(args)
	assert.Nil(t, httpErr)

	wl, ok := result.([]*models.ArbiterListItem)
	assert.True(t, ok)
	assert.Len(t, wl, 1)
	assert.Equal(t, int64(202), wl[0].TypeID)

	mocks.settings.AssertExpectations(t)
	mocks.lists.AssertExpectations(t)
}

// --- Solar System Search ---

func Test_ArbiterSearchSolarSystems_Returns200_WithResults(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	results := []*models.SolarSystemSearchResult{
		{SolarSystemID: 30000142, Name: "Jita", SecurityClass: "high", Security: 0.9},
	}
	mocks.solarSys.On("SearchSolarSystems", mock.Anything, "Jit", 10).Return(results, nil)

	req := httptest.NewRequest("GET", "/v1/solar-systems/search?q=Jit&limit=10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.SearchSolarSystems(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	ss, ok := result.([]*models.SolarSystemSearchResult)
	assert.True(t, ok)
	assert.Len(t, ss, 1)
	assert.Equal(t, "Jita", ss[0].Name)
	assert.Equal(t, "high", ss[0].SecurityClass)

	mocks.solarSys.AssertExpectations(t)
}

func Test_ArbiterSearchSolarSystems_Returns200_EmptyWhenNoQuery(t *testing.T) {
	c, _ := setupArbiterControllerFull()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/solar-systems/search", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.SearchSolarSystems(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	ss, ok := result.([]*models.SolarSystemSearchResult)
	assert.True(t, ok)
	assert.Empty(t, ss)
}

// --- Scope Members ---

func Test_ArbiterGetScopeMembers_Returns404_WhenScopeNotFound(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.scopes.On("GetScope", mock.Anything, int64(99), userID).Return((*models.ArbiterScope)(nil), nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/scopes/99/members", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "99"}}

	result, httpErr := c.GetScopeMembers(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
	mocks.scopes.AssertExpectations(t)
}

func Test_ArbiterAddScopeMember_Returns400_WhenInvalidMemberType(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	scope := &models.ArbiterScope{ID: 1, UserID: userID, Name: "Test"}
	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.scopes.On("GetScope", mock.Anything, int64(1), userID).Return(scope, nil)

	body, _ := json.Marshal(map[string]any{"member_type": "alliance", "member_id": 12345})
	req := httptest.NewRequest("POST", "/v1/arbiter/scopes/1/members", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := c.AddScopeMember(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
	mocks.scopes.AssertExpectations(t)
}

// --- BOM Tree ---

func Test_GetBOMTree_Returns404_WhenNoBlueprintFound(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	settings := &models.ArbiterSettings{
		UserID:         userID,
		FinalStructure: "raitaru",
		FinalRig:       "t2",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(settings, nil)
	mocks.lists.On("GetBlacklist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.lists.On("GetWhitelist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.bom.On("GetBlueprintForProduct", mock.Anything, int64(9001)).Return(int64(0), nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/bom/9001", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"typeID": "9001"}}

	result, httpErr := c.GetBOMTree(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
	mocks.bom.AssertExpectations(t)
}

func Test_GetBOMTree_Returns400_WhenInvalidTypeID(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/bom/notanumber", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"typeID": "notanumber"}}

	result, httpErr := c.GetBOMTree(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)

	mocks.settings.AssertExpectations(t)
}

func Test_GetBOMTree_LoadsScopeAssets_WhenScopeIDProvided(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	settings := &models.ArbiterSettings{
		UserID:         userID,
		FinalStructure: "raitaru",
		FinalRig:       "t2",
	}
	scopeAssets := map[int64]int64{9002: 5}

	sellPrice := float64(1_000_000)
	prices := map[int64]*models.MarketPrice{
		9002: {TypeID: 9002, SellPrice: &sellPrice},
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(settings, nil)
	mocks.lists.On("GetBlacklist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.lists.On("GetWhitelist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.bom.On("GetScopeAssets", mock.Anything, int64(42), userID).Return(scopeAssets, nil)
	mocks.bom.On("GetBlueprintForProduct", mock.Anything, int64(9002)).Return(int64(8002), nil)
	mocks.bom.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(prices, nil)
	mocks.bom.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	mocks.bom.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8002), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	mocks.bom.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8002), "reaction").Return([]*models.BlueprintMaterial{}, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/bom/9002?scope_id=42", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"typeID": "9002"}}

	result, httpErr := c.GetBOMTree(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	node, ok := result.(*models.BOMNode)
	assert.True(t, ok)
	// Scope has 5 units of this type — should appear in Available
	assert.Equal(t, int64(5), node.Available)

	mocks.settings.AssertExpectations(t)
	mocks.bom.AssertExpectations(t)
}

func Test_GetBOMTree_BuildAll_ForcesBuiltDecision(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	settings := &models.ArbiterSettings{
		UserID:         userID,
		FinalStructure: "raitaru",
		FinalRig:       "t2",
	}

	// Product costs 1M to buy, material costs 8M — normally "buy"
	productSell := float64(1_000_000)
	matSell := float64(4_000_000)
	prices := map[int64]*models.MarketPrice{
		9003: {TypeID: 9003, SellPrice: &productSell},
		9004: {TypeID: 9004, SellPrice: &matSell},
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(settings, nil)
	mocks.lists.On("GetBlacklist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.lists.On("GetWhitelist", mock.Anything, userID).Return([]*models.ArbiterListItem{}, nil)
	mocks.bom.On("GetBlueprintForProduct", mock.Anything, int64(9003)).Return(int64(8003), nil)
	mocks.bom.On("GetBlueprintForProduct", mock.Anything, int64(9004)).Return(int64(0), nil)
	mocks.bom.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(prices, nil)
	mocks.bom.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	mocks.bom.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8003), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 9004, TypeName: "Expensive Part", Quantity: 2},
		}, nil)
	mocks.bom.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	mocks.bom.On("GetReactionBlueprintForProduct", mock.Anything, int64(9004)).Return(int64(0), nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/bom/9003?build_all=true", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"typeID": "9003"}}

	result, httpErr := c.GetBOMTree(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	node, ok := result.(*models.BOMNode)
	assert.True(t, ok)
	// build_all forces "build_override" even though buying would be cheaper
	assert.Equal(t, "build_override", node.Decision)

	mocks.settings.AssertExpectations(t)
	mocks.bom.AssertExpectations(t)
}

// --- GetDecryptors tests ---

func Test_GetDecryptors_Returns200_WithDecryptors(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	decryptors := []*models.Decryptor{
		{TypeID: 34201, Name: "Accelerant Decryptor", MEModifier: 2, TEModifier: -1, RunModifier: 1, ProbabilityMultiplier: 1.2},
		{TypeID: 34202, Name: "Attainment Decryptor", MEModifier: 4, TEModifier: -2, RunModifier: -1, ProbabilityMultiplier: 1.8},
	}
	mocks.decryptors.On("GetDecryptors", mock.Anything).Return(decryptors, nil)

	req := httptest.NewRequest("GET", "/v1/arbiter/decryptors", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetDecryptors(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	returned, ok := result.([]*models.Decryptor)
	assert.True(t, ok)
	assert.Len(t, returned, 2)
	assert.Equal(t, int64(34201), returned[0].TypeID)
	assert.Equal(t, "Accelerant Decryptor", returned[0].Name)

	mocks.decryptors.AssertExpectations(t)
}

func Test_GetDecryptors_Returns500_WhenRepoFails(t *testing.T) {
	c, mocks := setupArbiterControllerFull()
	userID := int64(100)

	mocks.decryptors.On("GetDecryptors", mock.Anything).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/arbiter/decryptors", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := c.GetDecryptors(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mocks.decryptors.AssertExpectations(t)
}
