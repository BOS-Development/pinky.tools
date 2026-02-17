package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_ConstellationsShouldUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)
	constellationsRepo := repositories.NewConstellations(db)

	// Create parent region first
	regions := []models.Region{
		{ID: 10000002, Name: "The Forge"},
	}
	err = regionsRepo.Upsert(context.Background(), regions)
	assert.NoError(t, err)

	constellations := []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
		{ID: 20000021, Name: "Lonetrek", RegionID: 10000002},
	}

	err = constellationsRepo.Upsert(context.Background(), constellations)
	assert.NoError(t, err)

	// Update one constellation
	constellations[0].Name = "Kimotoro Updated"

	// Add new constellation
	constellations = append(constellations, models.Constellation{
		ID:       20000022,
		Name:     "New Constellation",
		RegionID: 10000002,
	})

	err = constellationsRepo.Upsert(context.Background(), constellations)
	assert.NoError(t, err)
}

func Test_ConstellationsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	constellationsRepo := repositories.NewConstellations(db)

	err = constellationsRepo.Upsert(context.Background(), []models.Constellation{})
	assert.NoError(t, err)

	err = constellationsRepo.Upsert(context.Background(), nil)
	assert.NoError(t, err)
}
