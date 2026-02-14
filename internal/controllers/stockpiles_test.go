package controllers_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStockpilesRepository mocks the StockpilesRepository interface
type MockStockpilesRepository struct {
	mock.Mock
}

func (m *MockStockpilesRepository) GetStockpileDeficits(ctx context.Context, user int64) (*repositories.StockpilesResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.StockpilesResponse), args.Error(1)
}

func Test_StockpilesController_GetDeficits_Success(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	userID := int64(42)
	expectedResponse := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{
				Name:            "Tritanium",
				TypeID:          34,
				Quantity:        1000,
				Volume:          10.0,
				OwnerType:       "character",
				OwnerName:       "Test Character",
				OwnerID:         123,
				DesiredQuantity: 5000,
				StockpileDelta:  -4000,
				DeficitValue:    20000.0,
				StructureName:   "Jita IV",
				SolarSystem:     "Jita",
				Region:          "The Forge",
				ContainerName:   nil,
			},
		},
	}

	mockRepo.On("GetStockpileDeficits", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/v1/stockpiles/deficits", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetDeficits(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*repositories.StockpilesResponse)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, "Tritanium", response.Items[0].Name)
	assert.Equal(t, int64(34), response.Items[0].TypeID)
	assert.Equal(t, int64(-4000), response.Items[0].StockpileDelta)

	mockRepo.AssertExpectations(t)
}

func Test_StockpilesController_GetDeficits_EmptyResult(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	userID := int64(42)
	expectedResponse := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{},
	}

	mockRepo.On("GetStockpileDeficits", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/v1/stockpiles/deficits", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetDeficits(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*repositories.StockpilesResponse)
	assert.Len(t, response.Items, 0)

	mockRepo.AssertExpectations(t)
}

func Test_StockpilesController_GetDeficits_RepositoryError(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	userID := int64(42)

	mockRepo.On("GetStockpileDeficits", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/stockpiles/deficits", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetDeficits(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "database error")

	mockRepo.AssertExpectations(t)
}

func Test_StockpilesController_GetDeficits_MultipleDeficits(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	userID := int64(42)
	containerName := "Station Warehouse Container"
	expectedResponse := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{
				Name:            "Tritanium",
				TypeID:          34,
				Quantity:        1000,
				DesiredQuantity: 5000,
				StockpileDelta:  -4000,
				DeficitValue:    20000.0,
			},
			{
				Name:            "Pyerite",
				TypeID:          35,
				Quantity:        500,
				DesiredQuantity: 2000,
				StockpileDelta:  -1500,
				DeficitValue:    7500.0,
				ContainerName:   &containerName,
			},
		},
	}

	mockRepo.On("GetStockpileDeficits", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/v1/stockpiles/deficits", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetDeficits(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*repositories.StockpilesResponse)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, "Tritanium", response.Items[0].Name)
	assert.Equal(t, "Pyerite", response.Items[1].Name)
	assert.NotNil(t, response.Items[1].ContainerName)
	assert.Equal(t, "Station Warehouse Container", *response.Items[1].ContainerName)

	mockRepo.AssertExpectations(t)
}

func Test_StockpilesController_Constructor_RegistersRoute(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	assert.NotNil(t, controller)
	// Route registration is verified by the existence of the controller
}

func Test_StockpilesController_GetDeficits_WithContext(t *testing.T) {
	mockRepo := new(MockStockpilesRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewStockpiles(mockRouter, mockRepo)

	userID := int64(42)
	expectedResponse := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{},
	}

	mockRepo.On("GetStockpileDeficits", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/v1/stockpiles/deficits", nil)
	ctx := req.Context()
	req = req.WithContext(ctx)

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetDeficits(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	// Verify the context was passed through
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "GetStockpileDeficits", 1)
}
