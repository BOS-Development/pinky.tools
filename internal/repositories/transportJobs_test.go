package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TransportJobsShouldCreateWithItems(t *testing.T) {
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
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30002187, 'Amarr', 20000020, 1.0) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60008494, 'Amarr VIII', 30002187, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 8200, Name: "Transport Job User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	notes := "Test transport"
	job := &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60008494,
		OriginSystemID:       30000142,
		DestinationSystemID:  30002187,
		TransportMethod:      "freighter",
		RoutePreference:      "secure",
		TotalVolumeM3:        50000,
		TotalCollateral:      1000000000,
		EstimatedCost:        5000000,
		Jumps:                9,
		FulfillmentType:      "self_haul",
		Notes:                &notes,
		Items: []*models.TransportJobItem{
			{TypeID: 34, Quantity: 100000, VolumeM3: 1000, EstimatedValue: 500000},
			{TypeID: 35, Quantity: 50000, VolumeM3: 500, EstimatedValue: 250000},
		},
	}

	created, err := jobsRepo.Create(ctx, job)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "planned", created.Status)
	assert.Equal(t, "freighter", created.TransportMethod)
	assert.Equal(t, "secure", created.RoutePreference)
	assert.Equal(t, 50000.0, created.TotalVolumeM3)
	assert.Equal(t, 1000000000.0, created.TotalCollateral)
	assert.Equal(t, 5000000.0, created.EstimatedCost)
	assert.Equal(t, 9, created.Jumps)
	assert.Equal(t, "self_haul", created.FulfillmentType)
	assert.Equal(t, &notes, created.Notes)
	assert.Len(t, created.Items, 2)
	assert.Equal(t, int64(34), created.Items[0].TypeID)
	assert.Equal(t, 100000, created.Items[0].Quantity)
}

func Test_TransportJobsShouldGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 8201, Name: "Job List User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	_, err = jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items:                []*models.TransportJobItem{},
	})
	require.NoError(t, err)

	jobs, err := jobsRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "Jita IV", jobs[0].OriginStationName)
	assert.Equal(t, "Jita", jobs[0].OriginSystemName)
}

func Test_TransportJobsShouldUpdateStatus(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 8202, Name: "Status Update User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	created, err := jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items:                []*models.TransportJobItem{},
	})
	require.NoError(t, err)
	assert.Equal(t, "planned", created.Status)

	// Move to in_transit
	err = jobsRepo.UpdateStatus(ctx, created.ID, user.ID, "in_transit")
	require.NoError(t, err)

	fetched, err := jobsRepo.GetByID(ctx, created.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "in_transit", fetched.Status)

	// Move to delivered
	err = jobsRepo.UpdateStatus(ctx, created.ID, user.ID, "delivered")
	require.NoError(t, err)

	fetched, err = jobsRepo.GetByID(ctx, created.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "delivered", fetched.Status)
}

func Test_TransportJobsShouldCancel(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 8203, Name: "Cancel Job User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	created, err := jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items:                []*models.TransportJobItem{},
	})
	require.NoError(t, err)

	err = jobsRepo.Cancel(ctx, created.ID, user.ID)
	require.NoError(t, err)

	fetched, err := jobsRepo.GetByID(ctx, created.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", fetched.Status)

	// Cancel again should fail
	err = jobsRepo.Cancel(ctx, created.ID, user.ID)
	assert.Error(t, err)
}

func Test_TransportJobsGetByUserBatchLoadsItems(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 0.9) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000125, true) ON CONFLICT DO NOTHING`)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 8204, Name: "Batch Items User"}
	err = userRepo.Add(ctx, user)
	require.NoError(t, err)

	// Create first job with 2 items
	job1, err := jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items: []*models.TransportJobItem{
			{TypeID: 34, Quantity: 1000, VolumeM3: 10, EstimatedValue: 5000},
			{TypeID: 35, Quantity: 2000, VolumeM3: 20, EstimatedValue: 6000},
		},
	})
	require.NoError(t, err)

	// Create second job with 1 item
	job2, err := jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items: []*models.TransportJobItem{
			{TypeID: 36, Quantity: 500, VolumeM3: 5, EstimatedValue: 1000},
		},
	})
	require.NoError(t, err)

	// Create third job with no items
	_, err = jobsRepo.Create(ctx, &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60003760,
		OriginSystemID:       30000142,
		DestinationSystemID:  30000142,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		FulfillmentType:      "self_haul",
		Items:                []*models.TransportJobItem{},
	})
	require.NoError(t, err)

	jobs, err := jobsRepo.GetByUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, jobs, 3)

	// Build a map by ID for stable assertions (GetByUser returns newest first)
	jobMap := map[int64]*models.TransportJob{}
	for _, j := range jobs {
		jobMap[j.ID] = j
	}

	// Verify job1 has 2 items
	require.Contains(t, jobMap, job1.ID)
	assert.Len(t, jobMap[job1.ID].Items, 2)
	assert.Equal(t, int64(34), jobMap[job1.ID].Items[0].TypeID)
	assert.Equal(t, int64(35), jobMap[job1.ID].Items[1].TypeID)

	// Verify job2 has 1 item
	require.Contains(t, jobMap, job2.ID)
	assert.Len(t, jobMap[job2.ID].Items, 1)
	assert.Equal(t, int64(36), jobMap[job2.ID].Items[0].TypeID)

	// Verify job3 has 0 items (empty slice, not nil)
	for _, j := range jobs {
		if j.ID != job1.ID && j.ID != job2.ID {
			assert.NotNil(t, j.Items)
			assert.Len(t, j.Items, 0)
		}
	}
}

func Test_TransportJobsShouldReturnEmptySlice(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	jobsRepo := repositories.NewTransportJobs(db)

	jobs, err := jobsRepo.GetByUser(context.Background(), 99999)
	require.NoError(t, err)
	assert.Len(t, jobs, 0)
}
