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

type MockProductionPlansRepository struct {
	mock.Mock
}

func (m *MockProductionPlansRepository) Create(ctx context.Context, plan *models.ProductionPlan) (*models.ProductionPlan, error) {
	args := m.Called(ctx, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductionPlan), args.Error(1)
}

func (m *MockProductionPlansRepository) GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductionPlan), args.Error(1)
}

func (m *MockProductionPlansRepository) GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductionPlan), args.Error(1)
}

func (m *MockProductionPlansRepository) Update(ctx context.Context, id, userID int64, name string, notes *string, defaultManufacturingStationID *int64, defaultReactionStationID *int64) error {
	args := m.Called(ctx, id, userID, name, notes, defaultManufacturingStationID, defaultReactionStationID)
	return args.Error(0)
}

func (m *MockProductionPlansRepository) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockProductionPlansRepository) CreateStep(ctx context.Context, step *models.ProductionPlanStep) (*models.ProductionPlanStep, error) {
	args := m.Called(ctx, step)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductionPlanStep), args.Error(1)
}

func (m *MockProductionPlansRepository) UpdateStep(ctx context.Context, stepID, planID, userID int64, step *models.ProductionPlanStep) error {
	args := m.Called(ctx, stepID, planID, userID, step)
	return args.Error(0)
}

func (m *MockProductionPlansRepository) BatchUpdateSteps(ctx context.Context, stepIDs []int64, planID, userID int64, step *models.ProductionPlanStep) (int64, error) {
	args := m.Called(ctx, stepIDs, planID, userID, step)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductionPlansRepository) DeleteStep(ctx context.Context, stepID, planID, userID int64) error {
	args := m.Called(ctx, stepID, planID, userID)
	return args.Error(0)
}

func (m *MockProductionPlansRepository) GetStepMaterials(ctx context.Context, stepID, planID, userID int64) ([]*models.PlanMaterial, error) {
	args := m.Called(ctx, stepID, planID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlanMaterial), args.Error(1)
}

func (m *MockProductionPlansRepository) GetContainersAtStation(ctx context.Context, userID, stationID int64) ([]*models.StationContainer, error) {
	args := m.Called(ctx, userID, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StationContainer), args.Error(1)
}

type MockProductionPlansCharacterRepository struct {
	mock.Mock
}

func (m *MockProductionPlansCharacterRepository) GetNames(ctx context.Context, userID int64) (map[int64]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]string), args.Error(1)
}

type MockProductionPlansCorporationRepository struct {
	mock.Mock
}

func (m *MockProductionPlansCorporationRepository) Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.PlayerCorporation), args.Error(1)
}

func (m *MockProductionPlansCorporationRepository) GetDivisions(ctx context.Context, corp, user int64) (*models.CorporationDivisions, error) {
	args := m.Called(ctx, corp, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CorporationDivisions), args.Error(1)
}

type MockProductionPlansUserStationRepository struct {
	mock.Mock
}

func (m *MockProductionPlansUserStationRepository) GetByID(ctx context.Context, id, userID int64) (*models.UserStation, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStation), args.Error(1)
}

type MockProductionPlansSdeRepository struct {
	mock.Mock
}

func (m *MockProductionPlansSdeRepository) GetBlueprintByProduct(ctx context.Context, productTypeID int64) (*repositories.BlueprintProductRow, error) {
	args := m.Called(ctx, productTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.BlueprintProductRow), args.Error(1)
}

func (m *MockProductionPlansSdeRepository) GetManufacturingBlueprint(ctx context.Context, blueprintTypeID int64) (*repositories.ManufacturingBlueprintRow, error) {
	args := m.Called(ctx, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ManufacturingBlueprintRow), args.Error(1)
}

func (m *MockProductionPlansSdeRepository) GetManufacturingMaterials(ctx context.Context, blueprintTypeID int64) ([]*repositories.ManufacturingMaterialRow, error) {
	args := m.Called(ctx, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.ManufacturingMaterialRow), args.Error(1)
}

type MockProductionPlansMarketRepository struct {
	mock.Mock
}

func (m *MockProductionPlansMarketRepository) GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockProductionPlansMarketRepository) GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

type MockProductionPlansCostIndicesRepository struct {
	mock.Mock
}

func (m *MockProductionPlansCostIndicesRepository) GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error) {
	args := m.Called(ctx, systemID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryCostIndex), args.Error(1)
}

type MockProductionPlansJobQueueRepository struct {
	mock.Mock
}

func (m *MockProductionPlansJobQueueRepository) Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryJobQueueEntry), args.Error(1)
}

type MockProductionPlanRunsRepository struct {
	mock.Mock
}

func (m *MockProductionPlanRunsRepository) Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error) {
	args := m.Called(ctx, run)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductionPlanRun), args.Error(1)
}

func (m *MockProductionPlanRunsRepository) GetByPlan(ctx context.Context, planID, userID int64) ([]*models.ProductionPlanRun, error) {
	args := m.Called(ctx, planID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductionPlanRun), args.Error(1)
}

func (m *MockProductionPlanRunsRepository) GetByID(ctx context.Context, runID, userID int64) (*models.ProductionPlanRun, error) {
	args := m.Called(ctx, runID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductionPlanRun), args.Error(1)
}

func (m *MockProductionPlanRunsRepository) Delete(ctx context.Context, runID, userID int64) error {
	args := m.Called(ctx, runID, userID)
	return args.Error(0)
}

// --- Helper ---

type productionPlanMocks struct {
	plansRepo       *MockProductionPlansRepository
	sdeRepo         *MockProductionPlansSdeRepository
	queueRepo       *MockProductionPlansJobQueueRepository
	marketRepo      *MockProductionPlansMarketRepository
	costIndicesRepo *MockProductionPlansCostIndicesRepository
	characterRepo   *MockProductionPlansCharacterRepository
	corpRepo        *MockProductionPlansCorporationRepository
	stationRepo     *MockProductionPlansUserStationRepository
	runsRepo        *MockProductionPlanRunsRepository
}

func setupProductionPlansController() (*controllers.ProductionPlans, *productionPlanMocks) {
	mocks := &productionPlanMocks{
		plansRepo:       new(MockProductionPlansRepository),
		sdeRepo:         new(MockProductionPlansSdeRepository),
		queueRepo:       new(MockProductionPlansJobQueueRepository),
		marketRepo:      new(MockProductionPlansMarketRepository),
		costIndicesRepo: new(MockProductionPlansCostIndicesRepository),
		characterRepo:   new(MockProductionPlansCharacterRepository),
		corpRepo:        new(MockProductionPlansCorporationRepository),
		stationRepo:     new(MockProductionPlansUserStationRepository),
		runsRepo:        new(MockProductionPlanRunsRepository),
	}

	controller := controllers.NewProductionPlans(
		&MockRouter{},
		mocks.plansRepo,
		mocks.sdeRepo,
		mocks.queueRepo,
		mocks.marketRepo,
		mocks.costIndicesRepo,
		mocks.characterRepo,
		mocks.corpRepo,
		mocks.stationRepo,
		mocks.runsRepo,
	)

	return controller, mocks
}

// --- GetPlans Tests ---

func Test_ProductionPlans_GetPlans_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	expectedPlans := []*models.ProductionPlan{
		{ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter", ProductName: "Rifter"},
		{ID: 2, UserID: 100, ProductTypeID: 34, Name: "Tritanium Plan", ProductName: "Tritanium"},
	}

	mocks.plansRepo.On("GetByUser", mock.Anything, userID).Return(expectedPlans, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetPlans(args)

	assert.Nil(t, httpErr)
	plans := result.([]*models.ProductionPlan)
	assert.Len(t, plans, 2)
	assert.Equal(t, "Rifter", plans[0].Name)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetPlans_Error(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/industry/plans", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetPlans(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- CreatePlan Tests ---

func Test_ProductionPlans_CreatePlan_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(587)).Return(&repositories.BlueprintProductRow{
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		ProductQuantity: 1,
	}, nil)

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787,
		ProductTypeID:   587,
		ProductName:     "Rifter",
		ProductQuantity: 1,
		Time:            7200,
	}, nil)

	mocks.plansRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.ProductionPlan) bool {
		return p.ProductTypeID == 587 && p.Name == "Rifter"
	})).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
	}, nil)

	mocks.plansRepo.On("CreateStep", mock.Anything, mock.MatchedBy(func(s *models.ProductionPlanStep) bool {
		return s.PlanID == 1 && s.BlueprintTypeID == 787 && s.Activity == "manufacturing" && s.ParentStepID == nil
	})).Return(&models.ProductionPlanStep{
		ID: 10, PlanID: 1, BlueprintTypeID: 787,
	}, nil)

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter", ProductName: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{ID: 10, PlanID: 1, BlueprintTypeID: 787, Activity: "manufacturing", ProductName: "Rifter"},
		},
	}, nil)

	body, _ := json.Marshal(map[string]any{"product_type_id": 587})
	req := httptest.NewRequest("POST", "/v1/industry/plans", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreatePlan(args)

	assert.Nil(t, httpErr)
	plan := result.(*models.ProductionPlan)
	assert.Equal(t, "Rifter", plan.Name)
	assert.Len(t, plan.Steps, 1)
	mocks.plansRepo.AssertExpectations(t)
	mocks.sdeRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CreatePlan_NoBlueprintFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(34)).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{"product_type_id": 34})
	req := httptest.NewRequest("POST", "/v1/industry/plans", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreatePlan(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ProductionPlans_CreatePlan_MissingProductTypeID(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest("POST", "/v1/industry/plans", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreatePlan(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- GetPlan Tests ---

func Test_ProductionPlans_GetPlan_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	expected := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{ID: 10, PlanID: 1, Activity: "manufacturing"},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/1", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GetPlan(args)

	assert.Nil(t, httpErr)
	plan := result.(*models.ProductionPlan)
	assert.Equal(t, "Rifter", plan.Name)
	assert.Len(t, plan.Steps, 1)
}

func Test_ProductionPlans_GetPlan_NotFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("GetByID", mock.Anything, int64(999), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/999", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "999"}}

	result, httpErr := controller.GetPlan(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- DeletePlan Tests ---

func Test_ProductionPlans_DeletePlan_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("Delete", mock.Anything, int64(1), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/industry/plans/1", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.DeletePlan(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

// --- CreateStep Tests ---

func Test_ProductionPlans_CreateStep_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100,
		Steps: []*models.ProductionPlanStep{{ID: 10, PlanID: 1}},
	}, nil)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(34)).Return(&repositories.BlueprintProductRow{
		BlueprintTypeID: 100,
		Activity:        "manufacturing",
		ProductQuantity: 1,
	}, nil)

	mocks.plansRepo.On("CreateStep", mock.Anything, mock.MatchedBy(func(s *models.ProductionPlanStep) bool {
		return s.PlanID == 1 && *s.ParentStepID == 10 && s.ProductTypeID == 34 && s.BlueprintTypeID == 100
	})).Return(&models.ProductionPlanStep{
		ID: 20, PlanID: 1, ProductTypeID: 34, BlueprintTypeID: 100,
	}, nil)

	body, _ := json.Marshal(map[string]any{"parent_step_id": 10, "product_type_id": 34})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/steps", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.CreateStep(args)

	assert.Nil(t, httpErr)
	step := result.(*models.ProductionPlanStep)
	assert.Equal(t, int64(20), step.ID)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CreateStep_NoBlueprintForMaterial(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, Steps: []*models.ProductionPlanStep{{ID: 10}},
	}, nil)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(9999)).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{"parent_step_id": 10, "product_type_id": 9999})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/steps", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.CreateStep(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- UpdateStep Tests ---

func Test_ProductionPlans_UpdateStep_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("UpdateStep", mock.Anything, int64(10), int64(1), userID, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"me_level": 8, "te_level": 16,
		"structure": "azbel", "rig": "t2", "security": "low",
		"facility_tax": 2.5,
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1/steps/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "stepId": "10"}}

	result, httpErr := controller.UpdateStep(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_UpdateStep_WithOutputLocation(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("UpdateStep", mock.Anything, int64(10), int64(1), userID, mock.MatchedBy(func(step *models.ProductionPlanStep) bool {
		return step.OutputOwnerType != nil && *step.OutputOwnerType == "corporation" &&
			step.OutputOwnerID != nil && *step.OutputOwnerID == int64(5000) &&
			step.OutputDivisionNumber != nil && *step.OutputDivisionNumber == 3 &&
			step.OutputContainerID != nil && *step.OutputContainerID == int64(9999)
	})).Return(nil)

	ownerType := "corporation"
	ownerID := int64(5000)
	divNum := 3
	containerID := int64(9999)
	body, _ := json.Marshal(map[string]any{
		"me_level": 10, "te_level": 20,
		"structure": "raitaru", "rig": "t2", "security": "high",
		"facility_tax":           1.0,
		"output_owner_type":      ownerType,
		"output_owner_id":        ownerID,
		"output_division_number": divNum,
		"output_container_id":    containerID,
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1/steps/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "stepId": "10"}}

	result, httpErr := controller.UpdateStep(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

// --- DeleteStep Tests ---

func Test_ProductionPlans_DeleteStep_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("DeleteStep", mock.Anything, int64(10), int64(1), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/industry/plans/1/steps/10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "stepId": "10"}}

	result, httpErr := controller.DeleteStep(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

// --- GetStepMaterials Tests ---

func Test_ProductionPlans_GetStepMaterials_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	bpTypeID := int64(100)
	activity := "manufacturing"
	expectedMaterials := []*models.PlanMaterial{
		{TypeID: 34, TypeName: "Tritanium", Quantity: 1000, HasBlueprint: false, IsProduced: false},
		{TypeID: 35, TypeName: "Pyerite", Quantity: 500, HasBlueprint: false, IsProduced: false},
		{TypeID: 5678, TypeName: "Component", Quantity: 10, HasBlueprint: true, BlueprintTypeID: &bpTypeID, Activity: &activity, IsProduced: false},
	}

	mocks.plansRepo.On("GetStepMaterials", mock.Anything, int64(10), int64(1), userID).Return(expectedMaterials, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/1/steps/10/materials", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "stepId": "10"}}

	result, httpErr := controller.GetStepMaterials(args)

	assert.Nil(t, httpErr)
	materials := result.([]*models.PlanMaterial)
	assert.Len(t, materials, 3)
	assert.False(t, materials[0].HasBlueprint)
	assert.True(t, materials[2].HasBlueprint)
	mocks.plansRepo.AssertExpectations(t)
}

// --- GenerateJobs Tests ---

func Test_ProductionPlans_GenerateJobs_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)

	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	mocks.runsRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *models.ProductionPlanRun) bool {
		return r.PlanID == 1 && r.UserID == 100 && r.Quantity == 40
	})).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 40,
	}, nil)

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)

	mocks.sdeRepo.On("GetManufacturingMaterials", mock.Anything, int64(787)).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 1000},
	}, nil)

	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787 && e.Runs == 40 && e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: 100, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 40, Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 40})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.NotNil(t, genResult.Run)
	assert.Equal(t, int64(50), genResult.Run.ID)
	assert.Equal(t, 40, genResult.Run.Quantity)
	assert.Len(t, genResult.Created, 1)
	assert.Len(t, genResult.Skipped, 0)
	assert.Equal(t, 40, genResult.Created[0].Runs)
}

func Test_ProductionPlans_GenerateJobs_WithChildStep(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID := int64(20)
	parentRef := rootStepID

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter",
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 51, PlanID: 1, UserID: 100, Quantity: 2,
	}, nil)

	// Root blueprint: produces 1 Rifter per run, needs 10 Components
	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)
	mocks.sdeRepo.On("GetManufacturingMaterials", mock.Anything, int64(787)).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
		{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 1000},
	}, nil)

	// Child blueprint: produces 5 Components per run
	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(1234)).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600,
	}, nil)
	mocks.sdeRepo.On("GetManufacturingMaterials", mock.Anything, int64(1234)).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 1234, TypeID: 34, TypeName: "Tritanium", Quantity: 100},
	}, nil)

	// Expect child queue entry created first (depth-first bottom-up)
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 1234
	})).Return(&models.IndustryJobQueueEntry{
		ID: 98, BlueprintTypeID: 1234, Activity: "manufacturing", Status: "planned",
	}, nil).Once()

	// Then root queue entry
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 2, Status: "planned",
	}, nil).Once()

	body, _ := json.Marshal(map[string]any{"quantity": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.Created, 2)
	assert.Len(t, genResult.Skipped, 0)
	// Child should be created first (bottom-up)
	assert.Equal(t, int64(1234), genResult.Created[0].BlueprintTypeID)
	assert.Equal(t, int64(787), genResult.Created[1].BlueprintTypeID)
}

func Test_ProductionPlans_GenerateJobs_PlanNotFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("GetByID", mock.Anything, int64(999), userID).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 10})
	req := httptest.NewRequest("POST", "/v1/industry/plans/999/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "999"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_ProductionPlans_GenerateJobs_InvalidQuantity(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	body, _ := json.Marshal(map[string]any{"quantity": 0})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ProductionPlans_GenerateJobs_BlueprintDataMissing(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{
				ID: 10, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 52, PlanID: 1, UserID: 100, Quantity: 10,
	}, nil)

	// Blueprint data not found
	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 10})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.Created, 0)
	assert.Len(t, genResult.Skipped, 1)
	assert.Equal(t, "blueprint data not found", genResult.Skipped[0].Reason)
}

// --- Default Station Tests ---

func Test_ProductionPlans_CreatePlan_WithDefaultStations(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mfgStationID := int64(5)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(587)).Return(&repositories.BlueprintProductRow{
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		ProductQuantity: 1,
	}, nil)

	mocks.sdeRepo.On("GetManufacturingBlueprint", mock.Anything, int64(787)).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)

	mocks.plansRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.ProductionPlan) bool {
		return p.ProductTypeID == 587 &&
			p.DefaultManufacturingStationID != nil && *p.DefaultManufacturingStationID == 5
	})).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		DefaultManufacturingStationID: &mfgStationID,
	}, nil)

	// Root step should have UserStationID set to manufacturing station
	mocks.plansRepo.On("CreateStep", mock.Anything, mock.MatchedBy(func(s *models.ProductionPlanStep) bool {
		return s.PlanID == 1 && s.Activity == "manufacturing" &&
			s.UserStationID != nil && *s.UserStationID == 5
	})).Return(&models.ProductionPlanStep{
		ID: 10, PlanID: 1, BlueprintTypeID: 787, UserStationID: &mfgStationID,
	}, nil)

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		DefaultManufacturingStationID: &mfgStationID,
		Steps: []*models.ProductionPlanStep{
			{ID: 10, PlanID: 1, BlueprintTypeID: 787, Activity: "manufacturing", UserStationID: &mfgStationID},
		},
	}, nil)

	body, _ := json.Marshal(map[string]any{
		"product_type_id":                   587,
		"default_manufacturing_station_id": 5,
	})
	req := httptest.NewRequest("POST", "/v1/industry/plans", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreatePlan(args)

	assert.Nil(t, httpErr)
	plan := result.(*models.ProductionPlan)
	assert.NotNil(t, plan.DefaultManufacturingStationID)
	assert.Equal(t, int64(5), *plan.DefaultManufacturingStationID)
	assert.NotNil(t, plan.Steps[0].UserStationID)
	assert.Equal(t, int64(5), *plan.Steps[0].UserStationID)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CreateStep_InheritsDefaultStation(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mfgStationID := int64(5)

	// Plan has a default manufacturing station
	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100,
		DefaultManufacturingStationID: &mfgStationID,
		Steps:                         []*models.ProductionPlanStep{{ID: 10, PlanID: 1}},
	}, nil)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(34)).Return(&repositories.BlueprintProductRow{
		BlueprintTypeID: 100,
		Activity:        "manufacturing",
		ProductQuantity: 1,
	}, nil)

	// Step should inherit the manufacturing station
	mocks.plansRepo.On("CreateStep", mock.Anything, mock.MatchedBy(func(s *models.ProductionPlanStep) bool {
		return s.PlanID == 1 && s.ProductTypeID == 34 &&
			s.UserStationID != nil && *s.UserStationID == 5
	})).Return(&models.ProductionPlanStep{
		ID: 20, PlanID: 1, ProductTypeID: 34, UserStationID: &mfgStationID,
	}, nil)

	body, _ := json.Marshal(map[string]any{"parent_step_id": 10, "product_type_id": 34})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/steps", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.CreateStep(args)

	assert.Nil(t, httpErr)
	step := result.(*models.ProductionPlanStep)
	assert.NotNil(t, step.UserStationID)
	assert.Equal(t, int64(5), *step.UserStationID)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CreateStep_InheritsReactionStation(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rxnStationID := int64(7)

	// Plan has a default reaction station
	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100,
		DefaultReactionStationID: &rxnStationID,
		Steps:                    []*models.ProductionPlanStep{{ID: 10, PlanID: 1}},
	}, nil)

	mocks.sdeRepo.On("GetBlueprintByProduct", mock.Anything, int64(9000)).Return(&repositories.BlueprintProductRow{
		BlueprintTypeID: 9001,
		Activity:        "reaction",
		ProductQuantity: 1,
	}, nil)

	// Step should inherit the reaction station
	mocks.plansRepo.On("CreateStep", mock.Anything, mock.MatchedBy(func(s *models.ProductionPlanStep) bool {
		return s.PlanID == 1 && s.ProductTypeID == 9000 &&
			s.UserStationID != nil && *s.UserStationID == 7
	})).Return(&models.ProductionPlanStep{
		ID: 30, PlanID: 1, ProductTypeID: 9000, UserStationID: &rxnStationID,
	}, nil)

	body, _ := json.Marshal(map[string]any{"parent_step_id": 10, "product_type_id": 9000})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/steps", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.CreateStep(args)

	assert.Nil(t, httpErr)
	step := result.(*models.ProductionPlanStep)
	assert.NotNil(t, step.UserStationID)
	assert.Equal(t, int64(7), *step.UserStationID)
	mocks.plansRepo.AssertExpectations(t)
}

// --- GetHangars Tests ---

func Test_ProductionPlans_GetHangars_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	stationID := int64(60003760)

	mocks.stationRepo.On("GetByID", mock.Anything, int64(5), userID).Return(&models.UserStation{
		ID: 5, StationID: stationID,
	}, nil)

	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{
		1001: "Main Char",
		1002: "Alt Char",
	}, nil)

	mocks.corpRepo.On("Get", mock.Anything, userID).Return([]repositories.PlayerCorporation{
		{ID: 2001, Name: "My Corp"},
	}, nil)

	mocks.corpRepo.On("GetDivisions", mock.Anything, int64(2001), userID).Return(&models.CorporationDivisions{
		Hanger: map[int]string{1: "Main Hangar", 2: "Manufacturing"},
		Wallet: map[int]string{},
	}, nil)

	divNum := 2
	mocks.plansRepo.On("GetContainersAtStation", mock.Anything, userID, stationID).Return([]*models.StationContainer{
		{ID: 99001, Name: "BPC Container", OwnerType: "character", OwnerID: 1001},
		{ID: 99002, Name: "Components Box", OwnerType: "corporation", OwnerID: 2001, DivisionNumber: &divNum},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/hangars?user_station_id=5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetHangars(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify character/corp/container data
	mocks.stationRepo.AssertExpectations(t)
	mocks.characterRepo.AssertExpectations(t)
	mocks.corpRepo.AssertExpectations(t)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetHangars_MissingStationID(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/industry/plans/hangars", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetHangars(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- BatchUpdateSteps Tests ---

func Test_ProductionPlans_BatchUpdateSteps_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("BatchUpdateSteps", mock.Anything,
		[]int64{10, 20, 30}, int64(1), userID,
		mock.MatchedBy(func(step *models.ProductionPlanStep) bool {
			return step.MELevel == 8 && step.TELevel == 16 && step.Structure == "azbel"
		}),
	).Return(int64(3), nil)

	body, _ := json.Marshal(map[string]any{
		"step_ids":     []int64{10, 20, 30},
		"me_level":     8,
		"te_level":     16,
		"structure":    "azbel",
		"rig":          "t2",
		"security":     "low",
		"facility_tax": 2.5,
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1/steps/batch", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.BatchUpdateSteps(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resultMap := result.(map[string]any)
	assert.Equal(t, "updated", resultMap["status"])
	assert.Equal(t, int64(3), resultMap["rows_affected"])
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_BatchUpdateSteps_EmptyStepIDs(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	body, _ := json.Marshal(map[string]any{
		"step_ids":     []int64{},
		"me_level":     10,
		"te_level":     20,
		"structure":    "raitaru",
		"rig":          "t2",
		"security":     "high",
		"facility_tax": 1.0,
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1/steps/batch", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.BatchUpdateSteps(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ProductionPlans_GetHangars_StationNotFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.stationRepo.On("GetByID", mock.Anything, int64(999), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/hangars?user_station_id=999", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetHangars(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- Plan Runs Tests ---

func Test_ProductionPlans_GetPlanRuns_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(&models.ProductionPlan{
		ID: 1, UserID: 100, Name: "Test Plan",
	}, nil)

	expectedRuns := []*models.ProductionPlanRun{
		{ID: 10, PlanID: 1, UserID: 100, Quantity: 5, Status: "completed", PlanName: "Test Plan"},
		{ID: 11, PlanID: 1, UserID: 100, Quantity: 10, Status: "pending", PlanName: "Test Plan"},
	}

	mocks.runsRepo.On("GetByPlan", mock.Anything, int64(1), userID).Return(expectedRuns, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/1/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GetPlanRuns(args)

	assert.Nil(t, httpErr)
	runs := result.([]*models.ProductionPlanRun)
	assert.Len(t, runs, 2)
	assert.Equal(t, int64(10), runs[0].ID)
	assert.Equal(t, "completed", runs[0].Status)
	mocks.plansRepo.AssertExpectations(t)
	mocks.runsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetPlanRuns_PlanNotFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.plansRepo.On("GetByID", mock.Anything, int64(999), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/999/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "999"}}

	result, httpErr := controller.GetPlanRuns(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_ProductionPlans_GetPlanRun_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)

	expectedRun := &models.ProductionPlanRun{
		ID: 10, PlanID: 1, UserID: 100, Quantity: 5, Status: "in_progress",
		PlanName: "Test Plan",
		Jobs: []*models.IndustryJobQueueEntry{
			{ID: 99, Activity: "manufacturing", Status: "active"},
			{ID: 100, Activity: "manufacturing", Status: "planned"},
		},
		JobSummary: &models.PlanRunJobSummary{Total: 2, Active: 1, Planned: 1},
	}

	mocks.runsRepo.On("GetByID", mock.Anything, int64(10), userID).Return(expectedRun, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/1/runs/10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "runId": "10"}}

	result, httpErr := controller.GetPlanRun(args)

	assert.Nil(t, httpErr)
	run := result.(*models.ProductionPlanRun)
	assert.Equal(t, int64(10), run.ID)
	assert.Equal(t, "in_progress", run.Status)
	assert.Len(t, run.Jobs, 2)
	mocks.runsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetPlanRun_NotFound(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.runsRepo.On("GetByID", mock.Anything, int64(999), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/1/runs/999", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "runId": "999"}}

	result, httpErr := controller.GetPlanRun(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_ProductionPlans_DeletePlanRun_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.runsRepo.On("Delete", mock.Anything, int64(10), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/industry/plans/1/runs/10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1", "runId": "10"}}

	result, httpErr := controller.DeletePlanRun(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resultMap := result.(map[string]string)
	assert.Equal(t, "deleted", resultMap["status"])
	mocks.runsRepo.AssertExpectations(t)
}
