package updaters_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type MockPiUserRepository struct{ mock.Mock }

func (m *MockPiUserRepository) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type MockPiCharacterRepository struct{ mock.Mock }

func (m *MockPiCharacterRepository) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	args := m.Called(ctx, baseUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.Character), args.Error(1)
}

func (m *MockPiCharacterRepository) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	args := m.Called(ctx, id, userID, token, refreshToken, expiresOn)
	return args.Error(0)
}

type MockPiPlanetsRepository struct{ mock.Mock }

func (m *MockPiPlanetsRepository) UpsertPlanets(ctx context.Context, characterID, userID int64, planets []*models.PiPlanet) error {
	args := m.Called(ctx, characterID, userID, planets)
	return args.Error(0)
}

func (m *MockPiPlanetsRepository) UpsertColony(ctx context.Context, characterID, planetID int64, pins []*models.PiPin, contents []*models.PiPinContent, routes []*models.PiRoute) error {
	args := m.Called(ctx, characterID, planetID, pins, contents, routes)
	return args.Error(0)
}

func (m *MockPiPlanetsRepository) GetPlanetsForUser(ctx context.Context, userID int64) ([]*models.PiPlanet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PiPlanet), args.Error(1)
}

func (m *MockPiPlanetsRepository) GetPinsForPlanets(ctx context.Context, userID int64) ([]*models.PiPin, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PiPin), args.Error(1)
}

func (m *MockPiPlanetsRepository) GetPinContentsForUser(ctx context.Context, userID int64) ([]*models.PiPinContent, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PiPinContent), args.Error(1)
}

func (m *MockPiPlanetsRepository) UpdateStallNotifiedAt(ctx context.Context, characterID, planetID int64, notifiedAt *time.Time) error {
	args := m.Called(ctx, characterID, planetID, notifiedAt)
	return args.Error(0)
}


type MockPiSolarSystemRepository struct{ mock.Mock }

func (m *MockPiSolarSystemRepository) GetNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]string), args.Error(1)
}

type MockPiSchematicRepository struct{ mock.Mock }

func (m *MockPiSchematicRepository) GetAllSchematics(ctx context.Context) ([]*models.SdePlanetSchematic, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SdePlanetSchematic), args.Error(1)
}

func (m *MockPiSchematicRepository) GetAllSchematicTypes(ctx context.Context) ([]*models.SdePlanetSchematicType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SdePlanetSchematicType), args.Error(1)
}

type MockPiItemTypeRepository struct{ mock.Mock }

func (m *MockPiItemTypeRepository) GetNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]string), args.Error(1)
}

type MockPiStallNotifier struct{ mock.Mock }

func (m *MockPiStallNotifier) NotifyPiStalls(ctx context.Context, userID int64, alerts []*updaters.PiStallAlert) {
	m.Called(ctx, userID, alerts)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// piTestMocks holds all dependencies used in stall detection tests.
type piTestMocks struct {
	piRepo        *MockPiPlanetsRepository
	schematicRepo *MockPiSchematicRepository
	itemTypeRepo  *MockPiItemTypeRepository
	systemRepo    *MockPiSolarSystemRepository
	charRepo      *MockPiCharacterRepository
	userRepo      *MockPiUserRepository
	notifier      *MockPiStallNotifier
}

func setupPiUpdater() (*updaters.PiUpdater, *piTestMocks) {
	mocks := &piTestMocks{
		piRepo:        new(MockPiPlanetsRepository),
		schematicRepo: new(MockPiSchematicRepository),
		itemTypeRepo:  new(MockPiItemTypeRepository),
		systemRepo:    new(MockPiSolarSystemRepository),
		charRepo:      new(MockPiCharacterRepository),
		userRepo:      new(MockPiUserRepository),
		notifier:      new(MockPiStallNotifier),
	}

	updater := updaters.NewPiUpdater(
		mocks.userRepo,
		mocks.charRepo,
		mocks.piRepo,
		nil, // esiClient — not needed for stall tests
		mocks.systemRepo,
		mocks.schematicRepo,
		mocks.itemTypeRepo,
	)
	updater.WithStallNotifier(mocks.notifier)

	return updater, mocks
}

// schematicID 101, cycleTime 3600s, input: typeID 3000, qty 40 → 40 units/hr
var testSchematic = &models.SdePlanetSchematic{SchematicID: 101, Name: "Water Electrolysis", CycleTime: 3600}
var testSchematicInput = &models.SdePlanetSchematicType{SchematicID: 101, TypeID: 3000, Quantity: 40, IsInput: true}
var testSchematicOutput = &models.SdePlanetSchematicType{SchematicID: 101, TypeID: 3001, Quantity: 5, IsInput: false}

func baseSetupSchematics(m *piTestMocks) {
	m.schematicRepo.On("GetAllSchematics", mock.Anything).Return([]*models.SdePlanetSchematic{testSchematic}, nil)
	m.schematicRepo.On("GetAllSchematicTypes", mock.Anything).Return(
		[]*models.SdePlanetSchematicType{testSchematicInput, testSchematicOutput}, nil)
}

func baseSetupSystemNames(m *piTestMocks) {
	m.systemRepo.On("GetNames", mock.Anything, mock.Anything).Return(map[int64]string{10000001: "Jita"}, nil)
}

func baseCharacters(m *piTestMocks) {
	m.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
}

func schematicID(id int) *int {
	return &id
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Test_PiCheckStalls_NoAlert_WhenInputsNotDepleted verifies that no alert fires
// when the projected depletion time is still in the future.
func Test_PiCheckStalls_NoAlert_WhenInputsNotDepleted(t *testing.T) {
	updater, mocks := setupPiUpdater()

	// Planet last updated 1 hour ago
	lastUpdate := time.Now().Add(-1 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:   10,
		UserID:        1,
		PlanetID:      100,
		PlanetType:    "barren",
		SolarSystemID: 10000001,
		LastUpdate:    lastUpdate,
	}

	sid := 101
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: &sid,
	}

	// 40 units/hr consumption; stock = 400 → 10 hours remaining → depletion 9h in future
	contents := []*models.PiPinContent{
		{CharacterID: 10, PlanetID: 100, PinID: 9999, TypeID: 3000, Amount: 400},
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return(contents, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	mocks.notifier.AssertNotCalled(t, "NotifyPiStalls")
}

// Test_PiCheckStalls_Alert_WhenInputsDepleted verifies that an alert fires when
// the projected depletion time is in the past.
func Test_PiCheckStalls_Alert_WhenInputsDepleted(t *testing.T) {
	updater, mocks := setupPiUpdater()

	// Planet last updated 5 hours ago
	lastUpdate := time.Now().Add(-5 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:   10,
		UserID:        1,
		PlanetID:      100,
		PlanetType:    "barren",
		SolarSystemID: 10000001,
		LastUpdate:    lastUpdate,
	}

	sid := 101
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: &sid,
	}

	// 40 units/hr consumption; stock = 80 → 2 hours → depleted 3h ago (5h - 2h = 3h in past)
	contents := []*models.PiPinContent{
		{CharacterID: 10, PlanetID: 100, PinID: 9999, TypeID: 3000, Amount: 80},
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return(contents, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)
	mocks.itemTypeRepo.On("GetNames", mock.Anything, []int64{3000}).Return(map[int64]string{3000: "Water"}, nil)
	mocks.piRepo.On("UpdateStallNotifiedAt", mock.Anything, int64(10), int64(100), mock.AnythingOfType("*time.Time")).Return(nil)

	var capturedAlerts []*updaters.PiStallAlert
	mocks.notifier.On("NotifyPiStalls", mock.Anything, int64(1), mock.Anything).
		Run(func(args mock.Arguments) {
			capturedAlerts = args.Get(2).([]*updaters.PiStallAlert)
		})

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	mocks.notifier.AssertCalled(t, "NotifyPiStalls", mock.Anything, int64(1), mock.Anything)
	assert.Len(t, capturedAlerts, 1)
	assert.NotNil(t, capturedAlerts[0].DepletionTime)
	assert.Equal(t, "Water", capturedAlerts[0].DepletedInputName)
	assert.True(t, capturedAlerts[0].DepletionTime.Before(time.Now()))
	mocks.piRepo.AssertExpectations(t)
}

// Test_PiCheckStalls_CombinedAlert_ExtractorAndDepletion verifies that when both
// an expired extractor and depleted inputs exist, a single combined alert is sent.
func Test_PiCheckStalls_CombinedAlert_ExtractorAndDepletion(t *testing.T) {
	updater, mocks := setupPiUpdater()

	lastUpdate := time.Now().Add(-5 * time.Hour)
	expiredTime := time.Now().Add(-1 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:   10,
		UserID:        1,
		PlanetID:      100,
		PlanetType:    "barren",
		SolarSystemID: 10000001,
		LastUpdate:    lastUpdate,
	}

	extractorPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       2001,
		PinCategory: "extractor",
		ExpiryTime:  &expiredTime,
	}

	sid := 101
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: &sid,
	}

	// 40 units/hr, stock = 80 → depleted 3h ago
	contents := []*models.PiPinContent{
		{CharacterID: 10, PlanetID: 100, PinID: 9999, TypeID: 3000, Amount: 80},
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{extractorPin, factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return(contents, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)
	mocks.itemTypeRepo.On("GetNames", mock.Anything, []int64{3000}).Return(map[int64]string{3000: "Water"}, nil)
	mocks.piRepo.On("UpdateStallNotifiedAt", mock.Anything, int64(10), int64(100), mock.AnythingOfType("*time.Time")).Return(nil)

	var capturedAlerts []*updaters.PiStallAlert
	mocks.notifier.On("NotifyPiStalls", mock.Anything, int64(1), mock.Anything).
		Run(func(args mock.Arguments) {
			capturedAlerts = args.Get(2).([]*updaters.PiStallAlert)
		})

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	mocks.notifier.AssertCalled(t, "NotifyPiStalls", mock.Anything, int64(1), mock.Anything)
	assert.Len(t, capturedAlerts, 1)
	assert.Len(t, capturedAlerts[0].StalledPins, 1)
	assert.Equal(t, "extractor", capturedAlerts[0].StalledPins[0].PinCategory)
	assert.NotNil(t, capturedAlerts[0].DepletionTime)
	assert.Equal(t, "Water", capturedAlerts[0].DepletedInputName)
}

// Test_PiCheckStalls_NoAlert_WhenFactoryHasNoSchematic verifies that factory pins
// without a schematic are skipped and produce no depletion alert.
func Test_PiCheckStalls_NoAlert_WhenFactoryHasNoSchematic(t *testing.T) {
	updater, mocks := setupPiUpdater()

	lastUpdate := time.Now().Add(-10 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:   10,
		UserID:        1,
		PlanetID:      100,
		PlanetType:    "barren",
		SolarSystemID: 10000001,
		LastUpdate:    lastUpdate,
	}

	// Factory with no schematic — no depletion can be calculated
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: nil,
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return([]*models.PiPinContent{}, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	mocks.notifier.AssertNotCalled(t, "NotifyPiStalls")
}

// Test_PiCheckStalls_Dedup_NoAlert_WhenAlreadyNotified verifies that a planet
// that already has last_stall_notified_at set does not generate a second alert.
func Test_PiCheckStalls_Dedup_NoAlert_WhenAlreadyNotified(t *testing.T) {
	updater, mocks := setupPiUpdater()

	lastUpdate := time.Now().Add(-5 * time.Hour)
	notifiedAt := time.Now().Add(-2 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:         10,
		UserID:              1,
		PlanetID:            100,
		PlanetType:          "barren",
		SolarSystemID:       10000001,
		LastUpdate:          lastUpdate,
		LastStallNotifiedAt: &notifiedAt,
	}

	sid := 101
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: &sid,
	}

	// Depleted (80 units / 40 per hr = 2h, but last update was 5h ago → depleted 3h ago)
	contents := []*models.PiPinContent{
		{CharacterID: 10, PlanetID: 100, PinID: 9999, TypeID: 3000, Amount: 80},
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return(contents, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	// Must NOT fire a second notification
	mocks.notifier.AssertNotCalled(t, "NotifyPiStalls")
}

// Test_PiCheckStalls_Recovery_ClearsNotifiedAt verifies that when a previously-notified
// planet recovers (no issues), last_stall_notified_at is cleared.
func Test_PiCheckStalls_Recovery_ClearsNotifiedAt(t *testing.T) {
	updater, mocks := setupPiUpdater()

	lastUpdate := time.Now().Add(-1 * time.Hour)
	notifiedAt := time.Now().Add(-3 * time.Hour)
	planet := &models.PiPlanet{
		CharacterID:         10,
		UserID:              1,
		PlanetID:            100,
		PlanetType:          "barren",
		SolarSystemID:       10000001,
		LastUpdate:          lastUpdate,
		LastStallNotifiedAt: &notifiedAt,
	}

	sid := 101
	factoryPin := &models.PiPin{
		CharacterID: 10,
		PlanetID:    100,
		PinID:       1001,
		PinCategory: "factory",
		SchematicID: &sid,
	}

	// Plenty of stock: 40 units/hr, 2000 units → 50h remaining → depletion well in future
	contents := []*models.PiPinContent{
		{CharacterID: 10, PlanetID: 100, PinID: 9999, TypeID: 3000, Amount: 2000},
	}

	mocks.charRepo.On("GetAll", mock.Anything, int64(1)).Return([]*repositories.Character{}, nil)
	mocks.piRepo.On("GetPlanetsForUser", mock.Anything, int64(1)).Return([]*models.PiPlanet{planet}, nil)
	mocks.piRepo.On("GetPinsForPlanets", mock.Anything, int64(1)).Return([]*models.PiPin{factoryPin}, nil)
	mocks.piRepo.On("GetPinContentsForUser", mock.Anything, int64(1)).Return(contents, nil)
	baseSetupSchematics(mocks)
	baseSetupSystemNames(mocks)
	// Expect UpdateStallNotifiedAt called with nil to clear
	mocks.piRepo.On("UpdateStallNotifiedAt", mock.Anything, int64(10), int64(100), (*time.Time)(nil)).Return(nil)

	err := updater.UpdateUserPlanets(context.Background(), 1)
	assert.NoError(t, err)

	mocks.notifier.AssertNotCalled(t, "NotifyPiStalls")
	mocks.piRepo.AssertCalled(t, "UpdateStallNotifiedAt", mock.Anything, int64(10), int64(100), (*time.Time)(nil))
}
