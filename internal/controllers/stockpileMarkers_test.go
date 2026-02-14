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

// Mock repository
type MockStockpileMarkersRepository struct {
	mock.Mock
}

func (m *MockStockpileMarkersRepository) GetByUser(ctx context.Context, userID int64) ([]*models.StockpileMarker, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StockpileMarker), args.Error(1)
}

func (m *MockStockpileMarkersRepository) Upsert(ctx context.Context, marker *models.StockpileMarker) error {
	args := m.Called(ctx, marker)
	return args.Error(0)
}

func (m *MockStockpileMarkersRepository) Delete(ctx context.Context, marker *models.StockpileMarker) error {
	args := m.Called(ctx, marker)
	return args.Error(0)
}

// Mock router
type MockRouter struct{}

func (m *MockRouter) RegisterRestAPIRoute(path string, auth web.AuthAccess, handler func(*web.HandlerArgs) (any, *web.HttpError), methods ...string) {}

func Test_StockpileMarkersController_GetStockpiles_Success(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	expectedMarkers := []*models.StockpileMarker{
		{
			UserID:          42,
			TypeID:          34,
			OwnerType:       "character",
			OwnerID:         1337,
			LocationID:      60003760,
			DesiredQuantity: 1000,
		},
		{
			UserID:          42,
			TypeID:          35,
			OwnerType:       "corporation",
			OwnerID:         2001,
			LocationID:      60003760,
			DesiredQuantity: 500,
		},
	}

	mockRepo.On("GetByUser", mock.Anything, userID).Return(expectedMarkers, nil)

	req := httptest.NewRequest("GET", "/v1/stockpiles", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetStockpiles(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	markers := result.([]*models.StockpileMarker)
	assert.Len(t, markers, 2)
	assert.Equal(t, int64(34), markers[0].TypeID)
	assert.Equal(t, int64(35), markers[1].TypeID)

	mockRepo.AssertExpectations(t)
}

func Test_StockpileMarkersController_GetStockpiles_RepositoryError(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	mockRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/stockpiles", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetStockpiles(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_StockpileMarkersController_UpsertStockpile_Success(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	marker := models.StockpileMarker{
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		DesiredQuantity: 1000,
	}

	// Expect the marker with UserID set
	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(m *models.StockpileMarker) bool {
		return m.UserID == userID && m.TypeID == 34
	})).Return(nil)

	body, _ := json.Marshal(marker)
	req := httptest.NewRequest("POST", "/v1/stockpiles", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.UpsertStockpile(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_StockpileMarkersController_UpsertStockpile_InvalidJSON(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)

	req := httptest.NewRequest("POST", "/v1/stockpiles", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.UpsertStockpile(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_StockpileMarkersController_UpsertStockpile_RepositoryError(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	marker := models.StockpileMarker{
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		DesiredQuantity: 1000,
	}

	mockRepo.On("Upsert", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(marker)
	req := httptest.NewRequest("POST", "/v1/stockpiles", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.UpsertStockpile(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_StockpileMarkersController_DeleteStockpile_Success(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	marker := models.StockpileMarker{
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		DesiredQuantity: 1000,
	}

	// Expect the marker with UserID set
	mockRepo.On("Delete", mock.Anything, mock.MatchedBy(func(m *models.StockpileMarker) bool {
		return m.UserID == userID && m.TypeID == 34
	})).Return(nil)

	body, _ := json.Marshal(marker)
	req := httptest.NewRequest("DELETE", "/v1/stockpiles", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.DeleteStockpile(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_StockpileMarkersController_DeleteStockpile_InvalidJSON(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)

	req := httptest.NewRequest("DELETE", "/v1/stockpiles", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.DeleteStockpile(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_StockpileMarkersController_DeleteStockpile_RepositoryError(t *testing.T) {
	mockRepo := new(MockStockpileMarkersRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpileMarkers(mockRouter, mockRepo)

	userID := int64(42)
	marker := models.StockpileMarker{
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		DesiredQuantity: 1000,
	}

	mockRepo.On("Delete", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(marker)
	req := httptest.NewRequest("DELETE", "/v1/stockpiles", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.DeleteStockpile(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}
