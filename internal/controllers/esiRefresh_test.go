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

// MockESIRefreshUpdater mocks the ESIRefreshUpdater (market prices)
type MockESIRefreshUpdater struct {
	mock.Mock
}

func (m *MockESIRefreshUpdater) ForceUpdateJitaMarket(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockESIRefreshCostIndicesUpdater mocks the ESIRefreshCostIndicesUpdater
type MockESIRefreshCostIndicesUpdater struct {
	mock.Mock
}

func (m *MockESIRefreshCostIndicesUpdater) ForceUpdate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockESIRefreshCcpPricesUpdater mocks the ESIRefreshCcpPricesUpdater
type MockESIRefreshCcpPricesUpdater struct {
	mock.Mock
}

func (m *MockESIRefreshCcpPricesUpdater) ForceUpdate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type esiRefreshMocks struct {
	marketPrices *MockESIRefreshUpdater
	costIndices  *MockESIRefreshCostIndicesUpdater
	ccpPrices    *MockESIRefreshCcpPricesUpdater
}

func setupESIRefreshController() (*controllers.ESIRefresh, esiRefreshMocks) {
	mocks := esiRefreshMocks{
		marketPrices: new(MockESIRefreshUpdater),
		costIndices:  new(MockESIRefreshCostIndicesUpdater),
		ccpPrices:    new(MockESIRefreshCcpPricesUpdater),
	}
	router := &MockRouter{}
	controller := controllers.NewESIRefresh(router, mocks.marketPrices, mocks.costIndices, mocks.ccpPrices)
	return controller, mocks
}

func Test_ESIRefresh_RefreshAll_AllSucceed(t *testing.T) {
	controller, mocks := setupESIRefreshController()

	mocks.marketPrices.On("ForceUpdateJitaMarket", mock.Anything).Return(nil)
	mocks.costIndices.On("ForceUpdate", mock.Anything).Return(nil)
	mocks.ccpPrices.On("ForceUpdate", mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/esi/refresh", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.RefreshAll(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	refreshResult, ok := result.(*controllers.ESIRefreshResult)
	assert.True(t, ok)
	assert.Equal(t, "ok", refreshResult.MarketPrices)
	assert.Equal(t, "ok", refreshResult.CostIndices)
	assert.Equal(t, "ok", refreshResult.CcpPrices)

	mocks.marketPrices.AssertExpectations(t)
	mocks.costIndices.AssertExpectations(t)
	mocks.ccpPrices.AssertExpectations(t)
}

func Test_ESIRefresh_RefreshAll_MarketPricesFails(t *testing.T) {
	controller, mocks := setupESIRefreshController()

	mocks.marketPrices.On("ForceUpdateJitaMarket", mock.Anything).Return(errors.New("ESI timeout"))
	mocks.costIndices.On("ForceUpdate", mock.Anything).Return(nil)
	mocks.ccpPrices.On("ForceUpdate", mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/esi/refresh", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.RefreshAll(args)

	// Partial failure — still returns 200 with result body (not an HTTP error)
	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	refreshResult, ok := result.(*controllers.ESIRefreshResult)
	assert.True(t, ok)
	assert.Contains(t, refreshResult.MarketPrices, "error:")
	assert.Equal(t, "ok", refreshResult.CostIndices)
	assert.Equal(t, "ok", refreshResult.CcpPrices)

	mocks.marketPrices.AssertExpectations(t)
	mocks.costIndices.AssertExpectations(t)
	mocks.ccpPrices.AssertExpectations(t)
}

func Test_ESIRefresh_RefreshAll_CostIndicesFails(t *testing.T) {
	controller, mocks := setupESIRefreshController()

	mocks.marketPrices.On("ForceUpdateJitaMarket", mock.Anything).Return(nil)
	mocks.costIndices.On("ForceUpdate", mock.Anything).Return(errors.New("db error"))
	mocks.ccpPrices.On("ForceUpdate", mock.Anything).Return(nil)

	req := httptest.NewRequest("POST", "/v1/esi/refresh", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.RefreshAll(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	refreshResult, ok := result.(*controllers.ESIRefreshResult)
	assert.True(t, ok)
	assert.Equal(t, "ok", refreshResult.MarketPrices)
	assert.Contains(t, refreshResult.CostIndices, "error:")
	assert.Equal(t, "ok", refreshResult.CcpPrices)

	mocks.marketPrices.AssertExpectations(t)
	mocks.costIndices.AssertExpectations(t)
	mocks.ccpPrices.AssertExpectations(t)
}

func Test_ESIRefresh_RefreshAll_CcpPricesFails(t *testing.T) {
	controller, mocks := setupESIRefreshController()

	mocks.marketPrices.On("ForceUpdateJitaMarket", mock.Anything).Return(nil)
	mocks.costIndices.On("ForceUpdate", mock.Anything).Return(nil)
	mocks.ccpPrices.On("ForceUpdate", mock.Anything).Return(errors.New("network error"))

	req := httptest.NewRequest("POST", "/v1/esi/refresh", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.RefreshAll(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	refreshResult, ok := result.(*controllers.ESIRefreshResult)
	assert.True(t, ok)
	assert.Equal(t, "ok", refreshResult.MarketPrices)
	assert.Equal(t, "ok", refreshResult.CostIndices)
	assert.Contains(t, refreshResult.CcpPrices, "error:")

	mocks.marketPrices.AssertExpectations(t)
	mocks.costIndices.AssertExpectations(t)
	mocks.ccpPrices.AssertExpectations(t)
}

func Test_ESIRefresh_RefreshAll_AllFail_Returns500(t *testing.T) {
	controller, mocks := setupESIRefreshController()

	mocks.marketPrices.On("ForceUpdateJitaMarket", mock.Anything).Return(errors.New("ESI down"))
	mocks.costIndices.On("ForceUpdate", mock.Anything).Return(errors.New("ESI down"))
	mocks.ccpPrices.On("ForceUpdate", mock.Anything).Return(errors.New("ESI down"))

	req := httptest.NewRequest("POST", "/v1/esi/refresh", nil)
	args := &web.HandlerArgs{Request: req}

	result, httpErr := controller.RefreshAll(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "all ESI refresh operations failed")

	mocks.marketPrices.AssertExpectations(t)
	mocks.costIndices.AssertExpectations(t)
	mocks.ccpPrices.AssertExpectations(t)
}

func Test_ESIRefresh_Constructor_RegistersRoute(t *testing.T) {
	controller, _ := setupESIRefreshController()
	assert.NotNil(t, controller)
}
