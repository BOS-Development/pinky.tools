package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IndustryCostIndicesShouldUpsertAndGetCostIndex(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewIndustryCostIndices(db)
	ctx := context.Background()

	indices := []models.IndustryCostIndex{
		{SystemID: 30000142, Activity: "manufacturing", CostIndex: 0.05},
		{SystemID: 30000142, Activity: "researching_time_efficiency", CostIndex: 0.02},
		{SystemID: 30002187, Activity: "manufacturing", CostIndex: 0.10},
	}

	err = repo.UpsertIndices(ctx, indices)
	assert.NoError(t, err)

	// Verify individual lookups
	result, err := repo.GetCostIndex(ctx, 30000142, "manufacturing")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(30000142), result.SystemID)
	assert.Equal(t, "manufacturing", result.Activity)
	assert.InDelta(t, 0.05, result.CostIndex, 0.001)

	result, err = repo.GetCostIndex(ctx, 30002187, "manufacturing")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.InDelta(t, 0.10, result.CostIndex, 0.001)
}

func Test_IndustryCostIndicesShouldReturnNilForMissingIndex(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewIndustryCostIndices(db)
	ctx := context.Background()

	result, err := repo.GetCostIndex(ctx, 99999, "manufacturing")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func Test_IndustryCostIndicesShouldGetLastUpdateTime(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewIndustryCostIndices(db)
	ctx := context.Background()

	// Empty table returns nil
	lastUpdate, err := repo.GetLastUpdateTime(ctx)
	assert.NoError(t, err)
	assert.Nil(t, lastUpdate)

	// After inserting data, returns non-nil
	indices := []models.IndustryCostIndex{
		{SystemID: 30000142, Activity: "manufacturing", CostIndex: 0.05},
	}

	err = repo.UpsertIndices(ctx, indices)
	assert.NoError(t, err)

	lastUpdate, err = repo.GetLastUpdateTime(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, lastUpdate)
}

func Test_IndustryCostIndicesShouldUpdateExistingOnReUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewIndustryCostIndices(db)
	ctx := context.Background()

	// Insert initial data
	indices := []models.IndustryCostIndex{
		{SystemID: 30000142, Activity: "manufacturing", CostIndex: 0.05},
		{SystemID: 30002187, Activity: "manufacturing", CostIndex: 0.10},
	}

	err = repo.UpsertIndices(ctx, indices)
	assert.NoError(t, err)

	// Re-upsert with updated cost_index for one system
	indices = []models.IndustryCostIndex{
		{SystemID: 30000142, Activity: "manufacturing", CostIndex: 0.08},
	}

	err = repo.UpsertIndices(ctx, indices)
	assert.NoError(t, err)

	// Previously inserted entry should still exist (upsert, not truncate)
	result, err := repo.GetCostIndex(ctx, 30002187, "manufacturing")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.InDelta(t, 0.10, result.CostIndex, 0.001)

	// Updated value should reflect new cost_index
	result, err = repo.GetCostIndex(ctx, 30000142, "manufacturing")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.InDelta(t, 0.08, result.CostIndex, 0.001)
}

func Test_IndustryCostIndicesShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewIndustryCostIndices(db)
	ctx := context.Background()

	err = repo.UpsertIndices(ctx, []models.IndustryCostIndex{})
	assert.NoError(t, err)

	err = repo.UpsertIndices(ctx, nil)
	assert.NoError(t, err)
}
