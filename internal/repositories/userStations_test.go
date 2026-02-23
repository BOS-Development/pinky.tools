package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UserStationsShouldCreateAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	// Set up location data
	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000001, 'Test Player Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9000, Name: "Station Test User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	// Create station with rigs and services
	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000001,
		Structure:   "sotiyo",
		FacilityTax: 1.5,
		Rigs: []*models.UserStationRig{
			{RigName: "Standup XL-Set Ship Manufacturing Efficiency I", Category: "ship", Tier: "t1"},
			{RigName: "Standup XL-Set Structure and Component Manufacturing Efficiency I", Category: "component", Tier: "t1"},
		},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Manufacturing Plant I", Activity: "manufacturing"},
			{ServiceName: "Standup Capital Shipyard I", Activity: "manufacturing"},
		},
	})
	require.NoError(t, err)
	assert.NotZero(t, station.ID)

	// Get by user
	stations, err := stationsRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, stations, 1)

	s := stations[0]
	assert.Equal(t, station.ID, s.ID)
	assert.Equal(t, "sotiyo", s.Structure)
	assert.Equal(t, 1.5, s.FacilityTax)
	assert.Equal(t, "Test Player Station", s.StationName)
	assert.Equal(t, "Jita", s.SolarSystemName)
	assert.Equal(t, "high", s.Security)

	// Check rigs
	assert.Len(t, s.Rigs, 2)
	assert.Equal(t, "ship", s.Rigs[0].Category)
	assert.Equal(t, "t1", s.Rigs[0].Tier)
	assert.Equal(t, "component", s.Rigs[1].Category)

	// Check services
	assert.Len(t, s.Services, 2)
	assert.Equal(t, "manufacturing", s.Services[0].Activity)

	// Check activities
	assert.Len(t, s.Activities, 1)
	assert.Contains(t, s.Activities, "manufacturing")
}

func Test_UserStationsShouldGetByID(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000002, 'GetByID Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9010, Name: "GetByID Test"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000002,
		Structure:   "tatara",
		FacilityTax: 2.0,
		Rigs: []*models.UserStationRig{
			{RigName: "Standup L-Set Biochemical Reactor Efficiency II", Category: "reaction", Tier: "t2"},
		},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Biochemical Reactor I", Activity: "reaction"},
		},
	})
	require.NoError(t, err)

	fetched, err := stationsRepo.GetByID(ctx, station.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "tatara", fetched.Structure)
	assert.Len(t, fetched.Rigs, 1)
	assert.Equal(t, "reaction", fetched.Rigs[0].Category)
	assert.Equal(t, "t2", fetched.Rigs[0].Tier)
	assert.Len(t, fetched.Services, 1)
	assert.Equal(t, "reaction", fetched.Services[0].Activity)
	assert.Contains(t, fetched.Activities, "reaction")
}

func Test_UserStationsShouldReturnNilForWrongUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000003, 'Wrong User Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9020, Name: "Owner"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000003,
		Structure:   "raitaru",
		FacilityTax: 1.0,
		Rigs:        []*models.UserStationRig{},
		Services:    []*models.UserStationService{},
	})
	require.NoError(t, err)

	// Different user should not see this station
	fetched, err := stationsRepo.GetByID(ctx, station.ID, 9999)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func Test_UserStationsShouldUpdate(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000004, 'Update Test Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9030, Name: "Update Test"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000004,
		Structure:   "raitaru",
		FacilityTax: 1.0,
		Rigs: []*models.UserStationRig{
			{RigName: "Standup M-Set Ship Manufacturing Efficiency I", Category: "ship", Tier: "t1"},
		},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Manufacturing Plant I", Activity: "manufacturing"},
		},
	})
	require.NoError(t, err)

	// Update with new rigs and services
	err = stationsRepo.Update(ctx, &models.UserStation{
		ID:          station.ID,
		UserID:      user.ID,
		Structure:   "azbel",
		FacilityTax: 2.5,
		Rigs: []*models.UserStationRig{
			{RigName: "Standup L-Set Ship Manufacturing Efficiency II", Category: "ship", Tier: "t2"},
			{RigName: "Standup L-Set Equipment Manufacturing Efficiency I", Category: "equipment", Tier: "t1"},
		},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Manufacturing Plant I", Activity: "manufacturing"},
		},
	})
	require.NoError(t, err)

	fetched, err := stationsRepo.GetByID(ctx, station.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "azbel", fetched.Structure)
	assert.Equal(t, 2.5, fetched.FacilityTax)
	assert.Len(t, fetched.Rigs, 2)
	assert.Equal(t, "ship", fetched.Rigs[0].Category)
	assert.Equal(t, "t2", fetched.Rigs[0].Tier)
	assert.Equal(t, "equipment", fetched.Rigs[1].Category)
	assert.Len(t, fetched.Services, 1)
}

func Test_UserStationsShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000005, 'Delete Test Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9040, Name: "Delete Test"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000005,
		Structure:   "raitaru",
		FacilityTax: 1.0,
		Rigs: []*models.UserStationRig{
			{RigName: "Standup M-Set Ship Manufacturing Efficiency I", Category: "ship", Tier: "t1"},
		},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Manufacturing Plant I", Activity: "manufacturing"},
		},
	})
	require.NoError(t, err)

	err = stationsRepo.Delete(ctx, station.ID, user.ID)
	require.NoError(t, err)

	fetched, err := stationsRepo.GetByID(ctx, station.ID, user.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func Test_UserStationsShouldDeriveSecurityCorrectly(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	// Set up low-sec system
	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000043, 'Domain') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000600, 'Hed', 10000043) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30004000, 'LowSec System', 20000600, 0.3) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000006, 'LowSec Station', 30004000, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	// Set up null-sec system
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30004001, 'NullSec System', 20000600, -0.5) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000007, 'NullSec Station', 30004001, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9050, Name: "Security Test"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	// Create low-sec station
	lowStation, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID: user.ID, StationID: 99000006, Structure: "raitaru", FacilityTax: 1.0,
		Rigs: []*models.UserStationRig{}, Services: []*models.UserStationService{},
	})
	require.NoError(t, err)

	// Create null-sec station
	nullStation, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID: user.ID, StationID: 99000007, Structure: "raitaru", FacilityTax: 1.0,
		Rigs: []*models.UserStationRig{}, Services: []*models.UserStationService{},
	})
	require.NoError(t, err)

	lowFetched, err := stationsRepo.GetByID(ctx, lowStation.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "low", lowFetched.Security)

	nullFetched, err := stationsRepo.GetByID(ctx, nullStation.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "null", nullFetched.Security)
}

func Test_UserStationsMultipleActivities(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (99000008, 'Multi Activity Station', 30000142, 98000001, false) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	stationsRepo := repositories.NewUserStations(db)

	user := &repositories.User{ID: 9060, Name: "Multi Activity Test"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	station, err := stationsRepo.Create(ctx, &models.UserStation{
		UserID:      user.ID,
		StationID:   99000008,
		Structure:   "tatara",
		FacilityTax: 1.0,
		Rigs:        []*models.UserStationRig{},
		Services: []*models.UserStationService{
			{ServiceName: "Standup Manufacturing Plant I", Activity: "manufacturing"},
			{ServiceName: "Standup Composite Reactor I", Activity: "reaction"},
			{ServiceName: "Standup Biochemical Reactor I", Activity: "reaction"},
		},
	})
	require.NoError(t, err)

	fetched, err := stationsRepo.GetByID(ctx, station.ID, user.ID)
	require.NoError(t, err)
	assert.Len(t, fetched.Services, 3)
	// Activities should have unique values only
	assert.Len(t, fetched.Activities, 2)
	assert.Contains(t, fetched.Activities, "manufacturing")
	assert.Contains(t, fetched.Activities, "reaction")
}
