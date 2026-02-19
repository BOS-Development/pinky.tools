package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
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
type MockCharacterRepository struct {
	mock.Mock
}

func (m *MockCharacterRepository) Get(ctx context.Context, id string) (*repositories.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Character), args.Error(1)
}

func (m *MockCharacterRepository) Add(ctx context.Context, character *repositories.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterRepository) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	args := m.Called(ctx, baseUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.Character), args.Error(1)
}

type MockCharacterAssetUpdater struct {
	mock.Mock
}

func (m *MockCharacterAssetUpdater) UpdateCharacterAssets(ctx context.Context, char *repositories.Character, userID int64) error {
	args := m.Called(ctx, char, userID)
	return args.Error(0)
}

func Test_CharactersController_GetAllCharacters_Success(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	expectedChars := []*repositories.Character{
		{ID: 12345, Name: "Character 1"},
		{ID: 12346, Name: "Character 2"},
	}

	mockRepo.On("GetAll", mock.Anything, userID).Return(expectedChars, nil)

	req := httptest.NewRequest("GET", "/v1/characters/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetAllCharacters(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	models := result.([]controllers.CharacterModel)
	assert.Len(t, models, 2)
	assert.Equal(t, int64(12345), models[0].ID)
	assert.Equal(t, "Character 1", models[0].Name)

	mockRepo.AssertExpectations(t)
}

func Test_CharactersController_GetAllCharacters_RepositoryError(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	mockRepo.On("GetAll", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/characters/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.GetAllCharacters(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_CharactersController_GetCharacter_Success(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	expectedChar := &repositories.Character{
		ID:   12345,
		Name: "Test Character",
	}

	mockRepo.On("Get", mock.Anything, "12345").Return(expectedChar, nil)

	req := httptest.NewRequest("GET", "/v1/characters/12345", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "12345"},
	}

	result, httpErr := controller.GetCharacter(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	char := result.(*repositories.Character)
	assert.Equal(t, int64(12345), char.ID)
	assert.Equal(t, "Test Character", char.Name)

	mockRepo.AssertExpectations(t)
}

func Test_CharactersController_GetCharacter_MissingID(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)

	req := httptest.NewRequest("GET", "/v1/characters/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{}, // No ID
	}

	result, httpErr := controller.GetCharacter(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_CharactersController_GetCharacter_NotFound(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	mockRepo.On("Get", mock.Anything, "99999").Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/characters/99999", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "99999"},
	}

	result, httpErr := controller.GetCharacter(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

func Test_CharactersController_AddCharacter_Success(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	character := repositories.Character{
		ID:   12345,
		Name: "New Character",
	}

	mockRepo.On("Add", mock.Anything, mock.MatchedBy(func(c *repositories.Character) bool {
		return c.UserID == userID && c.ID == 12345
	})).Return(nil)

	// The updater is called in a goroutine after successful add
	mockUpdater.On("UpdateCharacterAssets", mock.Anything, mock.Anything, userID).Return(nil).Maybe()

	body, _ := json.Marshal(character)
	req := httptest.NewRequest("POST", "/v1/characters/", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.AddCharacter(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func Test_CharactersController_AddCharacter_InvalidJSON(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)

	req := httptest.NewRequest("POST", "/v1/characters/", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.AddCharacter(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_CharactersController_AddCharacter_RepositoryError(t *testing.T) {
	mockRepo := new(MockCharacterRepository)
	mockUpdater := new(MockCharacterAssetUpdater)
	mockRouter := &MockRouter{}

	controller := controllers.NewCharacters(mockRouter, mockRepo, mockUpdater, nil, nil)

	userID := int64(42)
	character := repositories.Character{
		ID:   12345,
		Name: "New Character",
	}

	mockRepo.On("Add", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(character)
	req := httptest.NewRequest("POST", "/v1/characters/", bytes.NewReader(body))
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
	}

	result, httpErr := controller.AddCharacter(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}
