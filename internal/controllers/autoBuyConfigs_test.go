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

// Mock AutoBuyConfigsRepository
type MockAutoBuyConfigsRepository struct {
	mock.Mock
}

func (m *MockAutoBuyConfigsRepository) GetByUser(ctx context.Context, userID int64) ([]*models.AutoBuyConfig, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AutoBuyConfig), args.Error(1)
}

func (m *MockAutoBuyConfigsRepository) Upsert(ctx context.Context, config *models.AutoBuyConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAutoBuyConfigsRepository) Delete(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockAutoBuyConfigsRepository) GetByID(ctx context.Context, id int64) (*models.AutoBuyConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AutoBuyConfig), args.Error(1)
}

// Mock AutoBuyConfigsSyncer
type MockAutoBuyConfigsSyncer struct {
	mock.Mock
}

func (m *MockAutoBuyConfigsSyncer) SyncForUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Mock BuyOrdersDeactivator
type MockBuyOrdersDeactivator struct {
	mock.Mock
}

func (m *MockBuyOrdersDeactivator) DeactivateAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) error {
	args := m.Called(ctx, autoBuyConfigID)
	return args.Error(0)
}

// --- GetMyConfigs Tests ---

func Test_AutoBuyConfigsController_GetMyConfigs_Success(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	containerID := int64(9000)
	expectedItems := []*models.AutoBuyConfig{
		{
			ID:              1,
			UserID:          userID,
			OwnerType:       "character",
			OwnerID:         456,
			LocationID:      60003760,
			ContainerID:     &containerID,
			PricePercentage: 90.0,
			PriceSource:     "jita_sell",
			IsActive:        true,
		},
	}

	mockRepo.On("GetByUser", mock.Anything, userID).Return(expectedItems, nil)

	req := httptest.NewRequest("GET", "/v1/auto-buy", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	items := result.([]*models.AutoBuyConfig)
	assert.Len(t, items, 1)
	assert.Equal(t, 90.0, items[0].PricePercentage)
	assert.Equal(t, int64(9000), *items[0].ContainerID)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_GetMyConfigs_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("GET", "/v1/auto-buy", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_GetMyConfigs_RepositoryError(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/auto-buy", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.GetMyConfigs(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

// --- CreateConfig Tests ---

func Test_AutoBuyConfigsController_CreateConfig_Success_Defaults(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoBuyConfig) bool {
		return c.UserID == userID &&
			c.OwnerType == "character" &&
			c.OwnerID == 456 &&
			c.LocationID == 60003760 &&
			c.ContainerID == nil &&
			c.PricePercentage == 90.0 &&
			c.PriceSource == "jita_sell"
	})).Return(nil)

	// The sync is triggered asynchronously, so we use mock.Anything for context
	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	// No priceSource or containerId sent — should default to "jita_sell" and nil container
	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"pricePercentage": 90.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_CreateConfig_Success_WithContainerID(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoBuyConfig) bool {
		return c.UserID == userID &&
			c.OwnerType == "character" &&
			c.OwnerID == 456 &&
			c.LocationID == 60003760 &&
			c.ContainerID != nil && *c.ContainerID == 9000 &&
			c.PricePercentage == 90.0 &&
			c.PriceSource == "jita_sell"
	})).Return(nil)

	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"containerId":     9000,
		"pricePercentage": 90.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_CreateConfig_InvalidJSON(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_CreateConfig_MissingOwnerType(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerId":         456,
		"locationId":      60003760,
		"pricePercentage": 90.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "ownerType is required")
}

func Test_AutoBuyConfigsController_CreateConfig_InvalidPercentage(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"pricePercentage": 250.0, // Over 200
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "pricePercentage must be between 0 and 200")
}

func Test_AutoBuyConfigsController_CreateConfig_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("POST", "/v1/auto-buy", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_CreateConfig_WithJitaBuy(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoBuyConfig) bool {
		return c.PriceSource == "jita_buy" && c.PricePercentage == 95.0
	})).Return(nil)

	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"pricePercentage": 95.0,
		"priceSource":     "jita_buy",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_CreateConfig_InvalidPriceSource(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]interface{}{
		"ownerType":       "character",
		"ownerId":         456,
		"locationId":      60003760,
		"pricePercentage": 90.0,
		"priceSource":     "invalid_source",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auto-buy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.CreateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "invalid priceSource")
}

// --- UpdateConfig Tests ---

func Test_AutoBuyConfigsController_UpdateConfig_Success(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	existingConfig := &models.AutoBuyConfig{
		ID:              configID,
		UserID:          userID,
		OwnerType:       "character",
		OwnerID:         456,
		LocationID:      60003760,
		ContainerID:     nil,
		PricePercentage: 90.0,
		PriceSource:     "jita_sell",
		IsActive:        true,
	}

	mockRepo.On("GetByID", mock.Anything, configID).Return(existingConfig, nil)
	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.AutoBuyConfig) bool {
		return c.ID == configID && c.PricePercentage == 85.0 && c.PriceSource == "jita_buy"
	})).Return(nil)
	mockSyncer.On("SyncForUser", mock.Anything, userID).Return(nil)

	body := map[string]interface{}{
		"pricePercentage": 85.0,
		"priceSource":     "jita_buy",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-buy/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_UpdateConfig_NotFound(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil)

	body := map[string]interface{}{
		"pricePercentage": 85.0,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/auto-buy/999", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_UpdateConfig_NotOwner(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	otherUserID := int64(999)

	existingConfig := &models.AutoBuyConfig{
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

	req := httptest.NewRequest("PUT", "/v1/auto-buy/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_UpdateConfig_InvalidID(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("PUT", "/v1/auto-buy/invalid", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "invalid"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_UpdateConfig_InvalidPercentage(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	existingConfig := &models.AutoBuyConfig{
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

	req := httptest.NewRequest("PUT", "/v1/auto-buy/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.UpdateConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "pricePercentage must be between 0 and 200")

	mockRepo.AssertExpectations(t)
}

// --- DeleteConfig Tests ---

func Test_AutoBuyConfigsController_DeleteConfig_Success(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	mockDeactivator.On("DeactivateAutoBuyOrders", mock.Anything, configID).Return(nil)
	mockRepo.On("Delete", mock.Anything, configID, userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/auto-buy/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockDeactivator.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_DeleteConfig_NotFound(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(999)

	mockDeactivator.On("DeactivateAutoBuyOrders", mock.Anything, configID).Return(nil)
	mockRepo.On("Delete", mock.Anything, configID, userID).Return(errors.New("auto-buy config not found or user is not the owner"))

	req := httptest.NewRequest("DELETE", "/v1/auto-buy/999", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AutoBuyConfigsController_DeleteConfig_InvalidID(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)

	req := httptest.NewRequest("DELETE", "/v1/auto-buy/invalid", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "invalid"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_DeleteConfig_Unauthorized(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	req := httptest.NewRequest("DELETE", "/v1/auto-buy/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    nil,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func Test_AutoBuyConfigsController_DeleteConfig_DeactivationError(t *testing.T) {
	mockRepo := new(MockAutoBuyConfigsRepository)
	mockSyncer := new(MockAutoBuyConfigsSyncer)
	mockDeactivator := new(MockBuyOrdersDeactivator)
	mockRouter := &MockRouter{}

	userID := int64(123)
	configID := int64(1)

	mockDeactivator.On("DeactivateAutoBuyOrders", mock.Anything, configID).Return(errors.New("deactivation failed"))

	req := httptest.NewRequest("DELETE", "/v1/auto-buy/1", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	}

	controller := controllers.NewAutoBuyConfigs(mockRouter, mockRepo, mockSyncer, mockDeactivator)
	result, httpErr := controller.DeleteConfig(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockDeactivator.AssertExpectations(t)
}
