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

// --- Mock repository ---

type MockUserStationsRepository struct {
	mock.Mock
}

func (m *MockUserStationsRepository) GetByUser(ctx context.Context, userID int64) ([]*models.UserStation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserStation), args.Error(1)
}

func (m *MockUserStationsRepository) GetByID(ctx context.Context, id, userID int64) (*models.UserStation, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStation), args.Error(1)
}

func (m *MockUserStationsRepository) Create(ctx context.Context, station *models.UserStation) (*models.UserStation, error) {
	args := m.Called(ctx, station)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStation), args.Error(1)
}

func (m *MockUserStationsRepository) Update(ctx context.Context, station *models.UserStation) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockUserStationsRepository) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// --- Tests ---

func Test_UserStationsGetStations(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	expectedStations := []*models.UserStation{
		{
			ID: 1, UserID: 42, StationID: 99000001, Structure: "sotiyo", FacilityTax: 1.5,
			StationName: "Test Station", SolarSystemName: "Jita", Security: "high",
			Rigs: []*models.UserStationRig{
				{ID: 1, Category: "ship", Tier: "t1"},
			},
			Services:   []*models.UserStationService{},
			Activities: []string{"manufacturing"},
		},
	}

	mockRepo.On("GetByUser", mock.Anything, int64(42)).Return(expectedStations, nil)

	userID := int64(42)
	result, httpErr := controller.GetStations(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/user-stations", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	stations := result.([]*models.UserStation)
	assert.Len(t, stations, 1)
	assert.Equal(t, "sotiyo", stations[0].Structure)
}

func Test_UserStationsGetStationsError(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)
	mockRepo.On("GetByUser", mock.Anything, int64(42)).Return(nil, errors.New("db error"))

	userID := int64(42)
	_, httpErr := controller.GetStations(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/user-stations", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_UserStationsCreateStation(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.UserStation")).Return(
		&models.UserStation{
			ID: 1, UserID: 42, StationID: 99000001, Structure: "sotiyo", FacilityTax: 1.5,
			Rigs:     []*models.UserStationRig{},
			Services: []*models.UserStationService{},
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"stationId":   99000001,
		"structure":   "sotiyo",
		"facilityTax": 1.5,
		"rigs":        []any{},
		"services":    []any{},
	})

	userID := int64(42)
	result, httpErr := controller.CreateStation(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/user-stations", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	station := result.(*models.UserStation)
	assert.Equal(t, int64(1), station.ID)
	assert.Equal(t, "sotiyo", station.Structure)
}

func Test_UserStationsCreateStationMissingStationID(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	body, _ := json.Marshal(map[string]any{
		"structure":    "sotiyo",
		"facility_tax": 1.5,
	})

	userID := int64(42)
	_, httpErr := controller.CreateStation(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/user-stations", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_UserStationsUpdateStation(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.UserStation")).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"structure":   "azbel",
		"facilityTax": 2.0,
		"rigs": []any{
			map[string]any{"rigName": "Ship Rig", "category": "ship", "tier": "t2"},
		},
		"services": []any{},
	})

	userID := int64(42)
	result, httpErr := controller.UpdateStation(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/user-stations/1", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]string)
	assert.Equal(t, "updated", resp["status"])
}

func Test_UserStationsDeleteStation(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	mockRepo.On("Delete", mock.Anything, int64(1), int64(42)).Return(nil)

	userID := int64(42)
	result, httpErr := controller.DeleteStation(&web.HandlerArgs{
		Request: httptest.NewRequest("DELETE", "/v1/user-stations/1", nil),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]string)
	assert.Equal(t, "deleted", resp["status"])
}

func Test_UserStationsParseScan(t *testing.T) {
	mockRepo := &MockUserStationsRepository{}
	mockRouter := &MockRouter{}

	controller := controllers.NewUserStations(mockRouter, mockRepo)

	body, _ := json.Marshal(map[string]any{
		"scanText": "Rig Slots\nStandup XL-Set Ship Manufacturing Efficiency I\nService Slots\nStandup Manufacturing Plant I",
	})

	userID := int64(42)
	result, httpErr := controller.ParseScan(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/user-stations/parse-scan", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	scanResult := result.(*models.ScanResult)
	assert.Equal(t, "sotiyo", scanResult.Structure)
	assert.Len(t, scanResult.Rigs, 1)
	assert.Equal(t, "ship", scanResult.Rigs[0].Category)
	assert.Len(t, scanResult.Services, 1)
	assert.Equal(t, "manufacturing", scanResult.Services[0].Activity)
}
