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

func int64Ptr(v int64) *int64 { return &v }

// Mock AutoSellContainersRepository
type MockAutoSellContainersRepository struct {
	mock.Mock
}

func (m *MockAutoSellContainersRepository) GetByUser(ctx context.Context, userID int64) ([]*models.AutoSellContainer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AutoSellContainer), args.Error(1)
}

func (m *MockAutoSellContainersRepository) Upsert(ctx context.Context, container *models.AutoSellContainer) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockAutoSellContainersRepository) Delete(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockAutoSellContainersRepository) GetByID(ctx context.Context, id int64) (*models.AutoSellContainer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AutoSellContainer), args.Error(1)
}

// Mock AutoSellSyncer
type MockAutoSellSyncer struct {
	mock.Mock
}

func (m *MockAutoSellSyncer) SyncForUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Mock ForSaleItemsDeactivator
type MockForSaleItemsDeactivator struct {
	mock.Mock
}

func (m *MockForSaleItemsDeactivator) DeactivateAutoSellListings(ctx context.Context, autoSellContainerID int64) error {
	args := m.Called(ctx, autoSellContainerID)
	return args.Error(0)
}

// --- GetMyConfigs Tests ---

func Test_AutoSellContainersController_GetMyConfigs_Success(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	expectedItems := []*models.AutoSellContainer{
		{
			ID:              1,
			UserID:          userID,
			OwnerType:       "character",
			OwnerID:         456,
			LocationID:      60003760,
			ContainerID:     int64Ptr(9000),
			PricePercentage: 90.0,
			IsActive:        true,
		},
	}

	mockRepo.On("GetByUser", mock.Anything, userID).Return(expectedItems, nil)

	req := httptest.NewRequest("GET", "/v1/auto-sell", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	items := result.([]*models.AutoSellContainer)
	assert.Len(t, items, 1)
	assert.Equal(t, 90.0, items[0].PricePercentage)
	assert.Equal(t, int64Ptr(9000), items[0].ContainerID)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_GetMyConfigs_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("GET", "/v1/auto-sell", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoSellContainersController_GetMyConfigs_RepositoryError(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/auto-sell", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

// --- CreateConfig Tests ---

func Test_AutoSellContainersController_CreateConfig_Success(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoSellContainer) bool {
		return c.UserID == userID &&
			c.OwnerType == "character" &&
			c.OwnerID == 456 &&
			c.LocationID == 60003760 &&
			c.ContainerID != nil && *c.ContainerID == 9000 &&
			c.PricePercentage == 90.0 &&
			c.PriceSource == "jita_buy"
	})).Return(nil)

	// The sync is triggered asynchronously, so we use mock.Anything for context
	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	// No priceSource sent — should default to "jita_buy"
	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 90.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_CreateConfig_InvalidJSON(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoSellContainersController_CreateConfig_MissingOwnerType(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 90.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "ownerType is required")
}

func Test_AutoSellContainersController_CreateConfig_InvalidPercentage(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 250.0, // Over 200
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "pricePercentage must be between 0 and 200")
}

func Test_AutoSellContainersController_CreateConfig_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("POST", "/v1/auto-sell", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoSellContainersController_CreateConfig_WithJitaSell(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoSellContainer) bool {
		return c.PriceSource == "jita_sell" && c.PricePercentage == 95.0
	})).Return(nil)

	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 95.0,
		"priceSource":     "jita_sell",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_CreateConfig_InvalidPriceSource(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 90.0,
		"priceSource":     "invalid_source",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-sell", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "invalid priceSource")
}

// --- UpdateConfig Tests ---

func Test_AutoSellContainersController_UpdateConfig_Success(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	existingConfig := &models.AutoSellContainer{
		ID:              configID,
		UserID:          userID,
		OwnerType:       "character",
		OwnerID:         456,
		LocationID:      60003760,
		ContainerID:     int64Ptr(9000),
		PricePercentage: 90.0,
		PriceSource:     "jita_buy",
		IsActive:        true,
	}

	mockRepo.On("GetByID", mock.Anything, configID).Return(existingConfig, nil)
	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoSellContainer) bool {
		return c.ID == configID && c.PricePercentage == 85.0 && c.PriceSource == "jita_sell"
	})).Return(nil)
	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	body := map[string]interface{}{
		"pricePercentage": 85.0,
		"priceSource":     "jita_sell",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-sell/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_UpdateConfig_NotFound(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil)

	body := map[string]interface{}{
		"pricePercentage": 85.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-sell/999", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_UpdateConfig_NotOwner(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	otherUserID := int64(999)

	existingConfig := &models.AutoSellContainer{
		ID:              1,
		UserID:          otherUserID, // Different owner
		OwnerType:       "character",
		OwnerID:         456,
		PricePercentage: 90.0,
	}

	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(existingConfig, nil)

	body := map[string]interface{}{
		"pricePercentage": 85.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-sell/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_UpdateConfig_InvalidID(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("PUT", "/v1/auto-sell/invalid", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "invalid"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoSellContainersController_UpdateConfig_InvalidPercentage(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	existingConfig := &models.AutoSellContainer{
		ID:              1,
		UserID:          userID,
		OwnerType:       "character",
		OwnerID:         456,
		PricePercentage: 90.0,
	}

	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(existingConfig, nil)

	body := map[string]interface{}{
		"pricePercentage": 0, // Invalid: must be > 0
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-sell/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "pricePercentage must be between 0 and 200")

	mockRepo.AssertExpectations(t)
}

// --- DeleteConfig Tests ---

func Test_AutoSellContainersController_DeleteConfig_Success(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	mockDeactivator.On("DeactivateAutoSellListings", mock.Anything, configID).Return(nil)
	mockRepo.On("Delete", mock.Anything, configID, userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/auto-sell/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockDeactivator.AssertExpectations(t)
}

func Test_AutoSellContainersController_DeleteConfig_NotFound(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(999)

	mockDeactivator.On("DeactivateAutoSellListings", mock.Anything, configID).Return(nil)
	mockRepo.On("Delete", mock.Anything, configID, userID).Return(errors.New("auto-sell container not found or user is not the owner"))

	req := httptest.NewRequest("DELETE", "/v1/auto-sell/999", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoSellContainersController_DeleteConfig_InvalidID(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("DELETE", "/v1/auto-sell/invalid", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "invalid"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoSellContainersController_DeleteConfig_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("DELETE", "/v1/auto-sell/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoSellContainersController_DeleteConfig_DeactivateError(t *testing.T) {
	mockRepo := new(MockAutoSellContainersRepository)
	mockSyncer := new(MockAutoSellSyncer)
	mockDeactivator := new(MockForSaleItemsDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	mockDeactivator.On("DeactivateAutoSellListings", mock.Anything, configID).Return(errors.New("deactivation failed"))

	req := httptest.NewRequest("DELETE", "/v1/auto-sell/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoSellContainers(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockDeactivator.AssertExpectations(t)
}
