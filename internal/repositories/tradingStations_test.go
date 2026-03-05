package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_TradingStations_ListStations_ReturnsPresetJita(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewTradingStations(db)

	stations, err := repo.ListStations(context.Background())
	assert.NoError(t, err)
	// The migrations seed 5 presets (Jita + 4 secondary hubs)
	assert.GreaterOrEqual(t, len(stations), 5)

	// Find the preset Jita station by station_id
	var foundJita bool
	for _, s := range stations {
		if s.StationID == 60003760 {
			foundJita = true
			assert.True(t, s.IsPreset, "Jita should be a preset station")
			assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", s.Name)
			assert.Equal(t, int64(30000142), s.SystemID)
			assert.Equal(t, int64(10000002), s.RegionID)
			break
		}
	}
	assert.True(t, foundJita, "expected preset station for Jita 4-4 (station_id 60003760)")
}

func Test_TradingStations_ListStations_ReturnsAllFivePresets(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewTradingStations(db)

	stations, err := repo.ListStations(context.Background())
	assert.NoError(t, err)

	// Collect all preset station_ids
	presetIDs := map[int64]bool{}
	for _, s := range stations {
		if s.IsPreset {
			presetIDs[s.StationID] = true
		}
	}

	expectedPresets := []int64{
		60003760, // Jita
		60008494, // Amarr
		60011866, // Dodixie
		60004588, // Rens
		60005686, // Hek
	}
	for _, id := range expectedPresets {
		assert.True(t, presetIDs[id], "expected preset station_id %d to be in list", id)
	}
}

func Test_TradingStations_ListStations_OrderedPresetFirst(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewTradingStations(db)

	stations, err := repo.ListStations(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(stations), 1)

	// Preset stations should appear before non-preset
	sawNonPreset := false
	for _, s := range stations {
		if !s.IsPreset {
			sawNonPreset = true
		}
		if sawNonPreset && s.IsPreset {
			t.Error("preset station appeared after non-preset station")
		}
	}
}

func Test_TradingStations_ListStations_ReturnsEmptySlice(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewTradingStations(db)

	stations, err := repo.ListStations(context.Background())
	assert.NoError(t, err)
	// Should return a non-nil slice (initialized as []*TradingStation{})
	assert.NotNil(t, stations)
}
