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
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock contact rules repository
type MockContactRulesRepository struct {
	mock.Mock
}

func (m *MockContactRulesRepository) GetByUser(ctx context.Context, userID int64) ([]*models.ContactRule, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ContactRule), args.Error(1)
}

func (m *MockContactRulesRepository) Create(ctx context.Context, rule *models.ContactRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockContactRulesRepository) Delete(ctx context.Context, ruleID int64, userID int64) error {
	args := m.Called(ctx, ruleID, userID)
	return args.Error(0)
}

func (m *MockContactRulesRepository) DeleteAutoContactsForRule(ctx context.Context, ruleID int64) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockContactRulesRepository) SearchCorporations(ctx context.Context, query string) ([]*repositories.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.SearchResult), args.Error(1)
}

func (m *MockContactRulesRepository) SearchAlliances(ctx context.Context, query string) ([]*repositories.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.SearchResult), args.Error(1)
}

// Mock contact rule applier
type MockContactRuleApplier struct {
	mock.Mock
}

func (m *MockContactRuleApplier) ApplyRule(ctx context.Context, rule *models.ContactRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockContactRuleApplier) ApplyRulesForNewCorporation(ctx context.Context, userID int64, corpID int64, allianceID int64) error {
	args := m.Called(ctx, userID, corpID, allianceID)
	return args.Error(0)
}

// --- GetMyRules tests ---

func Test_ContactRulesController_GetMyRules_Success(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	expectedRules := []*models.ContactRule{
		{ID: 1, UserID: 42, RuleType: "corporation"},
		{ID: 2, UserID: 42, RuleType: "everyone"},
	}
	mockRepo.On("GetByUser", mock.Anything, userID).Return(expectedRules, nil)

	req := httptest.NewRequest("GET", "/v1/contact-rules", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetMyRules(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rules := result.([]*models.ContactRule)
	assert.Len(t, rules, 2)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_GetMyRules_RepositoryError(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("GetByUser", mock.Anything, userID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/contact-rules", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.GetMyRules(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_GetMyRules_Unauthorized(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	req := httptest.NewRequest("GET", "/v1/contact-rules", nil)
	args := &web.HandlerArgs{Request: req, User: nil}

	result, httpErr := controller.GetMyRules(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

// --- CreateRule tests ---

func Test_ContactRulesController_CreateRule_Corporation(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	entityID := int64(2001)
	reqBody := map[string]any{
		"ruleType":   "corporation",
		"entityId":   entityID,
		"entityName": "Test Corp",
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *models.ContactRule) bool {
		return r.UserID == 42 && r.RuleType == "corporation" && *r.EntityID == 2001
	})).Return(nil)
	mockApplier.On("ApplyRule", mock.Anything, mock.Anything).Return(nil).Maybe()

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rule := result.(*models.ContactRule)
	assert.Equal(t, "corporation", rule.RuleType)
	assert.Equal(t, int64(2001), *rule.EntityID)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_CreateRule_Everyone(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	reqBody := map[string]any{
		"ruleType": "everyone",
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *models.ContactRule) bool {
		return r.UserID == 42 && r.RuleType == "everyone" && r.EntityID == nil
	})).Return(nil)
	mockApplier.On("ApplyRule", mock.Anything, mock.Anything).Return(nil).Maybe()

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	rule := result.(*models.ContactRule)
	assert.Equal(t, "everyone", rule.RuleType)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_CreateRule_InvalidRuleType(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	reqBody := map[string]any{"ruleType": "invalid"}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ContactRulesController_CreateRule_MissingEntityID(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	reqBody := map[string]any{"ruleType": "corporation"}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ContactRulesController_CreateRule_InvalidJSON(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader([]byte("invalid json")))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ContactRulesController_CreateRule_RepositoryError(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	reqBody := map[string]any{"ruleType": "everyone"}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/contact-rules", bytes.NewReader(body))
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_CreateRule_Unauthorized(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	req := httptest.NewRequest("POST", "/v1/contact-rules", nil)
	args := &web.HandlerArgs{Request: req, User: nil}

	result, httpErr := controller.CreateRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

// --- DeleteRule tests ---

func Test_ContactRulesController_DeleteRule_Success(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("DeleteAutoContactsForRule", mock.Anything, int64(5)).Return(nil)
	mockRepo.On("Delete", mock.Anything, int64(5), userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/contact-rules/5", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "5"},
	}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, httpErr)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_DeleteRule_NotFound(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("DeleteAutoContactsForRule", mock.Anything, int64(99)).Return(nil)
	mockRepo.On("Delete", mock.Anything, int64(99), userID).Return(errors.New("contact rule not found or user is not the owner"))

	req := httptest.NewRequest("DELETE", "/v1/contact-rules/99", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "99"},
	}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_DeleteRule_InvalidID(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	req := httptest.NewRequest("DELETE", "/v1/contact-rules/abc", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "abc"},
	}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ContactRulesController_DeleteRule_MissingID(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	req := httptest.NewRequest("DELETE", "/v1/contact-rules/", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_ContactRulesController_DeleteRule_DeleteAutoContactsError(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("DeleteAutoContactsForRule", mock.Anything, int64(5)).Return(errors.New("database error"))

	req := httptest.NewRequest("DELETE", "/v1/contact-rules/5", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "5"},
	}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_DeleteRule_Unauthorized(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	req := httptest.NewRequest("DELETE", "/v1/contact-rules/5", nil)
	args := &web.HandlerArgs{Request: req, User: nil, Params: map[string]string{"id": "5"}}

	result, httpErr := controller.DeleteRule(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

// --- SearchCorporations tests ---

func Test_ContactRulesController_SearchCorporations_Success(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	expectedResults := []*repositories.SearchResult{
		{ID: 2001, Name: "Test Corp"},
	}
	mockRepo.On("SearchCorporations", mock.Anything, "Test").Return(expectedResults, nil)

	req := httptest.NewRequest("GET", "/v1/contact-rules/corporations?q=Test", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchCorporations(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	results := result.([]*repositories.SearchResult)
	assert.Len(t, results, 1)
	assert.Equal(t, "Test Corp", results[0].Name)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_SearchCorporations_EmptyQuery(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	req := httptest.NewRequest("GET", "/v1/contact-rules/corporations?q=", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchCorporations(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	results := result.([]*repositories.SearchResult)
	assert.Len(t, results, 0)
}

func Test_ContactRulesController_SearchCorporations_Error(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("SearchCorporations", mock.Anything, "Test").Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/contact-rules/corporations?q=Test", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchCorporations(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}

// --- SearchAlliances tests ---

func Test_ContactRulesController_SearchAlliances_Success(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	expectedResults := []*repositories.SearchResult{
		{ID: 5001, Name: "Test Alliance"},
	}
	mockRepo.On("SearchAlliances", mock.Anything, "Test").Return(expectedResults, nil)

	req := httptest.NewRequest("GET", "/v1/contact-rules/alliances?q=Test", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchAlliances(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	results := result.([]*repositories.SearchResult)
	assert.Len(t, results, 1)
	assert.Equal(t, "Test Alliance", results[0].Name)
	mockRepo.AssertExpectations(t)
}

func Test_ContactRulesController_SearchAlliances_EmptyQuery(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	req := httptest.NewRequest("GET", "/v1/contact-rules/alliances?q=", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchAlliances(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)
	results := result.([]*repositories.SearchResult)
	assert.Len(t, results, 0)
}

func Test_ContactRulesController_SearchAlliances_Error(t *testing.T) {
	mockRepo := new(MockContactRulesRepository)
	mockApplier := new(MockContactRuleApplier)
	mockRouter := &MockRouter{}

	controller := controllers.NewContactRules(mockRouter, mockRepo, mockApplier)

	userID := int64(42)
	mockRepo.On("SearchAlliances", mock.Anything, "Test").Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/v1/contact-rules/alliances?q=Test", nil)
	args := &web.HandlerArgs{Request: req, User: &userID}

	result, httpErr := controller.SearchAlliances(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	mockRepo.AssertExpectations(t)
}
