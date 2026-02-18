package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_StationsShouldUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)
	stationsRepo := repositories.NewStations(db)

	// Create parent hierarchy
	regions := []models.Region{
		{ID: 10000002, Name: "The Forge"},
	}
	err = regionsRepo.Upsert(context.Background(), regions)
	assert.NoError(t, err)

	constellations := []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
	}
	err = constellationsRepo.Upsert(context.Background(), constellations)
	assert.NoError(t, err)

	systems := []models.SolarSystem{
		{ID: 30000142, Name: "Jita", ConstellationID: 20000020, Security: 0.9},
	}
	err = solarSystemsRepo.Upsert(context.Background(), systems)
	assert.NoError(t, err)

	stations := []models.Station{
		{
			ID:            60003760,
			Name:          "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
			SolarSystemID: 30000142,
			CorporationID: 1000035,
			IsNPC:         true,
		},
		{
			ID:            60003761,
			Name:          "Jita IV - Moon 4 - Some Other Station",
			SolarSystemID: 30000142,
			CorporationID: 1000035,
			IsNPC:         false,
		},
	}

	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)

	// Update one station
	stations[0].IsNPC = false

	// Add new station
	stations = append(stations, models.Station{
		ID:            60003762,
		Name:          "New Trade Hub",
		SolarSystemID: 30000142,
		CorporationID: 1000036,
		IsNPC:         true,
	})

	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)
}

func Test_StationsShouldGetStationsWithEmptyNames(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)
	stationsRepo := repositories.NewStations(db)

	// Create parent hierarchy
	err = regionsRepo.Upsert(context.Background(), []models.Region{
		{ID: 10000002, Name: "The Forge"},
	})
	assert.NoError(t, err)

	err = constellationsRepo.Upsert(context.Background(), []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
	})
	assert.NoError(t, err)

	err = solarSystemsRepo.Upsert(context.Background(), []models.SolarSystem{
		{ID: 30000142, Name: "Jita", ConstellationID: 20000020, Security: 0.9},
	})
	assert.NoError(t, err)

	// Insert stations: two NPC with empty names, one NPC with a name, one non-NPC with empty name
	stations := []models.Station{
		{ID: 60003760, Name: "", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
		{ID: 60003761, Name: "", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
		{ID: 60003762, Name: "Already Named Station", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
		{ID: 60003763, Name: "", SolarSystemID: 30000142, CorporationID: 1000036, IsNPC: false}, // player-owned, should be excluded
	}
	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)

	// Get stations with empty names — should only return NPC stations with empty names
	ids, err := stationsRepo.GetStationsWithEmptyNames(context.Background())
	assert.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.ElementsMatch(t, []int64{60003760, 60003761}, ids)
}

func Test_StationsShouldReturnEmptySliceWhenNoEmptyNames(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)
	stationsRepo := repositories.NewStations(db)

	// Create parent hierarchy
	err = regionsRepo.Upsert(context.Background(), []models.Region{
		{ID: 10000002, Name: "The Forge"},
	})
	assert.NoError(t, err)

	err = constellationsRepo.Upsert(context.Background(), []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
	})
	assert.NoError(t, err)

	err = solarSystemsRepo.Upsert(context.Background(), []models.SolarSystem{
		{ID: 30000142, Name: "Jita", ConstellationID: 20000020, Security: 0.9},
	})
	assert.NoError(t, err)

	// All stations have names
	stations := []models.Station{
		{ID: 60003760, Name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
	}
	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)

	ids, err := stationsRepo.GetStationsWithEmptyNames(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, ids)
}

func Test_StationsShouldUpdateNames(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)
	stationsRepo := repositories.NewStations(db)

	// Create parent hierarchy
	err = regionsRepo.Upsert(context.Background(), []models.Region{
		{ID: 10000002, Name: "The Forge"},
	})
	assert.NoError(t, err)

	err = constellationsRepo.Upsert(context.Background(), []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
	})
	assert.NoError(t, err)

	err = solarSystemsRepo.Upsert(context.Background(), []models.SolarSystem{
		{ID: 30000142, Name: "Jita", ConstellationID: 20000020, Security: 0.9},
	})
	assert.NoError(t, err)

	// Insert stations with empty names
	stations := []models.Station{
		{ID: 60003760, Name: "", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
		{ID: 60003761, Name: "", SolarSystemID: 30000142, CorporationID: 1000035, IsNPC: true},
	}
	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)

	// Verify they have empty names
	ids, err := stationsRepo.GetStationsWithEmptyNames(context.Background())
	assert.NoError(t, err)
	assert.Len(t, ids, 2)

	// Update names
	names := map[int64]string{
		60003760: "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
		60003761: "Jita IV - Moon 5 - Some Other Station",
	}
	err = stationsRepo.UpdateNames(context.Background(), names)
	assert.NoError(t, err)

	// Verify no more empty names
	ids, err = stationsRepo.GetStationsWithEmptyNames(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, ids)
}

func Test_StationsShouldHandleEmptyUpdateNames(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	stationsRepo := repositories.NewStations(db)

	// Should be a no-op, no error
	err = stationsRepo.UpdateNames(context.Background(), map[int64]string{})
	assert.NoError(t, err)

	err = stationsRepo.UpdateNames(context.Background(), nil)
	assert.NoError(t, err)
}

func Test_StationsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	stationsRepo := repositories.NewStations(db)

	err = stationsRepo.Upsert(context.Background(), []models.Station{})
	assert.NoError(t, err)

	err = stationsRepo.Upsert(context.Background(), nil)
	assert.NoError(t, err)
}
