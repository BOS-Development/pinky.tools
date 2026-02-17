package updaters_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// Mock ESI client for cost indices

type mockCostIndicesEsiClient struct {
	systems []*client.IndustryCostIndexSystem
	err     error
}

func (m *mockCostIndicesEsiClient) GetIndustryCostIndices(ctx context.Context) ([]*client.IndustryCostIndexSystem, error) {
	return m.systems, m.err
}

// Mock cost indices repository

type mockCostIndicesRepo struct {
	lastUpdateTime *time.Time
	lastUpdateErr  error
	upsertErr      error
	upsertedRows   []models.IndustryCostIndex
}

func (m *mockCostIndicesRepo) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	return m.lastUpdateTime, m.lastUpdateErr
}

func (m *mockCostIndicesRepo) UpsertIndices(ctx context.Context, indices []models.IndustryCostIndex) error {
	m.upsertedRows = indices
	return m.upsertErr
}

func Test_IndustryCostIndices_FullUpdate(t *testing.T) {
	esiClient := &mockCostIndicesEsiClient{
		systems: []*client.IndustryCostIndexSystem{
			{
				SolarSystemID: 30000142,
				CostIndices: []client.IndustryCostIndexActivity{
					{Activity: "manufacturing", CostIndex: 0.05},
					{Activity: "researching_time_efficiency", CostIndex: 0.02},
				},
			},
			{
				SolarSystemID: 30000144,
				CostIndices: []client.IndustryCostIndexActivity{
					{Activity: "manufacturing", CostIndex: 0.03},
				},
			},
		},
	}
	repo := &mockCostIndicesRepo{}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedRows, 3)

	// Verify the flattened rows contain the right data
	found := map[string]bool{}
	for _, row := range repo.upsertedRows {
		key := fmt.Sprintf("%d-%s", row.SystemID, row.Activity)
		found[key] = true
	}
	assert.True(t, found["30000142-manufacturing"])
	assert.True(t, found["30000142-researching_time_efficiency"])
	assert.True(t, found["30000144-manufacturing"])
}

func Test_IndustryCostIndices_SkipsRecentUpdate(t *testing.T) {
	recentTime := time.Now().Add(-10 * time.Minute)
	repo := &mockCostIndicesRepo{
		lastUpdateTime: &recentTime,
	}
	esiClient := &mockCostIndicesEsiClient{}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Nil(t, repo.upsertedRows)
}

func Test_IndustryCostIndices_UpdatesWhenStale(t *testing.T) {
	staleTime := time.Now().Add(-2 * time.Hour)
	esiClient := &mockCostIndicesEsiClient{
		systems: []*client.IndustryCostIndexSystem{
			{
				SolarSystemID: 30000142,
				CostIndices: []client.IndustryCostIndexActivity{
					{Activity: "manufacturing", CostIndex: 0.05},
				},
			},
		},
	}
	repo := &mockCostIndicesRepo{
		lastUpdateTime: &staleTime,
	}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedRows, 1)
}

func Test_IndustryCostIndices_ErrorGettingLastUpdate(t *testing.T) {
	repo := &mockCostIndicesRepo{
		lastUpdateErr: fmt.Errorf("db error"),
	}
	esiClient := &mockCostIndicesEsiClient{}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get last cost indices update time")
}

func Test_IndustryCostIndices_ErrorFetchingIndices(t *testing.T) {
	esiClient := &mockCostIndicesEsiClient{
		err: fmt.Errorf("network error"),
	}
	repo := &mockCostIndicesRepo{}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch industry cost indices")
}

func Test_IndustryCostIndices_ErrorUpserting(t *testing.T) {
	esiClient := &mockCostIndicesEsiClient{
		systems: []*client.IndustryCostIndexSystem{
			{
				SolarSystemID: 30000142,
				CostIndices: []client.IndustryCostIndexActivity{
					{Activity: "manufacturing", CostIndex: 0.05},
				},
			},
		},
	}
	repo := &mockCostIndicesRepo{
		upsertErr: fmt.Errorf("db error"),
	}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert cost indices")
}

func Test_IndustryCostIndices_FirstRunWithNoLastUpdate(t *testing.T) {
	esiClient := &mockCostIndicesEsiClient{
		systems: []*client.IndustryCostIndexSystem{
			{
				SolarSystemID: 30000142,
				CostIndices: []client.IndustryCostIndexActivity{
					{Activity: "manufacturing", CostIndex: 0.05},
				},
			},
		},
	}
	repo := &mockCostIndicesRepo{
		lastUpdateTime: nil,
	}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Len(t, repo.upsertedRows, 1)
}

func Test_IndustryCostIndices_EmptySystemList(t *testing.T) {
	esiClient := &mockCostIndicesEsiClient{
		systems: []*client.IndustryCostIndexSystem{},
	}
	repo := &mockCostIndicesRepo{}

	u := updaters.NewIndustryCostIndices(esiClient, repo)
	err := u.Update(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, repo.upsertedRows)
}
