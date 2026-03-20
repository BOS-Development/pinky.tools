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

// Ensure MockArbiterScanRepository satisfies services.ArbiterScanRepository
var _ services.ArbiterScanRepository = &MockArbiterScanRepository{}

// --- Setup helper ---

type arbiterMocks struct {
	settings *MockArbiterSettingsRepository
	scan     *MockArbiterScanRepository
}

func setupArbiterController() (*controllers.Arbiter, *arbiterMocks) {
	mocks := &arbiterMocks{
		settings: &MockArbiterSettingsRepository{},
		scan:     &MockArbiterScanRepository{},
	}
	c := controllers.NewArbiter(&MockRouter{}, mocks.settings, mocks.scan)
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
		ReactionSecurity:   "null",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSecurity:  "high",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		ComponentSecurity:  "null",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		FinalSecurity:      "null",
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

	body, _ := json.Marshal(map[string]string{
		"reaction_structure":  "athanor",
		"reaction_rig":        "t1",
		"reaction_security":   "null",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"invention_security":  "high",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"component_security":  "null",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
		"final_security":      "null",
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

	body, _ := json.Marshal(map[string]string{
		"reaction_structure":  "invalid_structure",
		"reaction_rig":        "t1",
		"reaction_security":   "null",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"invention_security":  "high",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"component_security":  "null",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
		"final_security":      "null",
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

	body, _ := json.Marshal(map[string]string{
		"reaction_structure":  "athanor",
		"reaction_rig":        "bad_rig",
		"reaction_security":   "null",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"invention_security":  "high",
		"component_structure": "raitaru",
		"component_rig":       "t2",
		"component_security":  "null",
		"final_structure":     "raitaru",
		"final_rig":           "t2",
		"final_security":      "null",
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

	body, _ := json.Marshal(map[string]string{
		"reaction_structure":  "tatara",
		"reaction_rig":        "t2",
		"reaction_security":   "null",
		"invention_structure": "raitaru",
		"invention_rig":       "t1",
		"invention_security":  "high",
		"component_structure": "azbel",
		"component_rig":       "t2",
		"component_security":  "low",
		"final_structure":     "sotiyo",
		"final_rig":           "t2",
		"final_security":      "null",
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
		ReactionSecurity:   "null",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSecurity:  "high",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		ComponentSecurity:  "null",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		FinalSecurity:      "null",
	}

	mocks.settings.On("GetArbiterEnabled", mock.Anything, userID).Return(true, nil)
	mocks.settings.On("GetArbiterSettings", mock.Anything, userID).Return(defaultSettings, nil)
	mocks.scan.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{}, nil)

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
