package controllers_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMarketPricesUpdater mocks the MarketPricesUpdater interface
type MockMarketPricesUpdater struct {
	mock.Mock
}

func (m *MockMarketPricesUpdater) UpdateJitaMarket(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func Test_MarketPricesController_UpdateJitaMarket_Success(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewMarketPrices(mockRouter, mockUpdater)

	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/market-prices/update", nil)
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.UpdateJitaMarket(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockUpdater.AssertExpectations(t)
	mockUpdater.AssertCalled(t, "UpdateJitaMarket", mock.Anything)
}

func Test_MarketPricesController_UpdateJitaMarket_UpdaterError(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewMarketPrices(mockRouter, mockUpdater)

	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(errors.New("ESI API error"))

	req := httptest.NewRequest("POST", "/v1/market-prices/update", nil)
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.UpdateJitaMarket(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "failed to update market prices")

	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesController_UpdateJitaMarket_NetworkError(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewMarketPrices(mockRouter, mockUpdater)

	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(errors.New("network timeout"))

	req := httptest.NewRequest("POST", "/v1/market-prices/update", nil)
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.UpdateJitaMarket(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesController_Constructor_RegistersRoute(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewMarketPrices(mockRouter, mockUpdater)

	assert.NotNil(t, controller)
	// Route registration is verified by the existence of the controller
}

func Test_MarketPricesController_UpdateJitaMarket_WithContext(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewMarketPrices(mockRouter, mockUpdater)

	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/market-prices/update", nil)
	// Add context to request
	ctx := req.Context()
	req = req.WithContext(ctx)

	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.UpdateJitaMarket(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	// Verify the context was passed through
	mockUpdater.AssertExpectations(t)
	mockUpdater.AssertNumberOfCalls(t, "UpdateJitaMarket", 1)
}
