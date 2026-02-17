package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_RegionsShouldUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)

	regions := []models.Region{
		{ID: 10000001, Name: "Derelik"},
		{ID: 10000002, Name: "The Forge"},
		{ID: 10000003, Name: "Vale of the Silent"},
	}

	err = regionsRepo.Upsert(context.Background(), regions)
	assert.NoError(t, err)

	// Update one region name
	regions[1].Name = "The Forge Updated"

	// Add a new region
	regions = append(regions, models.Region{ID: 10000004, Name: "The Bleak Lands"})

	err = regionsRepo.Upsert(context.Background(), regions)
	assert.NoError(t, err)

	// Verification would require a Get method, but we can at least verify no errors
}

func Test_RegionsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	regionsRepo := repositories.NewRegions(db)

	err = regionsRepo.Upsert(context.Background(), []models.Region{})
	assert.NoError(t, err)

	err = regionsRepo.Upsert(context.Background(), nil)
	assert.NoError(t, err)
}
