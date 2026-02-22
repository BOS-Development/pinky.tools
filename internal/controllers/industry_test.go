package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock repositories ---

type MockIndustryJobsRepository struct {
	mock.Mock
}

func (m *MockIndustryJobsRepository) GetActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.IndustryJob), args.Error(1)
}

func (m *MockIndustryJobsRepository) GetAllJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.IndustryJob), args.Error(1)
}

type MockIndustryJobQueueRepository struct {
	mock.Mock
}

func (m *MockIndustryJobQueueRepository) Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryJobQueueEntry), args.Error(1)
}

func (m *MockIndustryJobQueueRepository) GetByUser(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.IndustryJobQueueEntry), args.Error(1)
}

func (m *MockIndustryJobQueueRepository) Update(ctx context.Context, id, userID int64, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	args := m.Called(ctx, id, userID, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryJobQueueEntry), args.Error(1)
}

func (m *MockIndustryJobQueueRepository) Cancel(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

type MockIndustrySDERepository struct {
	mock.Mock
}

func (m *MockIndustrySDERepository) GetManufacturingBlueprint(ctx context.Context, blueprintTypeID int64) (*repositories.ManufacturingBlueprintRow, error) {
	args := m.Called(ctx, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ManufacturingBlueprintRow), args.Error(1)
}

func (m *MockIndustrySDERepository) GetManufacturingMaterials(ctx context.Context, blueprintTypeID int64) ([]*repositories.ManufacturingMaterialRow, error) {
	args := m.Called(ctx, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.ManufacturingMaterialRow), args.Error(1)
}

func (m *MockIndustrySDERepository) SearchBlueprints(ctx context.Context, query string, activity string, limit int) ([]*repositories.BlueprintSearchRow, error) {
	args := m.Called(ctx, query, activity, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.BlueprintSearchRow), args.Error(1)
}

func (m *MockIndustrySDERepository) GetManufacturingSystems(ctx context.Context) ([]*models.ReactionSystem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ReactionSystem), args.Error(1)
}

type MockIndustryMarketRepository struct {
	mock.Mock
}

func (m *MockIndustryMarketRepository) GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockIndustryMarketRepository) GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

type MockIndustryCostIndicesRepository struct {
	mock.Mock
}

func (m *MockIndustryCostIndicesRepository) GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error) {
	args := m.Called(ctx, systemID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryCostIndex), args.Error(1)
}

// --- Helper to create controller with mocks ---

type industryMocks struct {
	jobsRepo        *MockIndustryJobsRepository
	queueRepo       *MockIndustryJobQueueRepository
	sdeRepo         *MockIndustrySDERepository
	marketRepo      *MockIndustryMarketRepository
	costIndicesRepo *MockIndustryCostIndicesRepository
}

func setupIndustryController() (*controllers.Industry, *industryMocks) {
	mocks := &industryMocks{
		jobsRepo:        new(MockIndustryJobsRepository),
		queueRepo:       new(MockIndustryJobQueueRepository),
		sdeRepo:         new(MockIndustrySDERepository),
		marketRepo:      new(MockIndustryMarketRepository),
		costIndicesRepo: new(MockIndustryCostIndicesRepository),
	}

	controller := controllers.NewIndustry(
		&MockRouter{},
		mocks.jobsRepo,
		mocks.queueRepo,
		mocks.sdeRepo,
		mocks.marketRepo,
		mocks.costIndicesRepo,
	)

	return controller, mocks
}

// --- GetActiveJobs Tests ---

func Test_IndustryController_GetActiveJobs_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	cost := 1500000.0
	expectedJobs := []*models.IndustryJob{
		{
			JobID:           10001,
			InstallerID:     1001,
			UserID:          100,
			ActivityID:      1,
			BlueprintTypeID: 787,
			Runs:            10,
			Cost:            &cost,
			Status:          "active",
			ActivityName:    "Manufacturing",
			BlueprintName:   "Rifter Blueprint",
		},
	}

	mocks.jobsRepo.On("GetActiveJobs", mock.Anything, userID).Return(expectedJobs, nil)

	req := httptest.NewRequest("GET", "/v1/industry/jobs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetActiveJobs(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	jobs := result.([]*models.IndustryJob)
	assert.Len(t, jobs, 1)
	assert.Equal(t, int64(10001), jobs[0].JobID)
	assert.Equal(t, "Manufacturing", jobs[0].ActivityName)
	mocks.jobsRepo.AssertExpectations(t)
}

func Test_IndustryController_GetActiveJobs_Error(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	mocks.jobsRepo.On("GetActiveJobs", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/industry/jobs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetActiveJobs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.jobsRepo.AssertExpectations(t)
}

// --- GetAllJobs Tests ---

func Test_IndustryController_GetAllJobs_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	expectedJobs := []*models.IndustryJob{
		{JobID: 10001, UserID: 100, Status: "active", ActivityName: "Manufacturing"},
		{JobID: 10002, UserID: 100, Status: "delivered", ActivityName: "Reaction"},
	}

	mocks.jobsRepo.On("GetAllJobs", mock.Anything, userID).Return(expectedJobs, nil)

	req := httptest.NewRequest("GET", "/v1/industry/jobs/all", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetAllJobs(args)

	assert.Nil(t, httpErr)
	jobs := result.([]*models.IndustryJob)
	assert.Len(t, jobs, 2)
	mocks.jobsRepo.AssertExpectations(t)
}

// --- GetQueue Tests ---

func Test_IndustryController_GetQueue_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	expectedEntries := []*models.IndustryJobQueueEntry{
		{ID: 1, UserID: 100, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 10, Status: "planned"},
		{ID: 2, UserID: 100, BlueprintTypeID: 46166, Activity: "reaction", Runs: 100, Status: "active"},
	}

	mocks.queueRepo.On("GetByUser", mock.Anything, userID).Return(expectedEntries, nil)

	req := httptest.NewRequest("GET", "/v1/industry/queue", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetQueue(args)

	assert.Nil(t, httpErr)
	entries := result.([]*models.IndustryJobQueueEntry)
	assert.Len(t, entries, 2)
	assert.Equal(t, "planned", entries[0].Status)
	assert.Equal(t, "active", entries[1].Status)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_GetQueue_Error(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	mocks.queueRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/industry/queue", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetQueue(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.queueRepo.AssertExpectations(t)
}

// --- CreateQueueEntry Tests ---

func Test_IndustryController_CreateQueueEntry_Reaction(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	body := map[string]any{
		"blueprint_type_id": 46166,
		"activity":          "reaction",
		"runs":              100,
		"me_level":          0,
		"te_level":          0,
		"facility_tax":      0.25,
	}
	bodyBytes, _ := json.Marshal(body)

	expectedEntry := &models.IndustryJobQueueEntry{
		ID:              1,
		UserID:          100,
		BlueprintTypeID: 46166,
		Activity:        "reaction",
		Runs:            100,
		Status:          "planned",
	}

	mocks.queueRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.IndustryJobQueueEntry")).Return(expectedEntry, nil)

	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)

	assert.Nil(t, httpErr)
	entry := result.(*models.IndustryJobQueueEntry)
	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, "reaction", entry.Activity)
	assert.Equal(t, "planned", entry.Status)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_CreateQueueEntry_Manufacturing(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	systemID := int64(30000142)
	body := map[string]any{
		"blueprint_type_id":  787,
		"activity":           "manufacturing",
		"runs":               10,
		"me_level":           10,
		"te_level":           20,
		"industry_skill":     5,
		"adv_industry_skill": 5,
		"system_id":          systemID,
		"facility_tax":       1.0,
		"structure":          "raitaru",
		"rig":                "t2",
		"security":           "high",
	}
	bodyBytes, _ := json.Marshal(body)

	// Set up SDE + market mocks for manufacturing calculation
	sellPrice := 500000.0
	blueprint := &repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787,
		ProductTypeID:   587,
		ProductName:     "Rifter",
		ProductQuantity: 1,
		Time:            3600,
	}
	materials := []*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 22000},
	}
	jitaPrices := map[int64]*models.MarketPrice{
		34:  {TypeID: 34, SellPrice: &sellPrice},
		587: {TypeID: 587, SellPrice: &sellPrice},
	}
	adjustedPrices := map[int64]float64{34: 5.0}
	costIndex := &models.IndustryCostIndex{SystemID: systemID, Activity: "manufacturing", CostIndex: 0.05}

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(blueprint, nil)
	mocks.sdeRepo.On("GetManufacturingMaterials", mock.Anything, int64(787)).Return(materials, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(jitaPrices, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(adjustedPrices, nil)
	mocks.costIndicesRepo.On("GetCostIndex", mock.Anything, systemID, "manufacturing").Return(costIndex, nil)

	mocks.queueRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.IndustryJobQueueEntry")).
		Return(&models.IndustryJobQueueEntry{
			ID:              1,
			UserID:          100,
			BlueprintTypeID: 787,
			Activity:        "manufacturing",
			Runs:            10,
			Status:          "planned",
		}, nil)

	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)

	assert.Nil(t, httpErr)
	entry := result.(*models.IndustryJobQueueEntry)
	assert.Equal(t, int64(1), entry.ID)
	assert.Equal(t, "manufacturing", entry.Activity)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_CreateQueueEntry_InvalidBody(t *testing.T) {
	controller, _ := setupIndustryController()

	userID := int64(100)
	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_CreateQueueEntry_MissingFields(t *testing.T) {
	controller, _ := setupIndustryController()

	userID := int64(100)

	// Missing blueprint_type_id
	body := map[string]any{"activity": "manufacturing", "runs": 10}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_CreateQueueEntry_MissingActivity(t *testing.T) {
	controller, _ := setupIndustryController()

	userID := int64(100)
	body := map[string]any{"blueprint_type_id": 787, "runs": 10}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_CreateQueueEntry_ZeroRuns(t *testing.T) {
	controller, _ := setupIndustryController()

	userID := int64(100)
	body := map[string]any{"blueprint_type_id": 787, "activity": "manufacturing", "runs": 0}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/industry/queue", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateQueueEntry(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- UpdateQueueEntry Tests ---

func Test_IndustryController_UpdateQueueEntry_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	body := map[string]any{
		"blueprint_type_id": 787,
		"activity":          "reaction",
		"runs":              20,
		"me_level":          0,
		"te_level":          0,
		"facility_tax":      0.5,
	}
	bodyBytes, _ := json.Marshal(body)

	expectedEntry := &models.IndustryJobQueueEntry{
		ID:              5,
		UserID:          100,
		BlueprintTypeID: 787,
		Activity:        "reaction",
		Runs:            20,
		Status:          "planned",
	}

	mocks.queueRepo.On("Update", mock.Anything, int64(5), userID, mock.AnythingOfType("*models.IndustryJobQueueEntry")).Return(expectedEntry, nil)

	req := httptest.NewRequest("PUT", "/v1/industry/queue/5", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "5"},
	}

	result, httpErr := controller.UpdateQueueEntry(args)

	assert.Nil(t, httpErr)
	entry := result.(*models.IndustryJobQueueEntry)
	assert.Equal(t, int64(5), entry.ID)
	assert.Equal(t, 20, entry.Runs)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_UpdateQueueEntry_NotFound(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	body := map[string]any{
		"blueprint_type_id": 787,
		"activity":          "reaction",
		"runs":              20,
		"facility_tax":      0,
	}
	bodyBytes, _ := json.Marshal(body)

	mocks.queueRepo.On("Update", mock.Anything, int64(999), userID, mock.AnythingOfType("*models.IndustryJobQueueEntry")).Return(nil, nil)

	req := httptest.NewRequest("PUT", "/v1/industry/queue/999", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	result, httpErr := controller.UpdateQueueEntry(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_UpdateQueueEntry_InvalidID(t *testing.T) {
	controller, _ := setupIndustryController()

	userID := int64(100)
	body := map[string]any{"blueprint_type_id": 787, "activity": "reaction", "runs": 20, "facility_tax": 0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/industry/queue/abc", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "abc"},
	}

	result, httpErr := controller.UpdateQueueEntry(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- CancelQueueEntry Tests ---

func Test_IndustryController_CancelQueueEntry_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	mocks.queueRepo.On("Cancel", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/industry/queue/5", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "5"},
	}

	result, httpErr := controller.CancelQueueEntry(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.queueRepo.AssertExpectations(t)
}

func Test_IndustryController_CancelQueueEntry_NotFound(t *testing.T) {
	controller, mocks := setupIndustryController()

	userID := int64(100)
	mocks.queueRepo.On("Cancel", mock.Anything, int64(999), userID).Return(errors.New("not found or not cancellable"))

	req := httptest.NewRequest("DELETE", "/v1/industry/queue/999", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	result, httpErr := controller.CancelQueueEntry(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	mocks.queueRepo.AssertExpectations(t)
}

// --- Calculate Tests ---

func Test_IndustryController_Calculate_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	systemID := int64(30000142)
	body := map[string]any{
		"blueprint_type_id":  787,
		"runs":               10,
		"me_level":           10,
		"te_level":           20,
		"industry_skill":     5,
		"adv_industry_skill": 5,
		"system_id":          systemID,
		"facility_tax":       1.0,
		"structure":          "raitaru",
		"rig":                "t2",
		"security":           "high",
	}
	bodyBytes, _ := json.Marshal(body)

	sellPrice := 5.0
	productPrice := 500000.0
	blueprint := &repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787,
		ProductTypeID:   587,
		ProductName:     "Rifter",
		ProductQuantity: 1,
		Time:            3600,
	}
	materials := []*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 22000},
	}
	jitaPrices := map[int64]*models.MarketPrice{
		34:  {TypeID: 34, SellPrice: &sellPrice},
		587: {TypeID: 587, SellPrice: &productPrice},
	}
	adjustedPrices := map[int64]float64{34: 5.0}
	costIndex := &models.IndustryCostIndex{SystemID: systemID, Activity: "manufacturing", CostIndex: 0.05}

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(blueprint, nil)
	mocks.sdeRepo.On("GetManufacturingMaterials", mock.Anything, int64(787)).Return(materials, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(jitaPrices, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(adjustedPrices, nil)
	mocks.costIndicesRepo.On("GetCostIndex", mock.Anything, systemID, "manufacturing").Return(costIndex, nil)

	req := httptest.NewRequest("POST", "/v1/industry/calculate", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.Calculate(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	calcResult := result.(*models.ManufacturingCalcResult)
	assert.Equal(t, int64(787), calcResult.BlueprintTypeID)
	assert.Equal(t, int64(587), calcResult.ProductTypeID)
	assert.Equal(t, "Rifter", calcResult.ProductName)
	assert.Equal(t, 10, calcResult.Runs)
	assert.Greater(t, calcResult.MEFactor, 0.0)
	assert.Greater(t, calcResult.TEFactor, 0.0)
	assert.Greater(t, calcResult.InputCost, 0.0)
	assert.Len(t, calcResult.Materials, 1)
	mocks.sdeRepo.AssertExpectations(t)
	mocks.marketRepo.AssertExpectations(t)
	mocks.costIndicesRepo.AssertExpectations(t)
}

func Test_IndustryController_Calculate_InvalidBody(t *testing.T) {
	controller, _ := setupIndustryController()

	req := httptest.NewRequest("POST", "/v1/industry/calculate", bytes.NewReader([]byte("bad")))
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.Calculate(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_Calculate_MissingBlueprintTypeID(t *testing.T) {
	controller, _ := setupIndustryController()

	body := map[string]any{"runs": 10}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/industry/calculate", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.Calculate(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_Calculate_BlueprintNotFound(t *testing.T) {
	controller, mocks := setupIndustryController()

	body := map[string]any{"blueprint_type_id": 99999, "runs": 1}
	bodyBytes, _ := json.Marshal(body)

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(99999)).Return(nil, nil)

	req := httptest.NewRequest("POST", "/v1/industry/calculate", bytes.NewReader(bodyBytes))
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.Calculate(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- SearchBlueprints Tests ---

func Test_IndustryController_SearchBlueprints_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	expectedResults := []*repositories.BlueprintSearchRow{
		{BlueprintTypeID: 787, BlueprintName: "Rifter Blueprint", ProductTypeID: 587, ProductName: "Rifter", Activity: "manufacturing"},
		{BlueprintTypeID: 788, BlueprintName: "Slasher Blueprint", ProductTypeID: 585, ProductName: "Slasher", Activity: "manufacturing"},
	}

	mocks.sdeRepo.On("SearchBlueprints", mock.Anything, "rifter", "", 20).Return(expectedResults, nil)

	req := httptest.NewRequest("GET", "/v1/industry/blueprints?q=rifter", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.SearchBlueprints(args)

	assert.Nil(t, httpErr)
	results := result.([]*repositories.BlueprintSearchRow)
	assert.Len(t, results, 2)
	assert.Equal(t, "Rifter Blueprint", results[0].BlueprintName)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_IndustryController_SearchBlueprints_WithActivity(t *testing.T) {
	controller, mocks := setupIndustryController()

	expectedResults := []*repositories.BlueprintSearchRow{
		{BlueprintTypeID: 787, BlueprintName: "Rifter Blueprint", ProductTypeID: 587, ProductName: "Rifter", Activity: "manufacturing"},
	}

	mocks.sdeRepo.On("SearchBlueprints", mock.Anything, "rifter", "manufacturing", 20).Return(expectedResults, nil)

	req := httptest.NewRequest("GET", "/v1/industry/blueprints?q=rifter&activity=manufacturing", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.SearchBlueprints(args)

	assert.Nil(t, httpErr)
	results := result.([]*repositories.BlueprintSearchRow)
	assert.Len(t, results, 1)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_IndustryController_SearchBlueprints_MissingQuery(t *testing.T) {
	controller, _ := setupIndustryController()

	req := httptest.NewRequest("GET", "/v1/industry/blueprints", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.SearchBlueprints(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_IndustryController_SearchBlueprints_NilResults(t *testing.T) {
	controller, mocks := setupIndustryController()

	mocks.sdeRepo.On("SearchBlueprints", mock.Anything, "nonexistent", "", 20).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/blueprints?q=nonexistent", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.SearchBlueprints(args)

	assert.Nil(t, httpErr)
	results := result.([]*repositories.BlueprintSearchRow)
	assert.Len(t, results, 0)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_IndustryController_SearchBlueprints_CustomLimit(t *testing.T) {
	controller, mocks := setupIndustryController()

	mocks.sdeRepo.On("SearchBlueprints", mock.Anything, "rifter", "", 5).Return([]*repositories.BlueprintSearchRow{}, nil)

	req := httptest.NewRequest("GET", "/v1/industry/blueprints?q=rifter&limit=5", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.SearchBlueprints(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.sdeRepo.AssertExpectations(t)
}

// --- GetSystems Tests ---

func Test_IndustryController_GetSystems_Success(t *testing.T) {
	controller, mocks := setupIndustryController()

	expectedSystems := []*models.ReactionSystem{
		{SystemID: 30000142, Name: "Jita", SecurityStatus: 0.9, CostIndex: 0.05},
		{SystemID: 30002187, Name: "Amarr", SecurityStatus: 1.0, CostIndex: 0.03},
	}

	mocks.sdeRepo.On("GetManufacturingSystems", mock.Anything).Return(expectedSystems, nil)

	req := httptest.NewRequest("GET", "/v1/industry/systems", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.GetSystems(args)

	assert.Nil(t, httpErr)
	systems := result.([]*models.ReactionSystem)
	assert.Len(t, systems, 2)
	assert.Equal(t, "Jita", systems[0].Name)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_IndustryController_GetSystems_NilResults(t *testing.T) {
	controller, mocks := setupIndustryController()

	mocks.sdeRepo.On("GetManufacturingSystems", mock.Anything).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/systems", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.GetSystems(args)

	assert.Nil(t, httpErr)
	systems := result.([]*models.ReactionSystem)
	assert.Len(t, systems, 0)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_IndustryController_GetSystems_Error(t *testing.T) {
	controller, mocks := setupIndustryController()

	mocks.sdeRepo.On("GetManufacturingSystems", mock.Anything).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/industry/systems", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.GetSystems(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.sdeRepo.AssertExpectations(t)
}
