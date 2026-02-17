package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_SolarSystemsShouldUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)
	solarSystemsRepo := repositories.NewSolarSystems(db)

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
		{ID: 30000143, Name: "Perimeter", ConstellationID: 20000020, Security: 1.0},
	}

	err = solarSystemsRepo.Upsert(context.Background(), systems)
	assert.NoError(t, err)

	// Update one system
	systems[0].Security = 0.95

	// Add new system
	systems = append(systems, models.SolarSystem{
		ID:              30000144,
		Name:            "New System",
		ConstellationID: 20000020,
		Security:        0.5,
	})

	err = solarSystemsRepo.Upsert(context.Background(), systems)
	assert.NoError(t, err)
}

func Test_SolarSystemsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	solarSystemsRepo := repositories.NewSolarSystems(db)

	err = solarSystemsRepo.Upsert(context.Background(), []models.SolarSystem{})
	assert.NoError(t, err)

	err = solarSystemsRepo.Upsert(context.Background(), nil)
	assert.NoError(t, err)
}
