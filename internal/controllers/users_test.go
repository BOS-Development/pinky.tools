package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Get(ctx context.Context, id int64) (*repositories.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.User), args.Error(1)
}

func (m *MockUserRepository) Add(ctx context.Context, user *repositories.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockAssetsStatusRepository struct {
	mock.Mock
}

func (m *MockAssetsStatusRepository) GetAssetsLastUpdated(ctx context.Context, userID int64) (*time.Time, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func Test_UsersController_GetUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	expectedUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	mockRepo.On("Get", mock.Anything, int64(42)).Return(expectedUser, nil)

	req := httptest.NewRequest("GET", "/v1/users/42", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": "42"},
	}

	result, httpErr := controller.GetUser(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	user := result.(*repositories.User)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "Test User", user.Name)

	mockRepo.AssertExpectations(t)
}

func Test_UsersController_GetUser_MissingID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	req := httptest.NewRequest("GET", "/v1/users/", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{}, // No id parameter
	}

	result, httpErr := controller.GetUser(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "Must provide user id")
}

func Test_UsersController_GetUser_InvalidID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	req := httptest.NewRequest("GET", "/v1/users/invalid", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": "invalid"},
	}

	result, httpErr := controller.GetUser(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
	assert.Contains(t, httpErr.Error.Error(), "id must be a number")
}

func Test_UsersController_GetUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	mockRepo.On("Get", mock.Anything, int64(42)).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/users/42", nil)
	args := &web.HandlerArgs{
		Request: req,
		Params:  map[string]string{"id": "42"},
	}

	result, httpErr := controller.GetUser(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_UsersController_AddUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	user := repositories.User{
		ID:   42,
		Name: "New User",
	}

	mockRepo.On("Add", mock.Anything, mock.MatchedBy(func(u *repositories.User) bool {
		return u.ID == 42 && u.Name == "New User"
	})).Return(nil)

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/v1/users/", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.AddUser(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_UsersController_AddUser_InvalidJSON(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	req := httptest.NewRequest("POST", "/v1/users/", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.AddUser(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_UsersController_AddUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	user := repositories.User{
		ID:   42,
		Name: "New User",
	}

	mockRepo.On("Add", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/v1/users/", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
	}

	result, httpErr := controller.AddUser(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_UsersController_GetAssetStatus_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	userID := int64(42)
	lastUpdated := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)

	mockStatusRepo.On("GetAssetsLastUpdated", mock.Anything, userID).Return(&lastUpdated, nil)

	req := httptest.NewRequest("GET", "/v1/users/asset-status", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetAssetStatus(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*controllers.AssetStatusResponse)
	assert.Equal(t, &lastUpdated, response.LastUpdatedAt)
	assert.NotNil(t, response.NextUpdateAt)

	mockStatusRepo.AssertExpectations(t)
}

func Test_UsersController_GetAssetStatus_NeverUpdated(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	userID := int64(42)

	mockStatusRepo.On("GetAssetsLastUpdated", mock.Anything, userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/users/asset-status", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetAssetStatus(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	response := result.(*controllers.AssetStatusResponse)
	assert.Nil(t, response.LastUpdatedAt)
	assert.Nil(t, response.NextUpdateAt)

	mockStatusRepo.AssertExpectations(t)
}

func Test_UsersController_GetAssetStatus_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockStatusRepo := new(MockAssetsStatusRepository)
	mockRouter := &MockRouter{}

	controller := controllers.NewUsers(mockRouter, mockRepo, mockStatusRepo)

	userID := int64(42)

	mockStatusRepo.On("GetAssetsLastUpdated", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/users/asset-status", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetAssetStatus(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockStatusRepo.AssertExpectations(t)
}
