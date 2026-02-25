package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSolarSystemsForJFTests(t *testing.T, db interface{ ExecContext(ctx context.Context, query string, args ...interface{}) (interface{ RowsAffected() (int64, error) }, error) }) {
	// This helper is inlined below to avoid interface complexity
}

func Test_JFRoutesShouldCreateWithWaypointsAndDistances(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	// Set up location data
	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	// Insert solar systems with coordinates (in meters)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security, x, y, z) VALUES (30000142, 'Jita', 20000020, 0.9, 1.28e+17, 6.08e+16, 1.12e+17) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security, x, y, z) VALUES (30002813, 'Ignoitton', 20000020, 0.5, 1.20e+17, 5.80e+16, 1.08e+17) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security, x, y, z) VALUES (30001198, 'GE-8JV', 20000020, -0.3, 0.98e+17, 4.50e+16, 0.85e+17) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	routesRepo := repositories.NewJFRoutes(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)

	user := &repositories.User{ID: 8100, Name: "JF Route User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	// Get solar system coordinates
	systems, err := solarSystemsRepo.GetByIDs(ctx, []int64{30000142, 30002813, 30001198})
	require.NoError(t, err)

	systemCoords := make(map[int64]*models.SolarSystem)
	for _, s := range systems {
		systemCoords[s.ID] = s
	}

	route := &models.JFRoute{
		UserID:              user.ID,
		Name:                "Jita → GE-8JV",
		OriginSystemID:      30000142,
		DestinationSystemID: 30001198,
		Waypoints: []*models.JFRouteWaypoint{
			{Sequence: 0, SystemID: 30000142}, // Origin
			{Sequence: 1, SystemID: 30002813}, // Cyno stop
			{Sequence: 2, SystemID: 30001198}, // Destination
		},
	}

	created, err := routesRepo.Create(ctx, route, systemCoords)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Jita → GE-8JV", created.Name)
	assert.Equal(t, int64(30000142), created.OriginSystemID)
	assert.Equal(t, int64(30001198), created.DestinationSystemID)
	assert.Greater(t, created.TotalDistanceLY, 0.0)
	assert.Len(t, created.Waypoints, 3)

	// First waypoint distance is 0 (origin)
	assert.Equal(t, 0.0, created.Waypoints[0].DistanceLY)
	// Subsequent waypoints have positive distances
	assert.Greater(t, created.Waypoints[1].DistanceLY, 0.0)
	assert.Greater(t, created.Waypoints[2].DistanceLY, 0.0)
}

func Test_JFRoutesShouldGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30001198, 'GE-8JV', 20000020, -0.3) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	routesRepo := repositories.NewJFRoutes(db)

	user := &repositories.User{ID: 8101, Name: "JF Routes List User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	route := &models.JFRoute{
		UserID:              user.ID,
		Name:                "Test Route",
		OriginSystemID:      30000142,
		DestinationSystemID: 30001198,
		Waypoints: []*models.JFRouteWaypoint{
			{Sequence: 0, SystemID: 30000142},
			{Sequence: 1, SystemID: 30001198},
		},
	}

	_, err = routesRepo.Create(ctx, route, map[int64]*models.SolarSystem{})
	require.NoError(t, err)

	routes, err := routesRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, routes, 1)
	assert.Equal(t, "Test Route", routes[0].Name)
	assert.Equal(t, "Jita", routes[0].OriginSystemName)
	assert.Equal(t, "GE-8JV", routes[0].DestinationSystemName)
	assert.Len(t, routes[0].Waypoints, 2)
}

func Test_JFRoutesShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	routesRepo := repositories.NewJFRoutes(db)

	user := &repositories.User{ID: 8102, Name: "JF Route Delete User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	created, err := routesRepo.Create(ctx, &models.JFRoute{
		UserID:              user.ID,
		Name:                "Delete Me",
		OriginSystemID:      30000142,
		DestinationSystemID: 30000142,
		Waypoints: []*models.JFRouteWaypoint{
			{Sequence: 0, SystemID: 30000142},
		},
	}, map[int64]*models.SolarSystem{})
	require.NoError(t, err)

	err = routesRepo.Delete(ctx, created.ID, user.ID)
	require.NoError(t, err)

	routes, err := routesRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, routes, 0)
}

func Test_JFRoutesCalculateDistanceLY(t *testing.T) {
	// Test the LY distance calculation directly
	// Using known coordinates for a rough check
	x1, y1, z1 := 0.0, 0.0, 0.0
	x2, y2, z2 := 9.461e+15, 0.0, 0.0 // Exactly 1 LY in meters

	distance := repositories.CalculateDistanceLY(x1, y1, z1, x2, y2, z2)
	assert.InDelta(t, 1.0, distance, 0.001)

	// Zero distance
	distance = repositories.CalculateDistanceLY(x1, y1, z1, x1, y1, z1)
	assert.Equal(t, 0.0, distance)
}

func Test_JFRoutesShouldReturnEmptySlice(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	routesRepo := repositories.NewJFRoutes(db)

	routes, err := routesRepo.GetByUser(context.Background(), 99999)
	require.NoError(t, err)
	assert.Len(t, routes, 0)
}
