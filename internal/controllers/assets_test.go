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

// Mock repository
type MockAssetsRepository struct {
	mock.Mock
}

func (m *MockAssetsRepository) GetUserAssets(ctx context.Context, user int64) (*repositories.AssetsResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.AssetsResponse), args.Error(1)
}

func (m *MockAssetsRepository) GetUserAssetsSummary(ctx context.Context, user int64) (*repositories.AssetsSummary, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.AssetsSummary), args.Error(1)
}

func Test_AssetsController_GetUserAssets_Success(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	expectedResponse := &repositories.AssetsResponse{
		Structures: []*repositories.AssetStructure{
			{
				ID:   60003760,
				Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
				HangarAssets: []*repositories.Asset{
					{
						Name:     "Tritanium",
						TypeID:   34,
						Quantity: 1000,
						Volume:   10.0,
					},
				},
			},
		},
	}

	mockRepo.On("GetUserAssets", mock.Anything, userID).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/v1/assets/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssets(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*repositories.AssetsResponse)
	assert.Len(t, response.Structures, 1)
	assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", response.Structures[0].Name)
	assert.Len(t, response.Structures[0].HangarAssets, 1)

	mockRepo.AssertExpectations(t)
}

func Test_AssetsController_GetUserAssets_RepositoryError(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	mockRepo.On("GetUserAssets", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/assets/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssets(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AssetsController_GetUserAssets_EmptyResponse(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	emptyResponse := &repositories.AssetsResponse{
		Structures: []*repositories.AssetStructure{},
	}

	mockRepo.On("GetUserAssets", mock.Anything, userID).Return(emptyResponse, nil)

	req := httptest.NewRequest("GET", "/v1/assets/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssets(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*repositories.AssetsResponse)
	assert.Len(t, response.Structures, 0)

	mockRepo.AssertExpectations(t)
}

func Test_AssetsController_GetUserAssetsSummary_Success(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	expectedSummary := &repositories.AssetsSummary{
		TotalValue:   1234567.89,
		TotalDeficit: 9876.54,
	}

	mockRepo.On("GetUserAssetsSummary", mock.Anything, userID).Return(expectedSummary, nil)

	req := httptest.NewRequest("GET", "/v1/assets/summary", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssetsSummary(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	summary := result.(*repositories.AssetsSummary)
	assert.Equal(t, 1234567.89, summary.TotalValue)
	assert.Equal(t, 9876.54, summary.TotalDeficit)

	mockRepo.AssertExpectations(t)
}

func Test_AssetsController_GetUserAssetsSummary_RepositoryError(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	mockRepo.On("GetUserAssetsSummary", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/assets/summary", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssetsSummary(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_AssetsController_GetUserAssetsSummary_ZeroValues(t *testing.T) {
	mockRepo := new(MockAssetsRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewAssets(mockRouter, mockRepo)

	userID := int64(42)
	zeroSummary := &repositories.AssetsSummary{
		TotalValue:   0.0,
		TotalDeficit: 0.0,
	}

	mockRepo.On("GetUserAssetsSummary", mock.Anything, userID).Return(zeroSummary, nil)

	req := httptest.NewRequest("GET", "/v1/assets/summary", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetUserAssetsSummary(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	summary := result.(*repositories.AssetsSummary)
	assert.Equal(t, 0.0, summary.TotalValue)
	assert.Equal(t, 0.0, summary.TotalDeficit)

	mockRepo.AssertExpectations(t)
}
