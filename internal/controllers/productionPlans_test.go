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

func (m *MockProductionPlansRepository) Update(ctx context.Context, id, userID int64, name string, notes *string, defaultManufacturingStationID *int64, defaultReactionStationID *int64, transportFulfillment *string, transportMethod *string, transportProfileID *int64, courierRatePerM3 float64, courierCollateralRate float64) error {
	args := m.Called(ctx, id, userID, name, notes, defaultManufacturingStationID, defaultReactionStationID, transportFulfillment, transportMethod, transportProfileID, courierRatePerM3, courierCollateralRate)
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

func (m *MockProductionPlansRepository) GetByProductTypeAndUser(ctx context.Context, productTypeID, userID int64) ([]*models.ProductionPlan, error) {
	args := m.Called(ctx, productTypeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductionPlan), args.Error(1)
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

func (m *MockProductionPlansSdeRepository) GetBlueprintForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*repositories.ManufacturingBlueprintRow, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ManufacturingBlueprintRow), args.Error(1)
}

func (m *MockProductionPlansSdeRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*repositories.ManufacturingMaterialRow, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
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

func (m *MockProductionPlansJobQueueRepository) GetSlotUsage(ctx context.Context, userID int64) (map[int64]map[string]int, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]map[string]int), args.Error(1)
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

func (m *MockProductionPlanRunsRepository) GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlanRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProductionPlanRun), args.Error(1)
}

func (m *MockProductionPlanRunsRepository) CancelPlannedJobs(ctx context.Context, runID, userID int64) (int64, error) {
	args := m.Called(ctx, runID, userID)
	return args.Get(0).(int64), args.Error(1)
}

type MockProductionPlansTransportJobsRepository struct {
	mock.Mock
}

func (m *MockProductionPlansTransportJobsRepository) Create(ctx context.Context, job *models.TransportJob) (*models.TransportJob, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportJob), args.Error(1)
}

func (m *MockProductionPlansTransportJobsRepository) SetQueueEntryID(ctx context.Context, id int64, queueEntryID int64) error {
	args := m.Called(ctx, id, queueEntryID)
	return args.Error(0)
}

type MockProductionPlansTransportProfilesRepository struct {
	mock.Mock
}

func (m *MockProductionPlansTransportProfilesRepository) GetByID(ctx context.Context, id, userID int64) (*models.TransportProfile, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}

func (m *MockProductionPlansTransportProfilesRepository) GetDefaultByMethod(ctx context.Context, userID int64, method string) (*models.TransportProfile, error) {
	args := m.Called(ctx, userID, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}

type MockProductionPlansJFRoutesRepository struct {
	mock.Mock
}

func (m *MockProductionPlansJFRoutesRepository) FindBySystemPair(ctx context.Context, userID, originSystemID, destSystemID int64) (*models.JFRoute, error) {
	args := m.Called(ctx, userID, originSystemID, destSystemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JFRoute), args.Error(1)
}

type MockProductionPlansEsiClient struct {
	mock.Mock
}

func (m *MockProductionPlansEsiClient) GetRoute(ctx context.Context, origin, destination int64, flag string) ([]int32, error) {
	args := m.Called(ctx, origin, destination, flag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int32), args.Error(1)
}

// --- Mock for character skills ---

type MockProductionPlansCharacterSkillsRepository struct {
	mock.Mock
}

func (m *MockProductionPlansCharacterSkillsRepository) GetSkillsForUser(ctx context.Context, userID int64) ([]*models.CharacterSkill, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CharacterSkill), args.Error(1)
}

// --- Helper ---

type productionPlanMocks struct {
	plansRepo        *MockProductionPlansRepository
	sdeRepo          *MockProductionPlansSdeRepository
	queueRepo        *MockProductionPlansJobQueueRepository
	marketRepo       *MockProductionPlansMarketRepository
	costIndicesRepo  *MockProductionPlansCostIndicesRepository
	characterRepo    *MockProductionPlansCharacterRepository
	corpRepo         *MockProductionPlansCorporationRepository
	stationRepo      *MockProductionPlansUserStationRepository
	runsRepo         *MockProductionPlanRunsRepository
	transportJobRepo *MockProductionPlansTransportJobsRepository
	profilesRepo     *MockProductionPlansTransportProfilesRepository
	jfRoutesRepo     *MockProductionPlansJFRoutesRepository
	esiClient        *MockProductionPlansEsiClient
	skillsRepo       *MockProductionPlansCharacterSkillsRepository
}

func setupProductionPlansController() (*controllers.ProductionPlans, *productionPlanMocks) {
	mocks := &productionPlanMocks{
		plansRepo:        new(MockProductionPlansRepository),
		sdeRepo:          new(MockProductionPlansSdeRepository),
		queueRepo:        new(MockProductionPlansJobQueueRepository),
		marketRepo:       new(MockProductionPlansMarketRepository),
		costIndicesRepo:  new(MockProductionPlansCostIndicesRepository),
		characterRepo:    new(MockProductionPlansCharacterRepository),
		corpRepo:         new(MockProductionPlansCorporationRepository),
		stationRepo:      new(MockProductionPlansUserStationRepository),
		runsRepo:         new(MockProductionPlanRunsRepository),
		transportJobRepo: new(MockProductionPlansTransportJobsRepository),
		profilesRepo:     new(MockProductionPlansTransportProfilesRepository),
		jfRoutesRepo:     new(MockProductionPlansJFRoutesRepository),
		esiClient:        new(MockProductionPlansEsiClient),
		skillsRepo:       new(MockProductionPlansCharacterSkillsRepository),
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
		mocks.transportJobRepo,
		mocks.profilesRepo,
		mocks.jfRoutesRepo,
		mocks.esiClient,
		mocks.skillsRepo,
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

// --- GetPlansByProduct Tests ---

func Test_ProductionPlans_GetPlansByProduct_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	typeID := int64(587)
	expectedPlans := []*models.ProductionPlan{
		{ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter Plan", ProductName: "Rifter"},
	}

	mocks.plansRepo.On("GetByProductTypeAndUser", mock.Anything, typeID, userID).Return(expectedPlans, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/by-product/587", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"typeId": "587"},
	}

	result, httpErr := controller.GetPlansByProduct(args)

	assert.Nil(t, httpErr)
	plans := result.([]*models.ProductionPlan)
	assert.Len(t, plans, 1)
	assert.Equal(t, "Rifter Plan", plans[0].Name)
	assert.Equal(t, int64(587), plans[0].ProductTypeID)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetPlansByProduct_EmptyResult(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	typeID := int64(999)

	mocks.plansRepo.On("GetByProductTypeAndUser", mock.Anything, typeID, userID).Return([]*models.ProductionPlan{}, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/by-product/999", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"typeId": "999"},
	}

	result, httpErr := controller.GetPlansByProduct(args)

	assert.Nil(t, httpErr)
	plans := result.([]*models.ProductionPlan)
	assert.Len(t, plans, 0)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetPlansByProduct_InvalidTypeID(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/industry/plans/by-product/notanumber", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"typeId": "notanumber"},
	}

	result, httpErr := controller.GetPlansByProduct(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ProductionPlans_GetPlansByProduct_RepoError(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	typeID := int64(587)

	mocks.plansRepo.On("GetByProductTypeAndUser", mock.Anything, typeID, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/industry/plans/by-product/587", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"typeId": "587"},
	}

	result, httpErr := controller.GetPlansByProduct(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.plansRepo.AssertExpectations(t)
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

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
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

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
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
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
		{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 1000},
	}, nil)

	// Child blueprint: produces 5 Components per run
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
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

// Test_ProductionPlans_GenerateJobs_ReactionUsesReactionTE verifies that reaction steps
// use the reactions TE formula (Reactions skill only, no blueprint TE, no Adv Industry)
// rather than the manufacturing TE formula when calculating estimated duration.
//
// The key distinction:
//   - Reactions TE: (1 - skill*0.04) * structTE * rigTE  — no blueprint TE, no Adv Industry
//   - Manufacturing TE: (1 - bpTE/100) * (1 - industry*0.04) * (1 - advIndustry*0.03) * structTE * rigTE
//
// step.IndustrySkill doubles as the Reactions skill level when activity == "reaction".
func Test_ProductionPlans_GenerateJobs_ReactionUsesReactionTE(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	stepID := int64(10)

	// IndustrySkill=3 means Reactions skill level 3 for reaction steps.
	// AdvIndustrySkill=4 is set but must NOT be applied to the duration for reactions.
	// BlueprintTE=10 is set but must NOT be applied either.
	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 9000, Name: "Platinum Technite",
		Steps: []*models.ProductionPlanStep{
			{
				ID: stepID, PlanID: 1, ProductTypeID: 9000, BlueprintTypeID: 9001,
				Activity:         "reaction",
				MELevel:          0,
				TELevel:          10, // must be ignored for reactions
				IndustrySkill:    3,  // this is the Reactions skill
				AdvIndustrySkill: 4,  // must be ignored for reactions
				Structure:        "athanor",
				Rig:              "none",
				Security:         "low",
				FacilityTax:      1.0,
				ProductName:      "Platinum Technite",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 10,
	}, nil)

	// Blueprint produces 1 unit per run, baseTime=3600 (typical reaction time)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(9001), "reaction").Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 9001, ProductTypeID: 9000, ProductName: "Platinum Technite",
		ProductQuantity: 1, Time: 3600,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(9001), "reaction").Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 9001, TypeID: 34, TypeName: "Platinum", Quantity: 100},
	}, nil)

	// Reactions TE formula for skill=3, athanor, no rig, low-sec:
	//   teFactor = (1 - 3*0.04) * (1 - 0) * (1 - 0*1.0) = 0.88
	//   secsPerRun = floor(3600 * 0.88) = 3168
	//   totalDuration (10 runs) = 31680
	expectedDuration := 31680

	// Manufacturing formula would give different (lower) result:
	//   mfgTE = (1-10/100)*(1-3*0.04)*(1-4*0.03)*(1-0)*(1-0*1.9) = 0.9*0.88*0.88 = 0.697536
	//   secsPerRun = floor(3600 * 0.697536) = 2511
	//   totalDuration (10 runs) = 25110
	manufacturingFormulaDuration := 25110

	// Capture the entry passed to Create so we can assert the duration on the actual input,
	// not on the mock's return value (which would not have EstimatedDuration set).
	var capturedEntry *models.IndustryJobQueueEntry
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		if e.BlueprintTypeID == 9001 && e.Activity == "reaction" {
			capturedEntry = e
			return true
		}
		return false
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: 100, BlueprintTypeID: 9001, Activity: "reaction", Runs: 10, Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 10})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.Created, 1)

	// Assert on the captured entry (what was passed to Create), not the mock return value
	assert.NotNil(t, capturedEntry, "expected queueRepo.Create to be called with a reaction entry")
	if capturedEntry != nil {
		assert.NotNil(t, capturedEntry.EstimatedDuration,
			"reaction step must produce a non-nil estimated duration")
		if capturedEntry.EstimatedDuration != nil {
			assert.Equal(t, expectedDuration, *capturedEntry.EstimatedDuration,
				"reaction step must use reactions TE formula (Reactions skill only), not manufacturing TE formula")
			assert.NotEqual(t, manufacturingFormulaDuration, *capturedEntry.EstimatedDuration,
				"reaction step must not apply blueprint TE and AdvIndustrySkill like manufacturing does")
		}
	}
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
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(nil, nil)

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

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
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

func Test_ProductionPlans_GetAllRuns_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	runs := []*models.ProductionPlanRun{
		{ID: 1, PlanID: 5, UserID: userID, Quantity: 10, PlanName: "Rifter Plan", ProductName: "Rifter", Status: "in_progress", JobSummary: &models.PlanRunJobSummary{Total: 5, Planned: 2, Active: 2, Completed: 1}},
		{ID: 2, PlanID: 8, UserID: userID, Quantity: 3, PlanName: "Slasher Plan", ProductName: "Slasher", Status: "completed", JobSummary: &models.PlanRunJobSummary{Total: 3, Completed: 3}},
	}
	mocks.runsRepo.On("GetByUser", mock.Anything, userID).Return(runs, nil)

	req := httptest.NewRequest("GET", "/v1/industry/plans/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetAllRuns(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resultRuns := result.([]*models.ProductionPlanRun)
	assert.Len(t, resultRuns, 2)
	assert.Equal(t, "Rifter Plan", resultRuns[0].PlanName)
	assert.Equal(t, "Slasher Plan", resultRuns[1].PlanName)
	mocks.runsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GetAllRuns_Error(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.runsRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/industry/plans/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetAllRuns(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.runsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CancelPlanRun_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.runsRepo.On("CancelPlannedJobs", mock.Anything, int64(10), userID).Return(int64(3), nil)

	req := httptest.NewRequest("POST", "/v1/industry/plans/runs/10/cancel", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"runId": "10"}}

	result, httpErr := controller.CancelPlanRun(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resultMap := result.(map[string]any)
	assert.Equal(t, "cancelled", resultMap["status"])
	assert.Equal(t, int64(3), resultMap["jobsCancelled"])
	mocks.runsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_CancelPlanRun_InvalidID(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)
	req := httptest.NewRequest("POST", "/v1/industry/plans/runs/abc/cancel", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"runId": "abc"}}

	result, httpErr := controller.CancelPlanRun(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ProductionPlans_CancelPlanRun_Error(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	mocks.runsRepo.On("CancelPlannedJobs", mock.Anything, int64(10), userID).Return(int64(0), errors.New("db error"))

	req := httptest.NewRequest("POST", "/v1/industry/plans/runs/10/cancel", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"runId": "10"}}

	result, httpErr := controller.CancelPlanRun(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.runsRepo.AssertExpectations(t)
}

// --- UpdatePlan Tests ---

func Test_ProductionPlans_UpdatePlan_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	fulfillment := "self_haul"
	method := "freighter"

	mocks.plansRepo.On("Update", mock.Anything, int64(1), userID, "Updated Name", (*string)(nil), (*int64)(nil), (*int64)(nil), &fulfillment, &method, (*int64)(nil), float64(0), float64(0)).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"name":                  "Updated Name",
		"transport_fulfillment": "self_haul",
		"transport_method":      "freighter",
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.UpdatePlan(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

func Test_ProductionPlans_UpdatePlan_WithCourierSettings(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	fulfillment := "courier_contract"

	mocks.plansRepo.On("Update", mock.Anything, int64(1), userID, "Courier Plan", (*string)(nil), (*int64)(nil), (*int64)(nil), &fulfillment, (*string)(nil), (*int64)(nil), float64(800), float64(0.02)).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"name":                    "Courier Plan",
		"transport_fulfillment":   "courier_contract",
		"courier_rate_per_m3":     800,
		"courier_collateral_rate": 0.02,
	})
	req := httptest.NewRequest("PUT", "/v1/industry/plans/1", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.UpdatePlan(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.plansRepo.AssertExpectations(t)
}

// --- Transport Generation Tests ---

func Test_ProductionPlans_GenerateJobs_NoTransportWhenFulfillmentNil(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID := int64(20)
	parentRef := rootStepID
	stationA := int64(5)
	stationB := int64(6)

	// Plan WITHOUT transport settings
	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter", UserStationID: &stationA,
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component", UserStationID: &stationB,
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 2,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
	}, nil)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, Activity: "manufacturing", Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	// Should have manufacturing jobs only, no transport
	assert.Len(t, genResult.TransportJobs, 0)
	// Transport repo should NOT have been called
	mocks.transportJobRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_ProductionPlans_GenerateJobs_NoTransportWhenSameStation(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID := int64(20)
	parentRef := rootStepID
	stationA := int64(5) // both steps at same user station
	fulfillment := "self_haul"
	method := "freighter"

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		TransportFulfillment: &fulfillment,
		TransportMethod:      &method,
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter", UserStationID: &stationA,
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component", UserStationID: &stationA,
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 2,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
	}, nil)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, Activity: "manufacturing", Status: "planned",
	}, nil)

	// Both steps resolve to same station_id=60003760
	mocks.stationRepo.On("GetByID", mock.Anything, stationA, userID).Return(&models.UserStation{
		ID: stationA, StationID: 60003760, SolarSystemID: 30000142,
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.TransportJobs, 0)
	mocks.transportJobRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_ProductionPlans_GenerateJobs_TransportCreatedDifferentStations(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID := int64(20)
	parentRef := rootStepID
	stationA := int64(5)
	stationB := int64(6)
	fulfillment := "courier_contract"

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		TransportFulfillment: &fulfillment,
		CourierRatePerM3:     800,
		CourierCollateralRate: 0.02,
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter", UserStationID: &stationA,
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component", UserStationID: &stationB,
			},
		},
	}

	sellPrice := 100.0
	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{
		5678: {TypeID: 5678, SellPrice: &sellPrice},
	}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 2,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
	}, nil)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	// Manufacturing queue entries
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, Activity: "manufacturing", Status: "planned",
	}, nil)

	// Station A (parent) and Station B (child) resolve to different station IDs
	mocks.stationRepo.On("GetByID", mock.Anything, stationA, userID).Return(&models.UserStation{
		ID: stationA, StationID: 60003760, SolarSystemID: 30000142,
		StationName: "Jita Station",
	}, nil)
	mocks.stationRepo.On("GetByID", mock.Anything, stationB, userID).Return(&models.UserStation{
		ID: stationB, StationID: 60004588, SolarSystemID: 30002187,
		StationName: "Amarr Station",
	}, nil)

	// Transport job creation
	transportJobID := int64(500)
	mocks.transportJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *models.TransportJob) bool {
		return j.OriginStationID == 60004588 && j.DestinationStationID == 60003760 &&
			j.FulfillmentType == "courier_contract" &&
			len(j.Items) == 1 && j.Items[0].TypeID == 5678
	})).Return(&models.TransportJob{
		ID:                   transportJobID,
		OriginStationID:      60004588,
		DestinationStationID: 60003760,
		FulfillmentType:      "courier_contract",
		EstimatedCost:        5000,
	}, nil)

	// Transport queue entry creation
	queueEntryID := int64(101)
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "transport" && e.TransportJobID != nil && *e.TransportJobID == transportJobID
	})).Return(&models.IndustryJobQueueEntry{
		ID: queueEntryID, Activity: "transport", Status: "planned",
	}, nil)

	// Link queue entry back
	mocks.transportJobRepo.On("SetQueueEntryID", mock.Anything, transportJobID, queueEntryID).Return(nil)

	body, _ := json.Marshal(map[string]any{"quantity": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.TransportJobs, 1)
	assert.Equal(t, int64(60004588), genResult.TransportJobs[0].OriginStationID)
	assert.Equal(t, int64(60003760), genResult.TransportJobs[0].DestinationStationID)
	// Transport queue entry should be in Created
	assert.True(t, len(genResult.Created) >= 3) // 2 manufacturing + 1 transport
	mocks.transportJobRepo.AssertExpectations(t)
}

func Test_ProductionPlans_GenerateJobs_TransportBatchedSameRoute(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID1 := int64(20)
	childStepID2 := int64(30)
	parentRef := rootStepID
	stationA := int64(5) // parent station
	stationB := int64(6) // both children at this station
	fulfillment := "courier_contract"

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		TransportFulfillment:  &fulfillment,
		CourierRatePerM3:      500,
		CourierCollateralRate:  0.01,
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter", UserStationID: &stationA,
			},
			{
				ID: childStepID1, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component A", UserStationID: &stationB,
			},
			{
				ID: childStepID2, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 9876, BlueprintTypeID: 4321,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component B", UserStationID: &stationB,
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 1,
	}, nil)

	// Root blueprint
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component A", Quantity: 10},
		{BlueprintTypeID: 787, TypeID: 9876, TypeName: "Component B", Quantity: 5},
	}, nil)

	// Child A blueprint
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component A",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	// Child B blueprint
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(4321), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 4321, ProductTypeID: 9876, ProductName: "Component B",
		ProductQuantity: 1, Time: 1800, ProductVolume: 20,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(4321), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	// Manufacturing queue entries
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, Activity: "manufacturing", Status: "planned",
	}, nil)

	// Stations
	mocks.stationRepo.On("GetByID", mock.Anything, stationA, userID).Return(&models.UserStation{
		ID: stationA, StationID: 60003760, SolarSystemID: 30000142,
		StationName: "Jita Station",
	}, nil)
	mocks.stationRepo.On("GetByID", mock.Anything, stationB, userID).Return(&models.UserStation{
		ID: stationB, StationID: 60004588, SolarSystemID: 30002187,
		StationName: "Amarr Station",
	}, nil)

	// Should create ONE transport job with 2 items (batched)
	transportJobID := int64(500)
	mocks.transportJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *models.TransportJob) bool {
		return len(j.Items) == 2 && j.FulfillmentType == "courier_contract"
	})).Return(&models.TransportJob{
		ID: transportJobID, EstimatedCost: 3000,
	}, nil)

	queueEntryID := int64(101)
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "transport"
	})).Return(&models.IndustryJobQueueEntry{
		ID: queueEntryID, Activity: "transport", Status: "planned",
	}, nil)

	mocks.transportJobRepo.On("SetQueueEntryID", mock.Anything, transportJobID, queueEntryID).Return(nil)

	body, _ := json.Marshal(map[string]any{"quantity": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	// Only 1 transport job (batched), not 2
	assert.Len(t, genResult.TransportJobs, 1)
	// Transport job repo should have been called exactly once
	mocks.transportJobRepo.AssertNumberOfCalls(t, "Create", 1)
}

func Test_ProductionPlans_GenerateJobs_SelfHaulGateTransport(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	rootStepID := int64(10)
	childStepID := int64(20)
	parentRef := rootStepID
	stationA := int64(5)
	stationB := int64(6)
	fulfillment := "self_haul"
	method := "freighter"
	profileID := int64(42)

	plan := &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		TransportFulfillment: &fulfillment,
		TransportMethod:      &method,
		TransportProfileID:   &profileID,
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter", UserStationID: &stationA,
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &parentRef,
				ProductTypeID: 5678, BlueprintTypeID: 1234,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Component", UserStationID: &stationB,
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: 100, Quantity: 2,
	}, nil)

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
	}, nil)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), mock.Anything).Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), mock.Anything).Return([]*repositories.ManufacturingMaterialRow{}, nil)

	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, Activity: "manufacturing", Status: "planned",
	}, nil)

	mocks.stationRepo.On("GetByID", mock.Anything, stationA, userID).Return(&models.UserStation{
		ID: stationA, StationID: 60003760, SolarSystemID: 30000142,
		StationName: "Jita Station",
	}, nil)
	mocks.stationRepo.On("GetByID", mock.Anything, stationB, userID).Return(&models.UserStation{
		ID: stationB, StationID: 60004588, SolarSystemID: 30002187,
		StationName: "Amarr Station",
	}, nil)

	// Profile lookup
	mocks.profilesRepo.On("GetByID", mock.Anything, profileID, userID).Return(&models.TransportProfile{
		ID:               profileID,
		TransportMethod:  "freighter",
		CargoM3:          60000,
		RatePerM3PerJump: 10,
		CollateralRate:   0.01,
		RoutePreference:  "shortest",
	}, nil)

	// ESI route call
	mocks.esiClient.On("GetRoute", mock.Anything, int64(30002187), int64(30000142), "shortest").Return(
		[]int32{30002187, 30002188, 30000142}, nil,
	)

	// Transport job creation
	transportJobID := int64(500)
	mocks.transportJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *models.TransportJob) bool {
		return j.TransportMethod == "freighter" && j.FulfillmentType == "self_haul" &&
			j.Jumps == 2 && j.TransportProfileID != nil && *j.TransportProfileID == profileID
	})).Return(&models.TransportJob{
		ID: transportJobID, EstimatedCost: 1000, Jumps: 2,
	}, nil)

	queueEntryID := int64(101)
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.Activity == "transport"
	})).Return(&models.IndustryJobQueueEntry{
		ID: queueEntryID, Activity: "transport", Status: "planned",
	}, nil)

	mocks.transportJobRepo.On("SetQueueEntryID", mock.Anything, transportJobID, queueEntryID).Return(nil)

	body, _ := json.Marshal(map[string]any{"quantity": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.TransportJobs, 1)
	assert.Equal(t, 2, genResult.TransportJobs[0].Jumps)
	mocks.esiClient.AssertExpectations(t)
	mocks.profilesRepo.AssertExpectations(t)
	mocks.transportJobRepo.AssertExpectations(t)
}

// --- PreviewPlan Tests ---

// skillVal is a convenience helper to build a CharacterSkill with ActiveLevel set.
func skillVal(charID, skillID int64, level int) *models.CharacterSkill {
	return &models.CharacterSkill{
		CharacterID: charID,
		SkillID:     skillID,
		ActiveLevel: level,
	}
}

func Test_ProductionPlans_PreviewPlan_Success(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	charA := int64(201)
	charB := int64(202)
	stepID := int64(10)
	childStepID := int64(20)

	// Plan with root step + one child step
	plan := &models.ProductionPlan{
		ID: 1, UserID: userID, Name: "Test Plan",
		Steps: []*models.ProductionPlanStep{
			{
				ID:               stepID,
				PlanID:           1,
				ProductTypeID:    587,
				BlueprintTypeID:  787,
				Activity:         "manufacturing",
				MELevel:          10,
				TELevel:          20,
				IndustrySkill:    5,
				AdvIndustrySkill: 5,
				Structure:        "raitaru",
				Rig:              "t2",
				Security:         "high",
				FacilityTax:      1.0,
				ProductName:      "Rifter",
			},
			{
				ID:               childStepID,
				PlanID:           1,
				ParentStepID:     &stepID,
				ProductTypeID:    5678,
				BlueprintTypeID:  1234,
				Activity:         "manufacturing",
				MELevel:          10,
				TELevel:          20,
				IndustrySkill:    5,
				AdvIndustrySkill: 5,
				Structure:        "raitaru",
				Rig:              "t2",
				Security:         "high",
				FacilityTax:      1.0,
				ProductName:      "Component",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	// Blueprint data for root step
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), "manufacturing").Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200, ProductVolume: 27500,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), "manufacturing").Return([]*repositories.ManufacturingMaterialRow{
		{BlueprintTypeID: 787, TypeID: 5678, TypeName: "Component", Quantity: 10},
	}, nil)

	// Blueprint data for child step
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1234), "manufacturing").Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 1234, ProductTypeID: 5678, ProductName: "Component",
		ProductQuantity: 5, Time: 3600, ProductVolume: 10,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1234), "manufacturing").Return([]*repositories.ManufacturingMaterialRow{}, nil)

	// Two eligible characters (both with Industry 5)
	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{
		charA: "Alpha",
		charB: "Beta",
	}, nil)

	// SkillID 3380=Industry, 3388=AdvIndustry, 3387=MassProduction
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return([]*models.CharacterSkill{
		skillVal(charA, 3380, 5), // Industry 5
		skillVal(charA, 3388, 5), // AdvIndustry 5
		skillVal(charA, 3387, 5), // MassProduction 5
		skillVal(charB, 3380, 4), // Industry 4
		skillVal(charB, 3388, 3), // AdvIndustry 3
		skillVal(charB, 3387, 3), // MassProduction 3
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/preview", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.PreviewPlan(args)

	assert.Nil(t, httpErr)
	preview := result.(*models.PlanPreviewResult)

	// Two eligible characters → two options
	assert.Equal(t, 2, preview.EligibleCharacters)
	assert.Equal(t, 2, preview.TotalJobs)
	assert.Len(t, preview.Options, 2)

	// Parallelism 1 and 2 are present
	assert.Equal(t, 1, preview.Options[0].Parallelism)
	assert.Equal(t, 2, preview.Options[1].Parallelism)

	// Parallelism 2 should not be slower than parallelism 1
	assert.LessOrEqual(t, preview.Options[1].EstimatedDurationSec, preview.Options[0].EstimatedDurationSec)

	// Character info is populated
	assert.Len(t, preview.Options[0].Characters, 1)
	assert.Len(t, preview.Options[1].Characters, 2)

	// Duration label is non-empty
	assert.NotEmpty(t, preview.Options[0].EstimatedDurationLabel)

	mocks.plansRepo.AssertExpectations(t)
	mocks.marketRepo.AssertExpectations(t)
	mocks.sdeRepo.AssertExpectations(t)
	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
}

func Test_ProductionPlans_PreviewPlan_NoCharacters(t *testing.T) {
	controller, mocks := setupProductionPlansController()

	userID := int64(100)
	stepID := int64(10)

	plan := &models.ProductionPlan{
		ID: 1, UserID: userID, Name: "Test Plan",
		Steps: []*models.ProductionPlanStep{
			{
				ID: stepID, PlanID: 1, ProductTypeID: 587,
				BlueprintTypeID: 787, Activity: "manufacturing",
				MELevel: 10, TELevel: 20, IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)

	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), "manufacturing").Return(&repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
		ProductQuantity: 1, Time: 7200,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), "manufacturing").Return([]*repositories.ManufacturingMaterialRow{}, nil)

	// No characters at all
	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{}, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return([]*models.CharacterSkill{}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/preview", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.PreviewPlan(args)

	assert.Nil(t, httpErr)
	preview := result.(*models.PlanPreviewResult)
	assert.Equal(t, 0, preview.EligibleCharacters)
	assert.Len(t, preview.Options, 0)
}

func Test_ProductionPlans_PreviewPlan_InvalidQuantity(t *testing.T) {
	controller, _ := setupProductionPlansController()

	userID := int64(100)

	body, _ := json.Marshal(map[string]any{"quantity": 0})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/preview", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.PreviewPlan(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- GenerateJobs Parallelism Tests ---

// buildParallelismPlan builds a simple single-step manufacturing plan for parallelism tests.
func buildParallelismPlan() *models.ProductionPlan {
	return &models.ProductionPlan{
		ID: 1, UserID: 100, ProductTypeID: 587, Name: "Rifter",
		Steps: []*models.ProductionPlanStep{
			{
				ID: 10, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 10, TELevel: 20,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "t2", Security: "high", FacilityTax: 1.0,
				ProductName: "Rifter",
			},
		},
	}
}

// setupParallelismMocks sets up the common SDE/market/runs mocks used by all parallelism tests.
func setupParallelismMocks(mocks *productionPlanMocks, userID int64, quantity int) {
	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(buildParallelismPlan(), nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: userID, Quantity: quantity,
	}, nil)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), mock.Anything).Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Rifter",
			ProductQuantity: 1, Time: 7200,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), mock.Anything).Return(
		[]*repositories.ManufacturingMaterialRow{
			{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 1000},
		}, nil)
}

// Test_ProductionPlans_GenerateJobs_WithParallelism0_BackwardCompat verifies that
// parallelism=0 (the default) preserves the original behaviour: no character
// assignment, CharacterID nil on all entries, CharacterAssignments nil.
func Test_ProductionPlans_GenerateJobs_WithParallelism0_BackwardCompat(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)

	setupParallelismMocks(mocks, userID, 5)

	// parallelism=0 must NOT call character/skills/slotUsage repos
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: userID, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 5, Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 5, "parallelism": 0})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.Created, 1)
	assert.Nil(t, genResult.CharacterAssignments)
	assert.Equal(t, 0, genResult.UnassignedCount)
	// CharacterID must not be set (original behaviour)
	assert.Nil(t, genResult.Created[0].CharacterID)

	// character/skills repos must not have been called
	mocks.characterRepo.AssertNotCalled(t, "GetNames", mock.Anything, mock.Anything)
	mocks.skillsRepo.AssertNotCalled(t, "GetSkillsForUser", mock.Anything, mock.Anything)
	mocks.queueRepo.AssertNotCalled(t, "GetSlotUsage", mock.Anything, mock.Anything)
}

// Test_ProductionPlans_GenerateJobs_WithParallelism1 verifies that parallelism=1
// assigns all jobs to the best eligible character and populates CharacterAssignments.
func Test_ProductionPlans_GenerateJobs_WithParallelism1(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)
	charA := int64(201)
	charB := int64(202)

	setupParallelismMocks(mocks, userID, 10)

	characterNames := map[int64]string{charA: "Alpha", charB: "Beta"}
	skills := []*models.CharacterSkill{
		skillVal(charA, 3380, 5), // Industry 5
		skillVal(charA, 3388, 5), // AdvIndustry 5
		skillVal(charA, 3387, 5), // MassProduction 5
		skillVal(charB, 3380, 3), // Industry 3
		skillVal(charB, 3387, 2), // MassProduction 2
	}

	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(characterNames, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return(skills, nil)
	mocks.queueRepo.On("GetSlotUsage", mock.Anything, userID).Return(map[int64]map[string]int{}, nil)

	// With parallelism=1 the best character gets all runs
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787 && e.Runs == 10
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: userID, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 10, Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 10, "parallelism": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Len(t, genResult.Created, 1)
	assert.Equal(t, 0, genResult.UnassignedCount)
	assert.NotNil(t, genResult.CharacterAssignments)
	// Exactly one character assigned (the best one: charA with highest skills)
	assert.Len(t, genResult.CharacterAssignments, 1)

	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
	mocks.queueRepo.AssertExpectations(t)
}

// Test_ProductionPlans_GenerateJobs_WithParallelism2_SplitsRuns verifies that
// parallelism=2 with 2 eligible characters splits a 10-run job across both.
func Test_ProductionPlans_GenerateJobs_WithParallelism2_SplitsRuns(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)
	charA := int64(201)
	charB := int64(202)

	setupParallelismMocks(mocks, userID, 10)

	characterNames := map[int64]string{charA: "Alpha", charB: "Beta"}
	skills := []*models.CharacterSkill{
		skillVal(charA, 3380, 5), // Industry 5
		skillVal(charA, 3388, 5), // AdvIndustry 5
		skillVal(charA, 3387, 5), // MassProduction 5
		skillVal(charB, 3380, 3), // Industry 3
		skillVal(charB, 3387, 2), // MassProduction 2
	}

	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(characterNames, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return(skills, nil)
	mocks.queueRepo.On("GetSlotUsage", mock.Anything, userID).Return(map[int64]map[string]int{}, nil)

	// With parallelism=2, simulateAssignment splits 10 runs: ceil(10/2)=5 for first, 5 for second
	// So we expect two Create calls with runs 5 each (or 5+5)
	totalRunsCreated := 0
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787 && e.Activity == "manufacturing"
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: userID, BlueprintTypeID: 787, Activity: "manufacturing", Status: "planned",
	}, nil).Run(func(args mock.Arguments) {
		entry := args.Get(1).(*models.IndustryJobQueueEntry)
		totalRunsCreated += entry.Runs
	})

	body, _ := json.Marshal(map[string]any{"quantity": 10, "parallelism": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	assert.Equal(t, 0, genResult.UnassignedCount)
	assert.NotNil(t, genResult.CharacterAssignments)
	// Two characters should be assigned
	assert.Len(t, genResult.CharacterAssignments, 2)
	// Two queue entries (one per character fragment)
	assert.Len(t, genResult.Created, 2)
	// Runs must sum to 10
	assert.Equal(t, 10, totalRunsCreated)

	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
	mocks.queueRepo.AssertExpectations(t)
}

// Test_ProductionPlans_GenerateJobs_WithParallelism_NoEligibleCharacters verifies that
// when parallelism >= 1 but no characters have industry skills, jobs are still created
// in the queue with a nil CharacterID (unassigned) so they are not silently dropped.
func Test_ProductionPlans_GenerateJobs_WithParallelism_NoEligibleCharacters(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)
	charA := int64(201)

	setupParallelismMocks(mocks, userID, 5)

	// Character exists but has no industry skills → BuildCharacterCapacities excludes them.
	// simulateAssignment gets an empty pool (parallelism capped to 0) so all jobs are
	// unassigned (characterID=0) but still appended to the assigned slice.
	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{charA: "NoSkills"}, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return([]*models.CharacterSkill{}, nil)
	mocks.queueRepo.On("GetSlotUsage", mock.Anything, userID).Return(map[int64]map[string]int{}, nil)

	// The job must still be created with CharacterID=nil.
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787 && e.CharacterID == nil
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: userID, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 5, Status: "planned",
	}, nil)

	body, _ := json.Marshal(map[string]any{"quantity": 5, "parallelism": 2})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(args)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)
	// unassigned count reflects all runs are unassigned
	assert.Greater(t, genResult.UnassignedCount, 0)
	// One entry created with nil CharacterID (bug 1 fix: not silently dropped)
	assert.Len(t, genResult.Created, 1)
	assert.Nil(t, genResult.Created[0].CharacterID)

	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
	mocks.queueRepo.AssertExpectations(t)
}

// Test_SimulateAssignment_UnassignedJobsStillCreated verifies that when a character
// has fewer slots than there are jobs at the same depth, the overflow jobs are still
// created as queue entries with nil CharacterID rather than being silently dropped.
//
// Scenario: 1 character with 1 mfg slot, plan with root + 2 child steps (both depth 1).
// The first child consumes the only slot; the second child has no eligible character.
// Both children must appear in Created: one with a CharacterID, one with nil.
// The root step at depth 0 gets the recycled slot and must also appear in Created.
func Test_SimulateAssignment_UnassignedJobsStillCreated(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)
	charA := int64(201)
	rootStepID := int64(10)
	childStep1ID := int64(20)
	childStep2ID := int64(30)

	// Plan: root step produces item 587 using materials 111 and 222.
	// Child step 1 produces material 111; child step 2 produces material 222.
	plan := &models.ProductionPlan{
		ID: 1, UserID: userID, Name: "Slot Overflow Plan",
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 0, TELevel: 0,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "none", Security: "high", FacilityTax: 1.0,
				ProductName: "Root Product",
			},
			{
				ID: childStep1ID, PlanID: 1, ParentStepID: &rootStepID,
				ProductTypeID: 111, BlueprintTypeID: 1001,
				Activity: "manufacturing", MELevel: 0, TELevel: 0,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "none", Security: "high", FacilityTax: 1.0,
				ProductName: "Child Product A",
			},
			{
				ID: childStep2ID, PlanID: 1, ParentStepID: &rootStepID,
				ProductTypeID: 222, BlueprintTypeID: 1002,
				Activity: "manufacturing", MELevel: 0, TELevel: 0,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "none", Security: "high", FacilityTax: 1.0,
				ProductName: "Child Product B",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: userID, Quantity: 1,
	}, nil)

	// Root blueprint: produces 1 item 587, materials are 111 (qty 1) and 222 (qty 1)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), "manufacturing").Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Root Product",
			ProductQuantity: 1, Time: 3600,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), "manufacturing").Return(
		[]*repositories.ManufacturingMaterialRow{
			{BlueprintTypeID: 787, TypeID: 111, TypeName: "Child Product A", Quantity: 1},
			{BlueprintTypeID: 787, TypeID: 222, TypeName: "Child Product B", Quantity: 1},
		}, nil)

	// Child blueprint 1: produces 1 item 111
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1001), "manufacturing").Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 1001, ProductTypeID: 111, ProductName: "Child Product A",
			ProductQuantity: 1, Time: 1800,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1001), "manufacturing").Return(
		[]*repositories.ManufacturingMaterialRow{}, nil)

	// Child blueprint 2: produces 1 item 222
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1002), "manufacturing").Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 1002, ProductTypeID: 222, ProductName: "Child Product B",
			ProductQuantity: 1, Time: 1800,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1002), "manufacturing").Return(
		[]*repositories.ManufacturingMaterialRow{}, nil)

	// Character with Industry=5, no MassProduction → exactly 1 mfg slot (base slot only).
	// This ensures the second child job at depth 1 has no eligible character.
	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{charA: "Alpha"}, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return([]*models.CharacterSkill{
		skillVal(charA, 3380, 5), // Industry 5
		skillVal(charA, 3388, 5), // AdvIndustry 5
		// No MassProduction → 1 slot total
	}, nil)
	mocks.queueRepo.On("GetSlotUsage", mock.Anything, userID).Return(map[int64]map[string]int{}, nil)

	// Track CharacterID values from Create calls to verify the mix of assigned/unassigned.
	var capturedCharIDs []*int64
	mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
		return e.BlueprintTypeID == 787 || e.BlueprintTypeID == 1001 || e.BlueprintTypeID == 1002
	})).Return(&models.IndustryJobQueueEntry{
		ID: 99, UserID: userID, Activity: "manufacturing", Runs: 1, Status: "planned",
	}, nil).Run(func(args mock.Arguments) {
		entry := args.Get(1).(*models.IndustryJobQueueEntry)
		capturedCharIDs = append(capturedCharIDs, entry.CharacterID)
	}).Times(3)

	body, _ := json.Marshal(map[string]any{"quantity": 1, "parallelism": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	handlerArgs := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(handlerArgs)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)

	// All 3 steps must produce a queue entry (Bug 1: unassigned jobs must not be dropped)
	assert.Len(t, genResult.Created, 3, "all 3 steps must appear in Created even when one is unassigned")

	// Count assigned vs unassigned in the captured Create arguments
	nilCount := 0
	nonNilCount := 0
	for _, charID := range capturedCharIDs {
		if charID == nil {
			nilCount++
		} else {
			nonNilCount++
		}
	}

	// Exactly one child job is unassigned (nil CharacterID)
	assert.Equal(t, 1, nilCount, "exactly one job should be unassigned (nil CharacterID)")
	// Two jobs are assigned: the first child + the root (slot recycled from depth 1 to depth 0)
	assert.Equal(t, 2, nonNilCount, "two jobs should be assigned (Bug 2 fix: slot recycled for root)")

	// UnassignedCount reflects the one unassigned child
	assert.Greater(t, genResult.UnassignedCount, 0)

	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
	mocks.queueRepo.AssertExpectations(t)
}

// Test_SimulateAssignment_SlotRecyclingAcrossDepths verifies that slot availability is
// reset when the depth level changes so a character's slots are reused across depth levels.
//
// Scenario: 1 character with 1 mfg slot, plan with root (depth 0) + 1 child (depth 1).
// Without recycling: child consumes slot → parent has no slot → parent unassigned (1 Created).
// With recycling: child consumes slot → depth changes → slot resets → parent gets slot (2 Created).
func Test_SimulateAssignment_SlotRecyclingAcrossDepths(t *testing.T) {
	controller, mocks := setupProductionPlansController()
	userID := int64(100)
	charA := int64(201)
	rootStepID := int64(10)
	childStepID := int64(20)

	plan := &models.ProductionPlan{
		ID: 1, UserID: userID, Name: "Depth Recycle Plan",
		Steps: []*models.ProductionPlanStep{
			{
				ID: rootStepID, PlanID: 1, ProductTypeID: 587, BlueprintTypeID: 787,
				Activity: "manufacturing", MELevel: 0, TELevel: 0,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "none", Security: "high", FacilityTax: 1.0,
				ProductName: "Root Product",
			},
			{
				ID: childStepID, PlanID: 1, ParentStepID: &rootStepID,
				ProductTypeID: 111, BlueprintTypeID: 1001,
				Activity: "manufacturing", MELevel: 0, TELevel: 0,
				IndustrySkill: 5, AdvIndustrySkill: 5,
				Structure: "raitaru", Rig: "none", Security: "high", FacilityTax: 1.0,
				ProductName: "Child Product",
			},
		},
	}

	mocks.plansRepo.On("GetByID", mock.Anything, int64(1), userID).Return(plan, nil)
	mocks.marketRepo.On("GetAllJitaPrices", mock.Anything).Return(map[int64]*models.MarketPrice{}, nil)
	mocks.marketRepo.On("GetAllAdjustedPrices", mock.Anything).Return(map[int64]float64{}, nil)
	mocks.runsRepo.On("Create", mock.Anything, mock.Anything).Return(&models.ProductionPlanRun{
		ID: 50, PlanID: 1, UserID: userID, Quantity: 1,
	}, nil)

	// Root blueprint: produces 1 item 587, material 111 (qty 1)
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(787), "manufacturing").Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 787, ProductTypeID: 587, ProductName: "Root Product",
			ProductQuantity: 1, Time: 3600,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(787), "manufacturing").Return(
		[]*repositories.ManufacturingMaterialRow{
			{BlueprintTypeID: 787, TypeID: 111, TypeName: "Child Product", Quantity: 1},
		}, nil)

	// Child blueprint: produces 1 item 111
	mocks.sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(1001), "manufacturing").Return(
		&repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 1001, ProductTypeID: 111, ProductName: "Child Product",
			ProductQuantity: 1, Time: 1800,
		}, nil)
	mocks.sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(1001), "manufacturing").Return(
		[]*repositories.ManufacturingMaterialRow{}, nil)

	// Character with exactly 1 mfg slot (Industry >= 1, no MassProduction)
	mocks.characterRepo.On("GetNames", mock.Anything, userID).Return(map[int64]string{charA: "Alpha"}, nil)
	mocks.skillsRepo.On("GetSkillsForUser", mock.Anything, userID).Return([]*models.CharacterSkill{
		skillVal(charA, 3380, 5), // Industry 5
		skillVal(charA, 3388, 5), // AdvIndustry 5
		// No MassProduction → 1 slot total
	}, nil)
	mocks.queueRepo.On("GetSlotUsage", mock.Anything, userID).Return(map[int64]map[string]int{}, nil)

	// Both the child (BP 1001) and the root (BP 787) must be created with a non-nil CharacterID.
	for _, bpID := range []int64{787, 1001} {
		bpIDCopy := bpID
		mocks.queueRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *models.IndustryJobQueueEntry) bool {
			return e.BlueprintTypeID == bpIDCopy
		})).Return(&models.IndustryJobQueueEntry{
			ID:          bpIDCopy,
			UserID:      userID,
			Activity:    "manufacturing",
			Runs:        1,
			Status:      "planned",
			CharacterID: &charA,
		}, nil)
	}

	body, _ := json.Marshal(map[string]any{"quantity": 1, "parallelism": 1})
	req := httptest.NewRequest("POST", "/v1/industry/plans/1/generate", bytes.NewReader(body))
	handlerArgs := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "1"}}

	result, httpErr := controller.GenerateJobs(handlerArgs)

	assert.Nil(t, httpErr)
	genResult := result.(*models.GenerateJobsResult)

	// Both root and child must be created (Bug 2 fix: slot recycled across depth levels)
	assert.Len(t, genResult.Created, 2, "both root and child steps must produce a queue entry with slot recycling")

	// No unassigned jobs: the slot was recycled between depth levels
	assert.Equal(t, 0, genResult.UnassignedCount, "no jobs should be unassigned when slots are recycled")

	mocks.characterRepo.AssertExpectations(t)
	mocks.skillsRepo.AssertExpectations(t)
	mocks.queueRepo.AssertExpectations(t)
}
