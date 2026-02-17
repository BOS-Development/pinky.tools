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

func Test_StationsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	stationsRepo := repositories.NewStations(db)

	err = stationsRepo.Upsert(context.Background(), []models.Station{})
	assert.NoError(t, err)

	err = stationsRepo.Upsert(context.Background(), nil)
	assert.NoError(t, err)
}
