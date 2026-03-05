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
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock HaulingRunsRepository ---

type MockHaulingRunsRepository struct {
	mock.Mock
}

func (m *MockHaulingRunsRepository) CreateRun(ctx context.Context, run *models.HaulingRun) (*models.HaulingRun, error) {
	args := m.Called(ctx, run)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HaulingRun), args.Error(1)
}

func (m *MockHaulingRunsRepository) GetRunByID(ctx context.Context, id int64, userID int64) (*models.HaulingRun, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HaulingRun), args.Error(1)
}

func (m *MockHaulingRunsRepository) ListRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRun), args.Error(1)
}

func (m *MockHaulingRunsRepository) UpdateRun(ctx context.Context, run *models.HaulingRun) error {
	args := m.Called(ctx, run)
	return args.Error(0)
}

func (m *MockHaulingRunsRepository) UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error {
	args := m.Called(ctx, id, userID, status)
	return args.Error(0)
}

func (m *MockHaulingRunsRepository) DeleteRun(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// --- Mock HaulingRunItemsRepository ---

type MockHaulingRunItemsRepository struct {
	mock.Mock
}

func (m *MockHaulingRunItemsRepository) AddItem(ctx context.Context, item *models.HaulingRunItem) (*models.HaulingRunItem, error) {
	args := m.Called(ctx, item)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HaulingRunItem), args.Error(1)
}

func (m *MockHaulingRunItemsRepository) GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRunItem), args.Error(1)
}

func (m *MockHaulingRunItemsRepository) UpdateItemAcquired(ctx context.Context, itemID int64, runID int64, quantityAcquired int64) error {
	args := m.Called(ctx, itemID, runID, quantityAcquired)
	return args.Error(0)
}

func (m *MockHaulingRunItemsRepository) RemoveItem(ctx context.Context, itemID int64, runID int64) error {
	args := m.Called(ctx, itemID, runID)
	return args.Error(0)
}

// --- Mock HaulingMarketRepository ---

type MockHaulingMarketRepo struct {
	mock.Mock
}

func (m *MockHaulingMarketRepo) GetScannerResults(ctx context.Context, sourceRegionID int64, sourceSystemID int64, destRegionID int64) ([]*models.HaulingArbitrageRow, error) {
	args := m.Called(ctx, sourceRegionID, sourceSystemID, destRegionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingArbitrageRow), args.Error(1)
}

// --- Mock HaulingMarketUpdater ---

type MockHaulingMarketUpdater struct {
	mock.Mock
}

func (m *MockHaulingMarketUpdater) ScanRegion(ctx context.Context, regionID int64, systemID int64) error {
	args := m.Called(ctx, regionID, systemID)
	return args.Error(0)
}

func (m *MockHaulingMarketUpdater) ScanStructure(ctx context.Context, structureID int64, token string) (bool, error) {
	args := m.Called(ctx, structureID, token)
	return args.Bool(0), args.Error(1)
}

// --- Mock HaulingStructureRepository ---

type MockHaulingStructureRepository struct {
	mock.Mock
}

func (m *MockHaulingStructureRepository) GetStructureScannerResults(ctx context.Context, sourceStructureID int64, destRegionID int64, destSystemID int64) ([]*models.HaulingArbitrageRow, error) {
	args := m.Called(ctx, sourceStructureID, destRegionID, destSystemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingArbitrageRow), args.Error(1)
}

func (m *MockHaulingStructureRepository) GetRegionToStructureResults(ctx context.Context, srcRegionID int64, srcSystemID int64, destStructureID int64) ([]*models.HaulingArbitrageRow, error) {
	args := m.Called(ctx, srcRegionID, srcSystemID, destStructureID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingArbitrageRow), args.Error(1)
}

// --- Mock HaulingPnlRepository ---

type MockHaulingPnlRepository struct {
	mock.Mock
}

func (m *MockHaulingPnlRepository) UpsertPnlEntry(ctx context.Context, entry *models.HaulingRunPnlEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockHaulingPnlRepository) GetPnlByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunPnlEntry, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRunPnlEntry), args.Error(1)
}

func (m *MockHaulingPnlRepository) GetPnlSummaryByRunID(ctx context.Context, runID int64) (*models.HaulingRunPnlSummary, error) {
	args := m.Called(ctx, runID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HaulingRunPnlSummary), args.Error(1)
}

// --- Mock HaulingRunNotifier ---

type MockHaulingRunNotifier struct {
	mock.Mock
}

func (m *MockHaulingRunNotifier) NotifyHaulingTier2(ctx context.Context, userID int64, run *models.HaulingRun, fillPct float64) {
	m.Called(ctx, userID, run, fillPct)
}

func (m *MockHaulingRunNotifier) NotifyHaulingComplete(ctx context.Context, userID int64, run *models.HaulingRun, summary *models.HaulingRunPnlSummary) {
	m.Called(ctx, userID, run, summary)
}

func (m *MockHaulingRunNotifier) SendHaulingDailyDigest(ctx context.Context, userID int64, runs []*models.HaulingRun) {
	m.Called(ctx, userID, runs)
}

// --- Mock HaulingAnalyticsRepository ---

type MockHaulingAnalyticsRepository struct {
	mock.Mock
}

func (m *MockHaulingAnalyticsRepository) GetRouteAnalytics(ctx context.Context, userID int64) ([]*models.HaulingRouteAnalytics, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRouteAnalytics), args.Error(1)
}

func (m *MockHaulingAnalyticsRepository) GetItemAnalytics(ctx context.Context, userID int64) ([]*models.HaulingItemAnalytics, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingItemAnalytics), args.Error(1)
}

func (m *MockHaulingAnalyticsRepository) GetProfitTimeSeries(ctx context.Context, userID int64) ([]*models.HaulingProfitDataPoint, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingProfitDataPoint), args.Error(1)
}

func (m *MockHaulingAnalyticsRepository) GetRunDurationSummary(ctx context.Context, userID int64) (*models.HaulingRunDurationSummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.HaulingRunDurationSummary), args.Error(1)
}

func (m *MockHaulingAnalyticsRepository) GetCompletedRuns(ctx context.Context, userID int64, limit, offset int) ([]*models.HaulingRun, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.HaulingRun), args.Get(1).(int64), args.Error(2)
}

// --- Setup helper ---

type haulingMocks struct {
	runs       *MockHaulingRunsRepository
	items      *MockHaulingRunItemsRepository
	market     *MockHaulingMarketRepo
	structures *MockHaulingStructureRepository
	scanner    *MockHaulingMarketUpdater
	pnl        *MockHaulingPnlRepository
	notifier   *MockHaulingRunNotifier
	analytics  *MockHaulingAnalyticsRepository
}

func setupHaulingController() (*controllers.HaulingRunsController, haulingMocks) {
	mocks := haulingMocks{
		runs:       new(MockHaulingRunsRepository),
		items:      new(MockHaulingRunItemsRepository),
		market:     new(MockHaulingMarketRepo),
		structures: new(MockHaulingStructureRepository),
		scanner:    new(MockHaulingMarketUpdater),
		pnl:        new(MockHaulingPnlRepository),
		notifier:   nil, // default: no notifier
		analytics:  new(MockHaulingAnalyticsRepository),
	}
	router := &MockRouter{}
	controller := controllers.NewHaulingRuns(router, mocks.runs, mocks.items, mocks.market, mocks.structures, mocks.scanner, mocks.pnl, nil, mocks.analytics)
	return controller, mocks
}

func setupHaulingControllerWithNotifier() (*controllers.HaulingRunsController, haulingMocks) {
	mocks := haulingMocks{
		runs:       new(MockHaulingRunsRepository),
		items:      new(MockHaulingRunItemsRepository),
		market:     new(MockHaulingMarketRepo),
		structures: new(MockHaulingStructureRepository),
		scanner:    new(MockHaulingMarketUpdater),
		pnl:        new(MockHaulingPnlRepository),
		notifier:   new(MockHaulingRunNotifier),
		analytics:  new(MockHaulingAnalyticsRepository),
	}
	router := &MockRouter{}
	controller := controllers.NewHaulingRuns(router, mocks.runs, mocks.items, mocks.market, mocks.structures, mocks.scanner, mocks.pnl, mocks.notifier, mocks.analytics)
	return controller, mocks
}

// --- Tests: ListRuns ---

func Test_HaulingRuns_ListRuns_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expectedRuns := []*models.HaulingRun{
		{ID: int64(1), Name: "Run 1", Status: "PLANNING", Items: []*models.HaulingRunItem{}},
		{ID: int64(2), Name: "Run 2", Status: "ACCUMULATING", Items: []*models.HaulingRunItem{}},
	}
	mocks.runs.On("ListRunsByUser", mock.Anything, userID).Return(expectedRuns, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListRuns(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	runs := result.([]*models.HaulingRun)
	assert.Len(t, runs, 2)
	mocks.runs.AssertExpectations(t)
}

func Test_HaulingRuns_ListRuns_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("ListRunsByUser", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/runs", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.ListRuns(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mocks.runs.AssertExpectations(t)
}

// --- Tests: CreateRun ---

func Test_HaulingRuns_CreateRun_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	bodyRun := models.HaulingRun{
		Name:         "New Run",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	}
	createdRun := &models.HaulingRun{
		ID:           int64(1),
		UserID:       userID,
		Name:         "New Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		Items:        []*models.HaulingRunItem{},
		CreatedAt:    "2026-03-01T00:00:00Z",
		UpdatedAt:    "2026-03-01T00:00:00Z",
	}

	mocks.runs.On("CreateRun", mock.Anything, mock.AnythingOfType("*models.HaulingRun")).Return(createdRun, nil)

	body, _ := json.Marshal(bodyRun)
	req := httptest.NewRequest("POST", "/v1/hauling/runs", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.CreateRun(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	run := result.(*models.HaulingRun)
	assert.Equal(t, int64(1), run.ID)
	assert.Equal(t, "PLANNING", run.Status)
	assert.NotNil(t, run.Items)
	mocks.runs.AssertExpectations(t)
}

func Test_HaulingRuns_CreateRun_InvalidBody(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("POST", "/v1/hauling/runs", bytes.NewReader([]byte("not json")))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.CreateRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_CreateRun_RepoError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("CreateRun", mock.Anything, mock.AnythingOfType("*models.HaulingRun")).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(models.HaulingRun{Name: "Run", FromRegionID: int64(1), ToRegionID: int64(2)})
	req := httptest.NewRequest("POST", "/v1/hauling/runs", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.CreateRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: GetRun ---

func Test_HaulingRuns_GetRun_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), Name: "My Run", Status: "PLANNING", Items: []*models.HaulingRunItem{}}
	items := []*models.HaulingRunItem{{ID: int64(1), TypeID: int64(34), TypeName: "Tritanium"}}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.items.On("GetItemsByRunID", mock.Anything, int64(5)).Return(items, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetRun(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	r := result.(*models.HaulingRun)
	assert.Equal(t, int64(5), r.ID)
	assert.Len(t, r.Items, 1)
	mocks.runs.AssertExpectations(t)
	mocks.items.AssertExpectations(t)
}

func Test_HaulingRuns_GetRun_NotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_HaulingRuns_GetRun_InvalidID(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/abc", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.GetRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Tests: UpdateRun ---

func Test_HaulingRuns_UpdateRun_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("UpdateRun", mock.Anything, mock.AnythingOfType("*models.HaulingRun")).Return(nil)

	body, _ := json.Marshal(models.HaulingRun{Name: "Updated", FromRegionID: int64(1), ToRegionID: int64(2)})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateRun(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.runs.AssertExpectations(t)
}

func Test_HaulingRuns_UpdateRun_InvalidID(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("PUT", "/v1/hauling/runs/abc", bytes.NewReader([]byte("{}")))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.UpdateRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_UpdateRun_RepoError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("UpdateRun", mock.Anything, mock.AnythingOfType("*models.HaulingRun")).Return(errors.New("not found"))

	body, _ := json.Marshal(models.HaulingRun{Name: "Updated", FromRegionID: int64(1), ToRegionID: int64(2)})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: UpdateStatus ---

func Test_HaulingRuns_UpdateStatus_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("UpdateRunStatus", mock.Anything, int64(5), userID, "ACCUMULATING").Return(nil)

	body, _ := json.Marshal(map[string]string{"status": "ACCUMULATING"})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/status", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateStatus(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.runs.AssertExpectations(t)
}

func Test_HaulingRuns_UpdateStatus_InvalidStatus(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]string{"status": "INVALID_STATUS"})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/status", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateStatus(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_UpdateStatus_AllValidStatuses(t *testing.T) {
	validStatuses := []string{"PLANNING", "ACCUMULATING", "READY", "IN_TRANSIT", "SELLING", "COMPLETE", "CANCELLED"}
	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			controller, mocks := setupHaulingController()
			userID := int64(100)

			mocks.runs.On("UpdateRunStatus", mock.Anything, int64(5), userID, status).Return(nil)

			body, _ := json.Marshal(map[string]string{"status": status})
			req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/status", bytes.NewReader(body))
			args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

			result, httpErr := controller.UpdateStatus(args)
			assert.Nil(t, httpErr)
			assert.Nil(t, result)
			mocks.runs.AssertExpectations(t)
		})
	}
}

// --- Tests: DeleteRun ---

func Test_HaulingRuns_DeleteRun_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("DeleteRun", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/hauling/runs/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.DeleteRun(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.runs.AssertExpectations(t)
}

func Test_HaulingRuns_DeleteRun_NotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("DeleteRun", mock.Anything, int64(5), userID).Return(errors.New("hauling run not found"))

	req := httptest.NewRequest("DELETE", "/v1/hauling/runs/5", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.DeleteRun(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: AddItem ---

func Test_HaulingRuns_AddItem_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "PLANNING", Items: []*models.HaulingRunItem{}}
	createdItem := &models.HaulingRunItem{ID: int64(1), RunID: int64(5), TypeID: int64(34), TypeName: "Tritanium", QuantityPlanned: int64(100)}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.items.On("AddItem", mock.Anything, mock.AnythingOfType("*models.HaulingRunItem")).Return(createdItem, nil)

	body, _ := json.Marshal(models.HaulingRunItem{TypeID: int64(34), TypeName: "Tritanium", QuantityPlanned: int64(100)})
	req := httptest.NewRequest("POST", "/v1/hauling/runs/5/items", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.AddItem(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	item := result.(*models.HaulingRunItem)
	assert.Equal(t, int64(1), item.ID)
	mocks.runs.AssertExpectations(t)
	mocks.items.AssertExpectations(t)
}

func Test_HaulingRuns_AddItem_RunNotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	body, _ := json.Marshal(models.HaulingRunItem{TypeID: int64(34), TypeName: "Tritanium", QuantityPlanned: int64(100)})
	req := httptest.NewRequest("POST", "/v1/hauling/runs/5/items", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.AddItem(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- Tests: UpdateItemAcquired ---

func Test_HaulingRuns_UpdateItemAcquired_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.items.On("UpdateItemAcquired", mock.Anything, int64(10), int64(5), int64(50)).Return(nil)

	body, _ := json.Marshal(map[string]int64{"quantityAcquired": 50})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/items/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5", "itemId": "10"}}

	result, httpErr := controller.UpdateItemAcquired(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.items.AssertExpectations(t)
}

func Test_HaulingRuns_UpdateItemAcquired_NegativeQuantity(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]int64{"quantityAcquired": -5})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/items/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5", "itemId": "10"}}

	result, httpErr := controller.UpdateItemAcquired(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_UpdateItemAcquired_InvalidRunID(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]int64{"quantityAcquired": 10})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/abc/items/10", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc", "itemId": "10"}}

	result, httpErr := controller.UpdateItemAcquired(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Tests: RemoveItem ---

func Test_HaulingRuns_RemoveItem_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Items: []*models.HaulingRunItem{}}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.items.On("RemoveItem", mock.Anything, int64(10), int64(5)).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/hauling/runs/5/items/10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5", "itemId": "10"}}

	result, httpErr := controller.RemoveItem(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mocks.runs.AssertExpectations(t)
	mocks.items.AssertExpectations(t)
}

func Test_HaulingRuns_RemoveItem_RunNotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	req := httptest.NewRequest("DELETE", "/v1/hauling/runs/5/items/10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5", "itemId": "10"}}

	result, httpErr := controller.RemoveItem(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- Tests: GetScannerResults ---

func Test_HaulingRuns_GetScannerResults_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	buyPrice := 1000.0
	sellPrice := 1200.0
	spread := 0.2
	net := 200.0
	expectedRows := []*models.HaulingArbitrageRow{
		{
			TypeID:       int64(34),
			TypeName:     "Tritanium",
			BuyPrice:     &buyPrice,
			SellPrice:    &sellPrice,
			Spread:       &spread,
			NetProfitISK: &net,
			Indicator:    "gap",
		},
	}

	mocks.market.On("GetScannerResults", mock.Anything, int64(10000002), int64(0), int64(10000043)).Return(expectedRows, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=10000002&dest_region_id=10000043", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rows := result.([]*models.HaulingArbitrageRow)
	assert.Len(t, rows, 1)
	mocks.market.AssertExpectations(t)
}

func Test_HaulingRuns_GetScannerResults_WithSystem(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.market.On("GetScannerResults", mock.Anything, int64(10000002), int64(30000142), int64(10000043)).Return([]*models.HaulingArbitrageRow{}, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=10000002&dest_region_id=10000043&source_system_id=30000142", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.market.AssertExpectations(t)
}

func Test_HaulingRuns_GetScannerResults_MissingParams(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=10000002", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetScannerResults_InvalidSourceRegion(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=abc&dest_region_id=10000043", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetScannerResults_RepoError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.market.On("GetScannerResults", mock.Anything, int64(10000002), int64(0), int64(10000043)).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=10000002&dest_region_id=10000043", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: TriggerScan ---

func Test_HaulingRuns_TriggerScan_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(nil)

	body, _ := json.Marshal(map[string]int64{"regionId": 10000002, "systemId": 0})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, "done", resp["status"])
	assert.Equal(t, true, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

func Test_HaulingRuns_TriggerScan_WithDestRegion(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(nil)
	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000043), int64(0)).Return(nil)

	body, _ := json.Marshal(map[string]int64{"regionId": 10000002, "systemId": 0, "destRegionId": 10000043})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, "done", resp["status"])
	assert.Equal(t, true, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

func Test_HaulingRuns_TriggerScan_SourceScanError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(errors.New("ESI timeout"))

	body, _ := json.Marshal(map[string]int64{"regionId": 10000002, "destRegionId": 10000043})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_HaulingRuns_TriggerScan_DestScanError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(nil)
	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000043), int64(0)).Return(errors.New("ESI timeout"))

	body, _ := json.Marshal(map[string]int64{"regionId": 10000002, "destRegionId": 10000043})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_HaulingRuns_TriggerScan_MissingRegionID(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]int64{"regionId": 0})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_TriggerScan_InvalidBody(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader([]byte("bad json")))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Tests: GetScannerResults — structure variants ---

func Test_HaulingRuns_GetScannerResults_SourceStructure(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	buyPrice := 900.0
	sellPrice := 1100.0
	spread := 0.22
	net := 200.0
	expectedRows := []*models.HaulingArbitrageRow{
		{TypeID: int64(35), TypeName: "Pyerite", BuyPrice: &buyPrice, SellPrice: &sellPrice, Spread: &spread, NetProfitISK: &net, Indicator: "gap"},
	}
	mocks.structures.On("GetStructureScannerResults", mock.Anything, int64(1234567890), int64(10000043), int64(0)).Return(expectedRows, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_structure_id=1234567890&dest_region_id=10000043", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rows := result.([]*models.HaulingArbitrageRow)
	assert.Len(t, rows, 1)
	mocks.structures.AssertExpectations(t)
}

func Test_HaulingRuns_GetScannerResults_DestStructure(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expectedRows := []*models.HaulingArbitrageRow{}
	mocks.structures.On("GetRegionToStructureResults", mock.Anything, int64(10000002), int64(0), int64(9876543210)).Return(expectedRows, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_region_id=10000002&dest_structure_id=9876543210", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.structures.AssertExpectations(t)
}

func Test_HaulingRuns_GetScannerResults_SourceAndDestStructure_Error(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_structure_id=1234567890&dest_structure_id=9876543210", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetScannerResults_SourceStructureMissingDestRegion(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_structure_id=1234567890", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetScannerResults_DestStructureMissingSourceRegion(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?dest_structure_id=9876543210", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetScannerResults_StructureRepoError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.structures.On("GetStructureScannerResults", mock.Anything, int64(1234567890), int64(10000043), int64(0)).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/scanner?source_structure_id=1234567890&dest_region_id=10000043", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetScannerResults(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: TriggerScan — structure variants ---

func Test_HaulingRuns_TriggerScan_SourceStructure(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanStructure", mock.Anything, int64(1234567890), "mytoken").Return(true, nil)

	body, _ := json.Marshal(map[string]interface{}{"sourceStructureId": int64(1234567890), "charToken": "mytoken"})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, "done", resp["status"])
	assert.Equal(t, true, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

func Test_HaulingRuns_TriggerScan_SourceStructureAccessDenied(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanStructure", mock.Anything, int64(1234567890), "mytoken").Return(false, nil)

	body, _ := json.Marshal(map[string]interface{}{"sourceStructureId": int64(1234567890), "charToken": "mytoken"})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, "done", resp["status"])
	assert.Equal(t, false, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

func Test_HaulingRuns_TriggerScan_SourceStructureMissingToken(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	body, _ := json.Marshal(map[string]interface{}{"sourceStructureId": int64(1234567890)})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_TriggerScan_StructureScanError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanStructure", mock.Anything, int64(1234567890), "mytoken").Return(false, errors.New("ESI error"))

	body, _ := json.Marshal(map[string]interface{}{"sourceStructureId": int64(1234567890), "charToken": "mytoken"})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_HaulingRuns_TriggerScan_DestStructure(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(nil)
	mocks.scanner.On("ScanStructure", mock.Anything, int64(9876543210), "mytoken").Return(true, nil)

	body, _ := json.Marshal(map[string]interface{}{"regionId": int64(10000002), "destStructureId": int64(9876543210), "charToken": "mytoken"})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, "done", resp["status"])
	assert.Equal(t, true, resp["accessOk"])
	mocks.scanner.AssertExpectations(t)
}

// --- Tests: GetPnl ---

func Test_HaulingRuns_GetPnl_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "SELLING", Items: []*models.HaulingRunItem{}}
	netProfit := 50000.0
	entries := []*models.HaulingRunPnlEntry{
		{ID: int64(1), RunID: int64(5), TypeID: int64(34), QuantitySold: int64(100), NetProfitISK: &netProfit},
	}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.pnl.On("GetPnlByRunID", mock.Anything, int64(5)).Return(entries, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5/pnl", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetPnl(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	e := result.([]*models.HaulingRunPnlEntry)
	assert.Len(t, e, 1)
	mocks.runs.AssertExpectations(t)
	mocks.pnl.AssertExpectations(t)
}

func Test_HaulingRuns_GetPnl_RunNotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5/pnl", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetPnl(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_HaulingRuns_GetPnl_InvalidID(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/abc/pnl", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "abc"}}

	result, httpErr := controller.GetPnl(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetPnl_RepoError(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "SELLING", Items: []*models.HaulingRunItem{}}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.pnl.On("GetPnlByRunID", mock.Anything, int64(5)).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5/pnl", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetPnl(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: UpsertPnlEntry ---

func Test_HaulingRuns_UpsertPnlEntry_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "SELLING", Items: []*models.HaulingRunItem{}}
	totalRevenue := 150000.0
	totalCost := 100000.0

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.pnl.On("UpsertPnlEntry", mock.Anything, mock.AnythingOfType("*models.HaulingRunPnlEntry")).Return(nil)

	entryBody := models.HaulingRunPnlEntry{
		TypeID:          int64(34),
		QuantitySold:    int64(100),
		TotalRevenueISK: &totalRevenue,
		TotalCostISK:    &totalCost,
	}
	body, _ := json.Marshal(entryBody)
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/pnl", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpsertPnlEntry(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	mocks.runs.AssertExpectations(t)
	mocks.pnl.AssertExpectations(t)
}

func Test_HaulingRuns_UpsertPnlEntry_RunNotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	body, _ := json.Marshal(models.HaulingRunPnlEntry{TypeID: int64(34), QuantitySold: int64(10)})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/pnl", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpsertPnlEntry(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_HaulingRuns_UpsertPnlEntry_InvalidBody(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "SELLING"}
	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)

	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/pnl", bytes.NewReader([]byte("not json")))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpsertPnlEntry(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- Tests: GetPnlSummary ---

func Test_HaulingRuns_GetPnlSummary_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	run := &models.HaulingRun{ID: int64(5), UserID: userID, Status: "SELLING", Items: []*models.HaulingRunItem{}}
	summary := &models.HaulingRunPnlSummary{
		TotalRevenueISK: 150000.0,
		TotalCostISK:    100000.0,
		NetProfitISK:    50000.0,
		MarginPct:       33.33,
		ItemsSold:       int64(2),
	}

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(run, nil)
	mocks.pnl.On("GetPnlSummaryByRunID", mock.Anything, int64(5)).Return(summary, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5/pnl/summary", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetPnlSummary(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	s := result.(*models.HaulingRunPnlSummary)
	assert.Equal(t, 50000.0, s.NetProfitISK)
	mocks.runs.AssertExpectations(t)
	mocks.pnl.AssertExpectations(t)
}

func Test_HaulingRuns_GetPnlSummary_RunNotFound(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/runs/5/pnl/summary", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.GetPnlSummary(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// --- Tests: UpdateStatus with notifier ---

func Test_HaulingRuns_UpdateStatus_CompleteTriggersNotifier(t *testing.T) {
	controller, mocks := setupHaulingControllerWithNotifier()
	userID := int64(100)

	mocks.runs.On("UpdateRunStatus", mock.Anything, int64(5), userID, "COMPLETE").Return(nil)
	// Async goroutine calls these — use Maybe() since timing is uncertain
	mocks.runs.On("GetRunByID", mock.Anything, int64(5), userID).Return(
		&models.HaulingRun{ID: int64(5), UserID: userID, Status: "COMPLETE"},
		nil,
	).Maybe()
	mocks.pnl.On("GetPnlSummaryByRunID", mock.Anything, int64(5)).Return(
		&models.HaulingRunPnlSummary{}, nil,
	).Maybe()
	mocks.notifier.On("NotifyHaulingComplete", mock.Anything, userID,
		mock.AnythingOfType("*models.HaulingRun"),
		mock.AnythingOfType("*models.HaulingRunPnlSummary"),
	).Maybe()

	body, _ := json.Marshal(map[string]string{"status": "COMPLETE"})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/status", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateStatus(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	// The main mock must be satisfied (sync call)
	mocks.runs.AssertCalled(t, "UpdateRunStatus", mock.Anything, int64(5), userID, "COMPLETE")
}

func Test_HaulingRuns_UpdateStatus_NonCompleteNoNotifier(t *testing.T) {
	controller, mocks := setupHaulingControllerWithNotifier()
	userID := int64(100)

	mocks.runs.On("UpdateRunStatus", mock.Anything, int64(5), userID, "SELLING").Return(nil)

	body, _ := json.Marshal(map[string]string{"status": "SELLING"})
	req := httptest.NewRequest("PUT", "/v1/hauling/runs/5/status", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.UpdateStatus(args)
	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	// Notifier should not be called for non-COMPLETE status
	mocks.notifier.AssertNotCalled(t, "NotifyHaulingComplete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mocks.runs.AssertExpectations(t)
}

// --- Tests: GetRouteAnalytics ---

func Test_HaulingRuns_GetRouteAnalytics_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expected := []*models.HaulingRouteAnalytics{
		{FromRegionID: int64(10000002), ToRegionID: int64(10000043), TotalRuns: int64(3), TotalProfitISK: 500000.0},
	}
	mocks.analytics.On("GetRouteAnalytics", mock.Anything, userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/routes", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetRouteAnalytics(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rows := result.([]*models.HaulingRouteAnalytics)
	assert.Len(t, rows, 1)
	assert.Equal(t, 500000.0, rows[0].TotalProfitISK)
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetRouteAnalytics_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetRouteAnalytics", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/routes", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetRouteAnalytics(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: GetItemAnalytics ---

func Test_HaulingRuns_GetItemAnalytics_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expected := []*models.HaulingItemAnalytics{
		{TypeID: int64(34), TypeName: "Tritanium", TotalRuns: int64(5), TotalQtySold: int64(1000), TotalProfitISK: 100000.0},
	}
	mocks.analytics.On("GetItemAnalytics", mock.Anything, userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/items", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetItemAnalytics(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	items := result.([]*models.HaulingItemAnalytics)
	assert.Len(t, items, 1)
	assert.Equal(t, "Tritanium", items[0].TypeName)
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetItemAnalytics_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetItemAnalytics", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/items", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetItemAnalytics(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: GetProfitTimeSeries ---

func Test_HaulingRuns_GetProfitTimeSeries_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expected := []*models.HaulingProfitDataPoint{
		{Date: "2026-03-01", FromRegionID: int64(10000002), ToRegionID: int64(10000043), ProfitISK: 250000.0, RunCount: int64(2)},
	}
	mocks.analytics.On("GetProfitTimeSeries", mock.Anything, userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/timeseries", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetProfitTimeSeries(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	points := result.([]*models.HaulingProfitDataPoint)
	assert.Len(t, points, 1)
	assert.Equal(t, "2026-03-01", points[0].Date)
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetProfitTimeSeries_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetProfitTimeSeries", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/timeseries", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetProfitTimeSeries(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: GetRunDurationSummary ---

func Test_HaulingRuns_GetRunDurationSummary_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	expected := &models.HaulingRunDurationSummary{
		TotalCompletedRuns: int64(10),
		AvgDurationDays:    3.5,
		MinDurationDays:    1.0,
		MaxDurationDays:    7.0,
		TotalProfitISK:     1500000.0,
	}
	mocks.analytics.On("GetRunDurationSummary", mock.Anything, userID).Return(expected, nil)

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/summary", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetRunDurationSummary(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	s := result.(*models.HaulingRunDurationSummary)
	assert.Equal(t, int64(10), s.TotalCompletedRuns)
	assert.Equal(t, 3.5, s.AvgDurationDays)
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetRunDurationSummary_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetRunDurationSummary", mock.Anything, userID).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/analytics/summary", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetRunDurationSummary(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Tests: GetHaulingHistory ---

func Test_HaulingRuns_GetHaulingHistory_Success(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	completedAt := "2026-03-01T12:00:00Z"
	runs := []*models.HaulingRun{
		{ID: int64(1), Status: "COMPLETE", CompletedAt: &completedAt, Items: []*models.HaulingRunItem{}},
		{ID: int64(2), Status: "CANCELLED", Items: []*models.HaulingRunItem{}},
	}
	mocks.analytics.On("GetCompletedRuns", mock.Anything, userID, 20, 0).Return(runs, int64(2), nil)

	req := httptest.NewRequest("GET", "/v1/hauling/history", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetHaulingHistory(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, int64(2), resp["total"])
	assert.Equal(t, 20, resp["limit"])
	assert.Equal(t, 0, resp["offset"])
	returnedRuns := resp["runs"].([]*models.HaulingRun)
	assert.Len(t, returnedRuns, 2)
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetHaulingHistory_WithPagination(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetCompletedRuns", mock.Anything, userID, 5, 10).Return([]*models.HaulingRun{}, int64(25), nil)

	req := httptest.NewRequest("GET", "/v1/hauling/history?limit=5&offset=10", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetHaulingHistory(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]interface{})
	assert.Equal(t, int64(25), resp["total"])
	assert.Equal(t, 5, resp["limit"])
	assert.Equal(t, 10, resp["offset"])
	mocks.analytics.AssertExpectations(t)
}

func Test_HaulingRuns_GetHaulingHistory_InvalidLimit(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/history?limit=abc", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetHaulingHistory(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetHaulingHistory_InvalidOffset(t *testing.T) {
	controller, _ := setupHaulingController()
	userID := int64(100)

	req := httptest.NewRequest("GET", "/v1/hauling/history?offset=xyz", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetHaulingHistory(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_HaulingRuns_GetHaulingHistory_Error(t *testing.T) {
	controller, mocks := setupHaulingController()
	userID := int64(100)

	mocks.analytics.On("GetCompletedRuns", mock.Anything, userID, 20, 0).Return(nil, int64(0), errors.New("db error"))

	req := httptest.NewRequest("GET", "/v1/hauling/history", nil)
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.GetHaulingHistory(args)
	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}
