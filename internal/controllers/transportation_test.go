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

// --- Mock repositories ---

type MockTransportProfilesRepo struct{ mock.Mock }

func (m *MockTransportProfilesRepo) GetByUser(ctx context.Context, userID int64) ([]*models.TransportProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TransportProfile), args.Error(1)
}
func (m *MockTransportProfilesRepo) GetByID(ctx context.Context, id, userID int64) (*models.TransportProfile, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}
func (m *MockTransportProfilesRepo) GetDefaultByMethod(ctx context.Context, userID int64, method string) (*models.TransportProfile, error) {
	args := m.Called(ctx, userID, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}
func (m *MockTransportProfilesRepo) Create(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}
func (m *MockTransportProfilesRepo) Update(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportProfile), args.Error(1)
}
func (m *MockTransportProfilesRepo) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

type MockJFRoutesRepo struct{ mock.Mock }

func (m *MockJFRoutesRepo) GetByUser(ctx context.Context, userID int64) ([]*models.JFRoute, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.JFRoute), args.Error(1)
}
func (m *MockJFRoutesRepo) GetByID(ctx context.Context, id, userID int64) (*models.JFRoute, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JFRoute), args.Error(1)
}
func (m *MockJFRoutesRepo) Create(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error) {
	args := m.Called(ctx, route, systemCoords)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JFRoute), args.Error(1)
}
func (m *MockJFRoutesRepo) Update(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error) {
	args := m.Called(ctx, route, systemCoords)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JFRoute), args.Error(1)
}
func (m *MockJFRoutesRepo) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

type MockTransportJobsRepo struct{ mock.Mock }

func (m *MockTransportJobsRepo) GetByUser(ctx context.Context, userID int64) ([]*models.TransportJob, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TransportJob), args.Error(1)
}
func (m *MockTransportJobsRepo) GetByID(ctx context.Context, id, userID int64) (*models.TransportJob, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportJob), args.Error(1)
}
func (m *MockTransportJobsRepo) Create(ctx context.Context, job *models.TransportJob) (*models.TransportJob, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportJob), args.Error(1)
}
func (m *MockTransportJobsRepo) UpdateStatus(ctx context.Context, id, userID int64, status string) error {
	args := m.Called(ctx, id, userID, status)
	return args.Error(0)
}
func (m *MockTransportJobsRepo) SetQueueEntryID(ctx context.Context, id int64, queueEntryID int64) error {
	args := m.Called(ctx, id, queueEntryID)
	return args.Error(0)
}
func (m *MockTransportJobsRepo) Cancel(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

type MockTransportTriggerConfigRepo struct{ mock.Mock }

func (m *MockTransportTriggerConfigRepo) GetByUser(ctx context.Context, userID int64) ([]*models.TransportTriggerConfig, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TransportTriggerConfig), args.Error(1)
}
func (m *MockTransportTriggerConfigRepo) Upsert(ctx context.Context, c *models.TransportTriggerConfig) (*models.TransportTriggerConfig, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransportTriggerConfig), args.Error(1)
}

type MockTransportJobQueueRepo struct{ mock.Mock }

func (m *MockTransportJobQueueRepo) Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IndustryJobQueueEntry), args.Error(1)
}

type MockTransportMarketPricesRepo struct{ mock.Mock }

func (m *MockTransportMarketPricesRepo) GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

type MockTransportSolarSystemsRepo struct{ mock.Mock }

func (m *MockTransportSolarSystemsRepo) GetByIDs(ctx context.Context, ids []int64) ([]*models.SolarSystem, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SolarSystem), args.Error(1)
}
func (m *MockTransportSolarSystemsRepo) Search(ctx context.Context, query string, limit int) ([]*models.SolarSystem, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SolarSystem), args.Error(1)
}

type MockTransportEsiClient struct{ mock.Mock }

func (m *MockTransportEsiClient) GetRoute(ctx context.Context, origin, destination int64, flag string) ([]int32, error) {
	args := m.Called(ctx, origin, destination, flag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int32), args.Error(1)
}

// --- Helper ---

func newTransportController() (*controllers.Transportation, *MockTransportProfilesRepo, *MockJFRoutesRepo, *MockTransportJobsRepo, *MockTransportTriggerConfigRepo, *MockTransportJobQueueRepo, *MockTransportMarketPricesRepo, *MockTransportSolarSystemsRepo, *MockTransportEsiClient) {
	profilesRepo := &MockTransportProfilesRepo{}
	jfRoutesRepo := &MockJFRoutesRepo{}
	jobsRepo := &MockTransportJobsRepo{}
	triggerRepo := &MockTransportTriggerConfigRepo{}
	queueRepo := &MockTransportJobQueueRepo{}
	marketRepo := &MockTransportMarketPricesRepo{}
	solarSysRepo := &MockTransportSolarSystemsRepo{}
	esiClient := &MockTransportEsiClient{}

	c := controllers.NewTransportation(
		&MockRouter{},
		profilesRepo,
		jfRoutesRepo,
		jobsRepo,
		triggerRepo,
		queueRepo,
		marketRepo,
		solarSysRepo,
		esiClient,
	)

	return c, profilesRepo, jfRoutesRepo, jobsRepo, triggerRepo, queueRepo, marketRepo, solarSysRepo, esiClient
}

// --- Transport Profile Tests ---

func Test_TransportGetProfiles(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	expected := []*models.TransportProfile{
		{ID: 1, UserID: 42, Name: "My Freighter", TransportMethod: "freighter", CargoM3: 350000},
	}
	profilesRepo.On("GetByUser", mock.Anything, int64(42)).Return(expected, nil)

	userID := int64(42)
	result, httpErr := c.GetProfiles(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/profiles", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	profiles := result.([]*models.TransportProfile)
	assert.Len(t, profiles, 1)
	assert.Equal(t, "My Freighter", profiles[0].Name)
}

func Test_TransportGetProfilesError(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	profilesRepo.On("GetByUser", mock.Anything, int64(42)).Return(nil, errors.New("db error"))

	userID := int64(42)
	_, httpErr := c.GetProfiles(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/profiles", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TransportCreateProfile(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	profilesRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.TransportProfile")).Return(
		&models.TransportProfile{
			ID: 1, UserID: 42, Name: "New Profile", TransportMethod: "freighter",
			CargoM3: 350000, RoutePreference: "shortest", CollateralPriceBasis: "sell",
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"name":            "New Profile",
		"transportMethod": "freighter",
		"cargoM3":         350000,
	})

	userID := int64(42)
	result, httpErr := c.CreateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/profiles", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	profile := result.(*models.TransportProfile)
	assert.Equal(t, "New Profile", profile.Name)
	assert.Equal(t, 350000.0, profile.CargoM3)
}

func Test_TransportCreateProfileMissingName(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"transportMethod": "freighter",
		"cargoM3":         350000,
	})

	userID := int64(42)
	_, httpErr := c.CreateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/profiles", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportCreateProfileMissingMethod(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"name":    "Test",
		"cargoM3": 350000,
	})

	userID := int64(42)
	_, httpErr := c.CreateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/profiles", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportCreateProfileInvalidCargo(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"name":            "Test",
		"transportMethod": "freighter",
		"cargoM3":         0,
	})

	userID := int64(42)
	_, httpErr := c.CreateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/profiles", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportUpdateProfile(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	profilesRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.TransportProfile")).Return(
		&models.TransportProfile{
			ID: 1, UserID: 42, Name: "Updated Profile", TransportMethod: "freighter",
			CargoM3: 400000, RoutePreference: "secure", CollateralPriceBasis: "sell",
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"name":            "Updated Profile",
		"transportMethod": "freighter",
		"cargoM3":         400000,
		"routePreference": "secure",
	})

	userID := int64(42)
	result, httpErr := c.UpdateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/profiles/1", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	profile := result.(*models.TransportProfile)
	assert.Equal(t, "Updated Profile", profile.Name)
	assert.Equal(t, 400000.0, profile.CargoM3)
}

func Test_TransportUpdateProfileNotFound(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	profilesRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.TransportProfile")).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{
		"name":            "Test",
		"transportMethod": "freighter",
		"cargoM3":         350000,
	})

	userID := int64(42)
	_, httpErr := c.UpdateProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/profiles/999", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_TransportDeleteProfile(t *testing.T) {
	c, profilesRepo, _, _, _, _, _, _, _ := newTransportController()

	profilesRepo.On("Delete", mock.Anything, int64(1), int64(42)).Return(nil)

	userID := int64(42)
	result, httpErr := c.DeleteProfile(&web.HandlerArgs{
		Request: httptest.NewRequest("DELETE", "/v1/transport/profiles/1", nil),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]bool)
	assert.True(t, resp["success"])
}

// --- JF Route Tests ---

func Test_TransportGetJFRoutes(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, _, _ := newTransportController()

	expected := []*models.JFRoute{
		{ID: 1, UserID: 42, Name: "Jita to Amarr", TotalDistanceLY: 8.5},
	}
	jfRoutesRepo.On("GetByUser", mock.Anything, int64(42)).Return(expected, nil)

	userID := int64(42)
	result, httpErr := c.GetJFRoutes(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/jf-routes", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	routes := result.([]*models.JFRoute)
	assert.Len(t, routes, 1)
	assert.Equal(t, "Jita to Amarr", routes[0].Name)
}

func Test_TransportGetJFRoutesError(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, _, _ := newTransportController()

	jfRoutesRepo.On("GetByUser", mock.Anything, int64(42)).Return(nil, errors.New("db error"))

	userID := int64(42)
	_, httpErr := c.GetJFRoutes(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/jf-routes", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TransportCreateJFRoute(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, solarSysRepo, _ := newTransportController()

	x1, y1, z1 := 1.0e17, 2.0e17, 3.0e17
	x2, y2, z2 := 4.0e17, 5.0e17, 6.0e17

	solarSysRepo.On("GetByIDs", mock.Anything, mock.AnythingOfType("[]int64")).Return(
		[]*models.SolarSystem{
			{ID: 30000142, Name: "Jita", X: &x1, Y: &y1, Z: &z1},
			{ID: 30002187, Name: "Amarr", X: &x2, Y: &y2, Z: &z2},
		}, nil,
	)

	jfRoutesRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.JFRoute"), mock.AnythingOfType("map[int64]*models.SolarSystem")).Return(
		&models.JFRoute{
			ID: 1, UserID: 42, Name: "Jita to Amarr",
			OriginSystemID: 30000142, DestinationSystemID: 30002187,
			TotalDistanceLY: 8.5,
			Waypoints: []*models.JFRouteWaypoint{
				{ID: 1, RouteID: 1, Sequence: 0, SystemID: 30000142, DistanceLY: 0},
				{ID: 2, RouteID: 1, Sequence: 1, SystemID: 30002187, DistanceLY: 8.5},
			},
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"name":                "Jita to Amarr",
		"originSystemId":      30000142,
		"destinationSystemId": 30002187,
		"waypoints": []any{
			map[string]any{"sequence": 0, "systemId": 30000142},
			map[string]any{"sequence": 1, "systemId": 30002187},
		},
	})

	userID := int64(42)
	result, httpErr := c.CreateJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jf-routes", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	route := result.(*models.JFRoute)
	assert.Equal(t, "Jita to Amarr", route.Name)
	assert.Len(t, route.Waypoints, 2)
}

func Test_TransportCreateJFRouteMissingName(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"waypoints": []any{
			map[string]any{"sequence": 0, "systemId": 30000142},
			map[string]any{"sequence": 1, "systemId": 30002187},
		},
	})

	userID := int64(42)
	_, httpErr := c.CreateJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jf-routes", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportCreateJFRouteTooFewWaypoints(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"name": "Bad Route",
		"waypoints": []any{
			map[string]any{"sequence": 0, "systemId": 30000142},
		},
	})

	userID := int64(42)
	_, httpErr := c.CreateJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jf-routes", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportUpdateJFRoute(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, solarSysRepo, _ := newTransportController()

	x1, y1, z1 := 1.0e17, 2.0e17, 3.0e17
	x2, y2, z2 := 4.0e17, 5.0e17, 6.0e17

	solarSysRepo.On("GetByIDs", mock.Anything, mock.AnythingOfType("[]int64")).Return(
		[]*models.SolarSystem{
			{ID: 30000142, Name: "Jita", X: &x1, Y: &y1, Z: &z1},
			{ID: 30002187, Name: "Amarr", X: &x2, Y: &y2, Z: &z2},
		}, nil,
	)

	jfRoutesRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.JFRoute"), mock.AnythingOfType("map[int64]*models.SolarSystem")).Return(
		&models.JFRoute{
			ID: 1, UserID: 42, Name: "Updated Route",
			TotalDistanceLY: 9.0,
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"name":                "Updated Route",
		"originSystemId":      30000142,
		"destinationSystemId": 30002187,
		"waypoints": []any{
			map[string]any{"sequence": 0, "systemId": 30000142},
			map[string]any{"sequence": 1, "systemId": 30002187},
		},
	})

	userID := int64(42)
	result, httpErr := c.UpdateJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/jf-routes/1", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	route := result.(*models.JFRoute)
	assert.Equal(t, "Updated Route", route.Name)
}

func Test_TransportUpdateJFRouteNotFound(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, solarSysRepo, _ := newTransportController()

	solarSysRepo.On("GetByIDs", mock.Anything, mock.AnythingOfType("[]int64")).Return(
		[]*models.SolarSystem{}, nil,
	)

	jfRoutesRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.JFRoute"), mock.AnythingOfType("map[int64]*models.SolarSystem")).Return(nil, nil)

	body, _ := json.Marshal(map[string]any{
		"name":                "Test",
		"originSystemId":      30000142,
		"destinationSystemId": 30002187,
		"waypoints": []any{
			map[string]any{"sequence": 0, "systemId": 30000142},
			map[string]any{"sequence": 1, "systemId": 30002187},
		},
	})

	userID := int64(42)
	_, httpErr := c.UpdateJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/jf-routes/999", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "999"},
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func Test_TransportDeleteJFRoute(t *testing.T) {
	c, _, jfRoutesRepo, _, _, _, _, _, _ := newTransportController()

	jfRoutesRepo.On("Delete", mock.Anything, int64(1), int64(42)).Return(nil)

	userID := int64(42)
	result, httpErr := c.DeleteJFRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("DELETE", "/v1/transport/jf-routes/1", nil),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]bool)
	assert.True(t, resp["success"])
}

// --- Transport Jobs Tests ---

func Test_TransportGetJobs(t *testing.T) {
	c, _, _, jobsRepo, _, _, _, _, _ := newTransportController()

	expected := []*models.TransportJob{
		{ID: 1, UserID: 42, TransportMethod: "freighter", Status: "planned", FulfillmentType: "self_haul"},
	}
	jobsRepo.On("GetByUser", mock.Anything, int64(42)).Return(expected, nil)

	userID := int64(42)
	result, httpErr := c.GetJobs(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/jobs", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	jobs := result.([]*models.TransportJob)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "freighter", jobs[0].TransportMethod)
}

func Test_TransportGetJobsError(t *testing.T) {
	c, _, _, jobsRepo, _, _, _, _, _ := newTransportController()

	jobsRepo.On("GetByUser", mock.Anything, int64(42)).Return(nil, errors.New("db error"))

	userID := int64(42)
	_, httpErr := c.GetJobs(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/jobs", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TransportCreateJobFreighter(t *testing.T) {
	c, profilesRepo, _, jobsRepo, _, queueRepo, _, _, esiClient := newTransportController()

	profileID := int64(10)
	profilesRepo.On("GetByID", mock.Anything, int64(10), int64(42)).Return(
		&models.TransportProfile{
			ID: 10, UserID: 42, Name: "Freighter", TransportMethod: "freighter",
			CargoM3: 350000, RatePerM3PerJump: 800, CollateralRate: 0.01,
		}, nil,
	)

	// ESI returns 10-system route (9 jumps)
	route := []int32{30000142, 30000143, 30000144, 30000145, 30000146, 30000147, 30000148, 30000149, 30000150, 30002187}
	esiClient.On("GetRoute", mock.Anything, int64(30000142), int64(30002187), "shortest").Return(route, nil)

	jobsRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.TransportJob")).Return(
		&models.TransportJob{
			ID: 1, UserID: 42, Status: "planned", TransportMethod: "freighter",
			TransportProfileID: &profileID,
			Items:              []*models.TransportJobItem{},
		}, nil,
	)

	cost := 0.0
	transportJobID := int64(1)
	queueRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.IndustryJobQueueEntry")).Return(
		&models.IndustryJobQueueEntry{
			ID: 100, UserID: 42, Activity: "transport", EstimatedCost: &cost, TransportJobID: &transportJobID,
		}, nil,
	)

	jobsRepo.On("SetQueueEntryID", mock.Anything, int64(1), int64(100)).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"originStationId":      60003760,
		"destinationStationId": 60008494,
		"originSystemId":       30000142,
		"destinationSystemId":  30002187,
		"transportMethod":      "freighter",
		"fulfillmentType":      "self_haul",
		"transportProfileId":   10,
		"items": []any{
			map[string]any{"typeId": 34, "quantity": 100000, "volumeM3": 1000, "estimatedValue": 500000},
		},
	})

	userID := int64(42)
	result, httpErr := c.CreateJob(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	job := result.(*models.TransportJob)
	assert.Equal(t, int64(1), job.ID)
	queueRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*models.IndustryJobQueueEntry"))
	jobsRepo.AssertCalled(t, "SetQueueEntryID", mock.Anything, int64(1), int64(100))
}

func Test_TransportCreateJobMissingStations(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"transportMethod": "freighter",
	})

	userID := int64(42)
	_, httpErr := c.CreateJob(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportCreateJobMissingMethod(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"originStationId":      60003760,
		"destinationStationId": 60008494,
	})

	userID := int64(42)
	_, httpErr := c.CreateJob(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportCreateJobJF(t *testing.T) {
	c, profilesRepo, jfRoutesRepo, jobsRepo, _, queueRepo, marketRepo, _, _ := newTransportController()

	profileID := int64(20)
	jfRouteID := int64(5)
	fuelTypeID := int64(16274)
	fuelPerLY := 500.0
	sellPrice := 1200.0

	profilesRepo.On("GetByID", mock.Anything, int64(20), int64(42)).Return(
		&models.TransportProfile{
			ID: 20, UserID: 42, Name: "JF", TransportMethod: "jump_freighter",
			CargoM3: 320000, CollateralRate: 0.02, CollateralPriceBasis: "sell",
			FuelTypeID: &fuelTypeID, FuelPerLY: &fuelPerLY, FuelConservationLevel: 5,
		}, nil,
	)

	jfRoutesRepo.On("GetByID", mock.Anything, int64(5), int64(42)).Return(
		&models.JFRoute{
			ID: 5, UserID: 42, Name: "Test Route", TotalDistanceLY: 7.5,
			Waypoints: []*models.JFRouteWaypoint{
				{ID: 1, Sequence: 0, SystemID: 30000142, DistanceLY: 0},
				{ID: 2, Sequence: 1, SystemID: 30001000, DistanceLY: 3.5},
				{ID: 3, Sequence: 2, SystemID: 30002187, DistanceLY: 4.0},
			},
		}, nil,
	)

	marketRepo.On("GetAllJitaPrices", mock.Anything).Return(
		map[int64]*models.MarketPrice{
			16274: {SellPrice: &sellPrice},
		}, nil,
	)

	jobsRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.TransportJob")).Return(
		&models.TransportJob{
			ID: 2, UserID: 42, Status: "planned", TransportMethod: "jump_freighter",
			TransportProfileID: &profileID, JFRouteID: &jfRouteID,
			Items: []*models.TransportJobItem{},
		}, nil,
	)

	cost := 0.0
	transportJobID := int64(2)
	queueRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.IndustryJobQueueEntry")).Return(
		&models.IndustryJobQueueEntry{
			ID: 101, UserID: 42, Activity: "transport", EstimatedCost: &cost, TransportJobID: &transportJobID,
		}, nil,
	)
	jobsRepo.On("SetQueueEntryID", mock.Anything, int64(2), int64(101)).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"originStationId":      60003760,
		"destinationStationId": 60008494,
		"originSystemId":       30000142,
		"destinationSystemId":  30002187,
		"transportMethod":      "jump_freighter",
		"fulfillmentType":      "self_haul",
		"transportProfileId":   20,
		"jfRouteId":            5,
		"items": []any{
			map[string]any{"typeId": 34, "quantity": 50000, "volumeM3": 500, "estimatedValue": 250000},
		},
	})

	userID := int64(42)
	result, httpErr := c.CreateJob(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	job := result.(*models.TransportJob)
	assert.Equal(t, int64(2), job.ID)
	profilesRepo.AssertCalled(t, "GetByID", mock.Anything, int64(20), int64(42))
	jfRoutesRepo.AssertCalled(t, "GetByID", mock.Anything, int64(5), int64(42))
	marketRepo.AssertCalled(t, "GetAllJitaPrices", mock.Anything)
}

func Test_TransportUpdateJobStatus(t *testing.T) {
	c, _, _, jobsRepo, _, _, _, _, _ := newTransportController()

	jobsRepo.On("UpdateStatus", mock.Anything, int64(1), int64(42), "in_transit").Return(nil)

	body, _ := json.Marshal(map[string]any{
		"status": "in_transit",
	})

	userID := int64(42)
	result, httpErr := c.UpdateJobStatus(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs/1/status", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]bool)
	assert.True(t, resp["success"])
}

func Test_TransportUpdateJobStatusInvalid(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"status": "bad_status",
	})

	userID := int64(42)
	_, httpErr := c.UpdateJobStatus(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs/1/status", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportUpdateJobStatusDelivered(t *testing.T) {
	c, _, _, jobsRepo, _, _, _, _, _ := newTransportController()

	jobsRepo.On("UpdateStatus", mock.Anything, int64(1), int64(42), "delivered").Return(nil)

	body, _ := json.Marshal(map[string]any{
		"status": "delivered",
	})

	userID := int64(42)
	result, httpErr := c.UpdateJobStatus(&web.HandlerArgs{
		Request: httptest.NewRequest("POST", "/v1/transport/jobs/1/status", bytes.NewReader(body)),
		User:    &userID,
		Params:  map[string]string{"id": "1"},
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]bool)
	assert.True(t, resp["success"])
}

// --- Route Calculation Tests ---

func Test_TransportGetRoute(t *testing.T) {
	c, _, _, _, _, _, _, _, esiClient := newTransportController()

	route := []int32{30000142, 30000143, 30000144, 30000145, 30002187}
	esiClient.On("GetRoute", mock.Anything, int64(30000142), int64(30002187), "secure").Return(route, nil)

	userID := int64(42)
	result, httpErr := c.GetRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/route?origin=30000142&destination=30002187&flag=secure", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]any)
	assert.Equal(t, route, resp["route"])
	assert.Equal(t, 4, resp["jumps"])
}

func Test_TransportGetRouteDefaultFlag(t *testing.T) {
	c, _, _, _, _, _, _, _, esiClient := newTransportController()

	route := []int32{30000142, 30002187}
	esiClient.On("GetRoute", mock.Anything, int64(30000142), int64(30002187), "shortest").Return(route, nil)

	userID := int64(42)
	result, httpErr := c.GetRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/route?origin=30000142&destination=30002187", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	resp := result.(map[string]any)
	assert.Equal(t, 1, resp["jumps"])
}

func Test_TransportGetRouteMissingParams(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	userID := int64(42)
	_, httpErr := c.GetRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/route", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportGetRouteError(t *testing.T) {
	c, _, _, _, _, _, _, _, esiClient := newTransportController()

	esiClient.On("GetRoute", mock.Anything, int64(30000142), int64(30002187), "shortest").Return(nil, errors.New("esi error"))

	userID := int64(42)
	_, httpErr := c.GetRoute(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/route?origin=30000142&destination=30002187", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

// --- Solar System Search Tests ---

func Test_TransportSearchSystems(t *testing.T) {
	c, _, _, _, _, _, _, solarSysRepo, _ := newTransportController()

	expected := []*models.SolarSystem{
		{ID: 30000142, Name: "Jita", Security: 0.9},
		{ID: 30000143, Name: "Jita (alt)", Security: 0.8},
	}
	solarSysRepo.On("Search", mock.Anything, "Jita", 20).Return(expected, nil)

	userID := int64(42)
	result, httpErr := c.SearchSystems(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/systems/search?q=Jita", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	systems := result.([]*models.SolarSystem)
	assert.Len(t, systems, 2)
	assert.Equal(t, "Jita", systems[0].Name)
}

func Test_TransportSearchSystemsEmpty(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	userID := int64(42)
	result, httpErr := c.SearchSystems(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/systems/search", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	systems := result.([]*models.SolarSystem)
	assert.Len(t, systems, 0)
}

// --- Trigger Config Tests ---

func Test_TransportGetTriggerConfig(t *testing.T) {
	c, _, _, _, triggerRepo, _, _, _, _ := newTransportController()

	expected := []*models.TransportTriggerConfig{
		{UserID: 42, TriggerType: "plan_generation", DefaultFulfillment: "courier_contract"},
	}
	triggerRepo.On("GetByUser", mock.Anything, int64(42)).Return(expected, nil)

	userID := int64(42)
	result, httpErr := c.GetTriggerConfig(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/trigger-config", nil),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	configs := result.([]*models.TransportTriggerConfig)
	assert.Len(t, configs, 1)
	assert.Equal(t, "plan_generation", configs[0].TriggerType)
}

func Test_TransportGetTriggerConfigError(t *testing.T) {
	c, _, _, _, triggerRepo, _, _, _, _ := newTransportController()

	triggerRepo.On("GetByUser", mock.Anything, int64(42)).Return(nil, errors.New("db error"))

	userID := int64(42)
	_, httpErr := c.GetTriggerConfig(&web.HandlerArgs{
		Request: httptest.NewRequest("GET", "/v1/transport/trigger-config", nil),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
}

func Test_TransportUpsertTriggerConfig(t *testing.T) {
	c, _, _, _, triggerRepo, _, _, _, _ := newTransportController()

	triggerRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*models.TransportTriggerConfig")).Return(
		&models.TransportTriggerConfig{
			UserID:              42,
			TriggerType:         "manual",
			DefaultFulfillment:  "self_haul",
			AllowedFulfillments: []string{"self_haul", "courier_contract"},
			CourierRatePerM3:    500,
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"triggerType":         "manual",
		"defaultFulfillment":  "self_haul",
		"allowedFulfillments": []string{"self_haul", "courier_contract"},
		"courierRatePerM3":    500,
	})

	userID := int64(42)
	result, httpErr := c.UpsertTriggerConfig(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/trigger-config", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	config := result.(*models.TransportTriggerConfig)
	assert.Equal(t, "manual", config.TriggerType)
	assert.Equal(t, 500.0, config.CourierRatePerM3)
}

func Test_TransportUpsertTriggerConfigMissingType(t *testing.T) {
	c, _, _, _, _, _, _, _, _ := newTransportController()

	body, _ := json.Marshal(map[string]any{
		"defaultFulfillment": "self_haul",
	})

	userID := int64(42)
	_, httpErr := c.UpsertTriggerConfig(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/trigger-config", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

func Test_TransportUpsertTriggerConfigDefaults(t *testing.T) {
	c, _, _, _, triggerRepo, _, _, _, _ := newTransportController()

	triggerRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *models.TransportTriggerConfig) bool {
		return c.DefaultFulfillment == "self_haul" && len(c.AllowedFulfillments) == 1 && c.AllowedFulfillments[0] == "self_haul"
	})).Return(
		&models.TransportTriggerConfig{
			UserID:              42,
			TriggerType:         "plan_generation",
			DefaultFulfillment:  "self_haul",
			AllowedFulfillments: []string{"self_haul"},
		}, nil,
	)

	body, _ := json.Marshal(map[string]any{
		"triggerType": "plan_generation",
	})

	userID := int64(42)
	result, httpErr := c.UpsertTriggerConfig(&web.HandlerArgs{
		Request: httptest.NewRequest("PUT", "/v1/transport/trigger-config", bytes.NewReader(body)),
		User:    &userID,
	})

	assert.Nil(t, httpErr)
	config := result.(*models.TransportTriggerConfig)
	assert.Equal(t, "self_haul", config.DefaultFulfillment)
}
