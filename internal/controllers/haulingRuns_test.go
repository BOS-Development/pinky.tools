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

// --- Setup helper ---

type haulingMocks struct {
	runs    *MockHaulingRunsRepository
	items   *MockHaulingRunItemsRepository
	market  *MockHaulingMarketRepo
	scanner *MockHaulingMarketUpdater
}

func setupHaulingController() (*controllers.HaulingRunsController, haulingMocks) {
	mocks := haulingMocks{
		runs:    new(MockHaulingRunsRepository),
		items:   new(MockHaulingRunItemsRepository),
		market:  new(MockHaulingMarketRepo),
		scanner: new(MockHaulingMarketUpdater),
	}
	router := &MockRouter{}
	controller := controllers.NewHaulingRuns(router, mocks.runs, mocks.items, mocks.market, mocks.scanner)
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

	// The scan runs in a background goroutine; set up with Maybe so it may or may not be called
	mocks.scanner.On("ScanRegion", mock.Anything, int64(10000002), int64(0)).Return(nil).Maybe()

	body, _ := json.Marshal(map[string]int64{"regionId": 10000002, "systemId": 0})
	req := httptest.NewRequest("POST", "/v1/hauling/scanner/scan", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID, Params: map[string]string{}}

	result, httpErr := controller.TriggerScan(args)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	resp := result.(map[string]string)
	assert.Equal(t, "scanning", resp["status"])
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
