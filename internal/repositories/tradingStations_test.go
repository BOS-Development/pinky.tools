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
	// The migration seeds Jita 4-4 as a preset
	assert.GreaterOrEqual(t, len(stations), 1)

	// Find the preset Jita station
	var jita interface{ GetIsPreset() bool }
	_ = jita
	var foundPreset bool
	for _, s := range stations {
		if s.IsPreset {
			foundPreset = true
			assert.Equal(t, int64(60003760), s.StationID)
			assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", s.Name)
			assert.Equal(t, int64(30000142), s.SystemID)
			assert.Equal(t, int64(10000002), s.RegionID)
			break
		}
	}
	assert.True(t, foundPreset, "expected at least one preset station (Jita 4-4)")
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
