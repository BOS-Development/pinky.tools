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
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockPlayerCorporationRepository struct {
	mock.Mock
}

func (m *MockPlayerCorporationRepository) Upsert(ctx context.Context, corp repositories.PlayerCorporation) error {
	args := m.Called(ctx, corp)
	return args.Error(0)
}

func (m *MockPlayerCorporationRepository) Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.PlayerCorporation), args.Error(1)
}

type MockEsiClient struct {
	mock.Mock
}

func (m *MockEsiClient) GetCharacterCorporation(ctx context.Context, characterID int64, token, refresh string, expire time.Time) (*models.Corporation, error) {
	args := m.Called(ctx, characterID, token, refresh, expire)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Corporation), args.Error(1)
}

type MockCorporationAssetUpdater struct {
	mock.Mock
}

func (m *MockCorporationAssetUpdater) UpdateCorporationAssets(ctx context.Context, corp repositories.PlayerCorporation, userID int64) error {
	args := m.Called(ctx, corp, userID)
	return args.Error(0)
}

func Test_CorporationsController_Get_Success(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)
	expectedCorps := []repositories.PlayerCorporation{
		{ID: 2001, Name: "Corp 1", EsiScopes: "esi-assets.read_corporation_assets.v1"},
		{ID: 2002, Name: "Corp 2", EsiScopes: ""},
	}

	mockRepo.On("Get", mock.Anything, userID).Return(expectedCorps, nil)

	req := httptest.NewRequest("GET", "/v1/corporations", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Get(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	corps := result.([]controllers.PlayerCorporation)
	assert.Len(t, corps, 2)
	assert.Equal(t, int64(2001), corps[0].ID)
	assert.Equal(t, "Corp 1", corps[0].Name)
	assert.Equal(t, "esi-assets.read_corporation_assets.v1", corps[0].EsiScopes)

	mockRepo.AssertExpectations(t)
}

func Test_CorporationsController_Get_RepositoryError(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)
	mockRepo.On("Get", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/corporations", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Get(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_CorporationsController_Add_Success(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)
	character := repositories.Character{
		ID:                12345,
		EsiToken:          "token123",
		EsiRefreshToken:   "refresh456",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	esiCorp := &models.Corporation{
		ID:   2001,
		Name: "Test Corporation",
	}

	mockEsiClient.On("GetCharacterCorporation", mock.Anything, character.ID, character.EsiToken, character.EsiRefreshToken, mock.Anything).Return(esiCorp, nil)
	mockRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c repositories.PlayerCorporation) bool {
		return c.UserID == userID && c.ID == 2001
	})).Return(nil)

	// The updater is called in a goroutine after successful upsert
	mockUpdater.On("UpdateCorporationAssets", mock.Anything, mock.Anything, userID).Return(nil).Maybe()

	body, _ := json.Marshal(character)
	req := httptest.NewRequest("POST", "/v1/corporations", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Add(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockEsiClient.AssertExpectations(t)
}

func Test_CorporationsController_Add_InvalidJSON(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)

	req := httptest.NewRequest("POST", "/v1/corporations", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Add(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_CorporationsController_Add_EsiError(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)
	character := repositories.Character{
		ID:                12345,
		EsiToken:          "token123",
		EsiRefreshToken:   "refresh456",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	mockEsiClient.On("GetCharacterCorporation", mock.Anything, character.ID, character.EsiToken, character.EsiRefreshToken, mock.Anything).Return(nil, errors.New("ESI error"))

	body, _ := json.Marshal(character)
	req := httptest.NewRequest("POST", "/v1/corporations", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Add(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockEsiClient.AssertExpectations(t)
}

func Test_CorporationsController_Add_RepositoryError(t *testing.T) {
	mockRepo := new(MockPlayerCorporationRepository)
	mockEsiClient := new(MockEsiClient)
	mockUpdater := new(MockCorporationAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCorporations(mockRouter, mockEsiClient, mockRepo, mockUpdater, nil)

	userID := int64(42)
	character := repositories.Character{
		ID:                12345,
		EsiToken:          "token123",
		EsiRefreshToken:   "refresh456",
		EsiTokenExpiresOn: time.Now().Add(time.Hour),
	}

	esiCorp := &models.Corporation{
		ID:   2001,
		Name: "Test Corporation",
	}

	mockEsiClient.On("GetCharacterCorporation", mock.Anything, character.ID, character.EsiToken, character.EsiRefreshToken, mock.Anything).Return(esiCorp, nil)
	mockRepo.On("Upsert", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(character)
	req := httptest.NewRequest("POST", "/v1/corporations", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.Add(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
	mockEsiClient.AssertExpectations(t)
}
