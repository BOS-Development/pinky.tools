package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StockpileDeficits_Integration(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)

	ctx := context.Background()

	// Insert test data directly using SQL for simplicity
	// Create user
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (1, 'Test User')`)
	require.NoError(t, err)

	// Create character
	_, err = db.ExecContext(ctx, `
		INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on)
		VALUES (12345, 1, 'Test Character', 'token', 'refresh', NOW())
	`)
	require.NoError(t, err)

	// Create item types
	_, err = db.ExecContext(ctx, `
		INSERT INTO asset_item_types (type_id, type_name, volume)
		VALUES
			(34, 'Tritanium', 0.01),
			(35, 'Pyerite', 0.01),
			(36, 'Mexallon', 0.01)
	`)
	require.NoError(t, err)

	// Create location data
	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO constellations (constellation_id, name, region_id)
		VALUES (20000020, 'Kimotoro', 10000002)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO solar_systems (solar_system_id, name, constellation_id, security)
		VALUES (30000142, 'Jita', 20000020, 1.0)
	`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station)
		VALUES (60003760, 'Jita IV - Moon 4 - Caldari Navy Assembly Plant', 30000142, 1000035, true)
	`)
	require.NoError(t, err)

	// Create market prices
	_, err = db.ExecContext(ctx, `
		INSERT INTO market_prices (type_id, region_id, buy_price, sell_price, daily_volume, updated_at)
		VALUES
			(34, 10000002, 5.45, 5.50, 1000000, NOW()),
			(35, 10000002, 10.20, 10.30, 500000, NOW()),
			(36, 10000002, 15.75, 15.90, 250000, NOW())
	`)
	require.NoError(t, err)

	// Create character assets
	_, err = db.ExecContext(ctx, `
		INSERT INTO character_assets
		(character_id, user_id, item_id, update_key, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag)
		VALUES
			(12345, 1, 1001, 'key1', false, false, 60003760, 'station', 5000, 34, 'Hangar'),
			(12345, 1, 1002, 'key2', false, false, 60003760, 'station', 8000, 35, 'Hangar'),
			(12345, 1, 1003, 'key3', false, false, 60003760, 'station', 12000, 36, 'Hangar')
	`)
	require.NoError(t, err)

	// Create stockpile markers
	stockpileMarkersRepo := repositories.NewStockpileMarkers(db)
	markers := []*models.StockpileMarker{
		{
			UserID:          1,
			TypeID:          34, // Deficit: has 5000, needs 10000 = -5000
			OwnerType:       "character",
			OwnerID:         12345,
			LocationID:      60003760,
			DesiredQuantity: 10000,
		},
		{
			UserID:          1,
			TypeID:          35, // Surplus: has 8000, needs 5000 = +3000 (should NOT appear)
			OwnerType:       "character",
			OwnerID:         12345,
			LocationID:      60003760,
			DesiredQuantity: 5000,
		},
		{
			UserID:          1,
			TypeID:          36, // Deficit: has 12000, needs 15000 = -3000
			OwnerType:       "character",
			OwnerID:         12345,
			LocationID:      60003760,
			DesiredQuantity: 15000,
		},
	}

	for _, marker := range markers {
		err = stockpileMarkersRepo.Upsert(ctx, marker)
		require.NoError(t, err)
	}

	// Execute: Get stockpile deficits
	assetsRepo := repositories.NewAssets(db)
	result, err := assetsRepo.GetStockpileDeficits(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify: Should only return 2 items (Tritanium and Mexallon with deficits)
	assert.Len(t, result.Items, 2, "Should only return items with deficits (stockpile_delta < 0)")

	// Find items by type_id
	itemsByType := make(map[int64]*repositories.StockpileItem)
	for _, item := range result.Items {
		itemsByType[item.TypeID] = item
	}

	// Verify Tritanium deficit
	tritanium := itemsByType[34]
	require.NotNil(t, tritanium, "Tritanium should be in results")
	assert.Equal(t, "Tritanium", tritanium.Name)
	assert.Equal(t, int64(5000), tritanium.Quantity, "Current quantity")
	assert.Equal(t, int64(10000), tritanium.DesiredQuantity, "Desired quantity")
	assert.Equal(t, int64(-5000), tritanium.StockpileDelta, "Deficit amount")
	assert.Equal(t, "Test Character", tritanium.OwnerName)
	assert.Equal(t, "character", tritanium.OwnerType)
	assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", tritanium.StructureName)
	assert.Equal(t, "Jita", tritanium.SolarSystem)
	assert.Equal(t, "The Forge", tritanium.Region)
	assert.Nil(t, tritanium.ContainerName)

	// Verify deficit value: deficit (5000) * buy_price (5.45) = 27,250
	assert.InDelta(t, 27250.0, tritanium.DeficitValue, 0.01, "Deficit ISK value")

	// Verify Mexallon deficit
	mexallon := itemsByType[36]
	require.NotNil(t, mexallon, "Mexallon should be in results")
	assert.Equal(t, "Mexallon", mexallon.Name)
	assert.Equal(t, int64(12000), mexallon.Quantity)
	assert.Equal(t, int64(15000), mexallon.DesiredQuantity)
	assert.Equal(t, int64(-3000), mexallon.StockpileDelta)

	// Verify deficit value: deficit (3000) * buy_price (15.75) = 47,250
	assert.InDelta(t, 47250.0, mexallon.DeficitValue, 0.01, "Deficit ISK value")

	// Verify Pyerite is NOT in results (has surplus, not deficit)
	_, hasPyerite := itemsByType[35]
	assert.False(t, hasPyerite, "Pyerite should NOT be in results (has surplus)")
}

func Test_StockpileDeficits_NoDeficits(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)

	ctx := context.Background()

	// Create user with no assets
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (99, 'Empty User')`)
	require.NoError(t, err)

	// Execute
	assetsRepo := repositories.NewAssets(db)
	result, err := assetsRepo.GetStockpileDeficits(ctx, 99)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify empty result
	assert.Empty(t, result.Items, "Should return empty array when user has no deficits")
}

func Test_StockpileDeficits_UserIsolation(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)

	ctx := context.Background()

	// Create two users
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (10, 'User 1'), (20, 'User 2')`)
	require.NoError(t, err)

	// Create characters
	_, err = db.ExecContext(ctx, `
		INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on)
		VALUES
			(100, 10, 'Char 1', 'token', 'refresh', NOW()),
			(200, 20, 'Char 2', 'token', 'refresh', NOW())
	`)
	require.NoError(t, err)

	// Setup minimal location data
	_, err = db.ExecContext(ctx, `
		INSERT INTO asset_item_types (type_id, type_name, volume) VALUES (34, 'Tritanium', 0.01);
		INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge');
		INSERT INTO constellations (constellation_id, name, region_id) VALUES (20000020, 'Kimotoro', 10000002);
		INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (30000142, 'Jita', 20000020, 1.0);
		INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES (60003760, 'Jita IV', 30000142, 1000035, true);
	`)
	require.NoError(t, err)

	// Create assets for both users
	_, err = db.ExecContext(ctx, `
		INSERT INTO character_assets
		(character_id, user_id, item_id, update_key, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag)
		VALUES
			(100, 10, 2001, 'k1', false, false, 60003760, 'station', 100, 34, 'Hangar'),
			(200, 20, 2002, 'k2', false, false, 60003760, 'station', 200, 34, 'Hangar')
	`)
	require.NoError(t, err)

	// Create stockpile markers with deficits
	stockpileRepo := repositories.NewStockpileMarkers(db)

	marker1 := &models.StockpileMarker{
		UserID:          10,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         100,
		LocationID:      60003760,
		DesiredQuantity: 1000, // Deficit of 900
	}

	marker2 := &models.StockpileMarker{
		UserID:          20,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         200,
		LocationID:      60003760,
		DesiredQuantity: 2000, // Deficit of 1800
	}

	err = stockpileRepo.Upsert(ctx, marker1)
	require.NoError(t, err)
	err = stockpileRepo.Upsert(ctx, marker2)
	require.NoError(t, err)

	// Execute for both users
	assetsRepo := repositories.NewAssets(db)

	result1, err := assetsRepo.GetStockpileDeficits(ctx, 10)
	require.NoError(t, err)

	result2, err := assetsRepo.GetStockpileDeficits(ctx, 20)
	require.NoError(t, err)

	// Verify user 1 sees only their deficit
	assert.Len(t, result1.Items, 1)
	assert.Equal(t, int64(-900), result1.Items[0].StockpileDelta)
	assert.Equal(t, int64(100), result1.Items[0].OwnerID)
	assert.Equal(t, "Char 1", result1.Items[0].OwnerName)

	// Verify user 2 sees only their deficit
	assert.Len(t, result2.Items, 1)
	assert.Equal(t, int64(-1800), result2.Items[0].StockpileDelta)
	assert.Equal(t, int64(200), result2.Items[0].OwnerID)
	assert.Equal(t, "Char 2", result2.Items[0].OwnerName)
}

// Test_StockpileDeficits_ContainerRecursion is a regression test for the bug where
// items in containers inside office folders at stations were not being found.
// This tests the 3-level hierarchy: station -> office -> container -> items
func Test_StockpileDeficits_ContainerRecursion(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)

	ctx := context.Background()

	// Create user
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (1, 'Test User')`)
	require.NoError(t, err)

	// Create corporation
	_, err = db.ExecContext(ctx, `
		INSERT INTO player_corporations (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on)
		VALUES (98000001, 1, 'Test Corp', 'token', 'refresh', NOW())
	`)
	require.NoError(t, err)

	// Create item types
	_, err = db.ExecContext(ctx, `
		INSERT INTO asset_item_types (type_id, type_name, volume)
		VALUES
			(2395, 'Proteins', 0.19),
			(3645, 'Water', 0.19),
			(17368, 'Station Warehouse Container', 1000),
			(27, 'Office', 1)
	`)
	require.NoError(t, err)

	// Create location hierarchy
	_, err = db.ExecContext(ctx, `
		INSERT INTO regions (region_id, name) VALUES (10000002, 'The Forge');
		INSERT INTO constellations (constellation_id, name, region_id)
			VALUES (20000020, 'Kimotoro', 10000002);
		INSERT INTO solar_systems (solar_system_id, name, constellation_id, security)
			VALUES (30000142, 'Jita', 20000020, 1.0);
		INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station)
			VALUES (60003760, 'Jita IV - Moon 4 - Caldari Navy Assembly Plant', 30000142, 1000035, true);
	`)
	require.NoError(t, err)

	// Create market prices
	_, err = db.ExecContext(ctx, `
		INSERT INTO market_prices (type_id, region_id, buy_price, sell_price, updated_at)
		VALUES
			(2395, 10000002, 886.59, 900.00, NOW()),
			(3645, 10000002, 471.82, 500.00, NOW())
	`)
	require.NoError(t, err)

	// Create corporation division names (both hangar and wallet to test filtering)
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_divisions (corporation_id, user_id, division_number, division_type, name)
		VALUES
			(98000001, 1, 3, 'hangar', 'PI Materials'),
			(98000001, 1, 3, 'wallet', 'Should Not Appear')
	`)
	require.NoError(t, err)

	// Create the location hierarchy:
	// 1. Office in station
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(98000001, 1, 1000001, false, true, 60003760, 'item', 1, 27, 'OfficeFolder', NOW())
	`)
	require.NoError(t, err)

	// 2. Container in office (this is the critical part - container is inside office, not directly at station)
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(98000001, 1, 1000002, false, true, 1000001, 'item', 1, 17368, 'CorpSAG3', NOW())
	`)
	require.NoError(t, err)

	// 3. Items in container
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(98000001, 1, 1000003, false, false, 1000002, 'item', 250000, 2395, 'Unlocked', NOW()),
			(98000001, 1, 1000004, false, false, 1000002, 'item', 300000, 3645, 'Unlocked', NOW())
	`)
	require.NoError(t, err)

	// Create stockpile markers with desired quantities higher than current
	stockpileRepo := repositories.NewStockpileMarkers(db)

	containerID := int64(1000002)
	divisionNum := int(3)

	proteinsMarker := &models.StockpileMarker{
		UserID:          1,
		TypeID:          2395, // Proteins
		OwnerType:       "corporation",
		OwnerID:         98000001,
		LocationID:      60003760,
		ContainerID:     &containerID,  // Container ID
		DivisionNumber:  &divisionNum,  // Division 3
		DesiredQuantity: 500000,        // Deficit of 250,000
	}

	waterMarker := &models.StockpileMarker{
		UserID:          1,
		TypeID:          3645, // Water
		OwnerType:       "corporation",
		OwnerID:         98000001,
		LocationID:      60003760,
		ContainerID:     &containerID,
		DivisionNumber:  &divisionNum,
		DesiredQuantity: 500000, // Deficit of 200,000
	}

	err = stockpileRepo.Upsert(ctx, proteinsMarker)
	require.NoError(t, err)
	err = stockpileRepo.Upsert(ctx, waterMarker)
	require.NoError(t, err)

	// Execute GetStockpileDeficits
	assetsRepo := repositories.NewAssets(db)
	result, err := assetsRepo.GetStockpileDeficits(ctx, 1)
	require.NoError(t, err)

	// CRITICAL ASSERTIONS: These verify the bug fix
	// Before the fix, this would return 0 items because the view didn't handle
	// the office -> container -> items hierarchy correctly
	assert.Len(t, result.Items, 2, "Should find both deficit items in container inside office")

	// Verify both items are found with correct deficits
	foundProteins := false
	foundWater := false

	for _, item := range result.Items {
		assert.Equal(t, "Test Corp", item.OwnerName)
		assert.Equal(t, int64(98000001), item.OwnerID)
		assert.Equal(t, "corporation", item.OwnerType)
		assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", item.StructureName)
		assert.Equal(t, "Jita", item.SolarSystem)
		assert.Equal(t, "The Forge", item.Region)
		assert.NotNil(t, item.ContainerName)
		assert.Contains(t, *item.ContainerName, "Station Warehouse Container")

		if item.TypeID == 2395 { // Proteins
			foundProteins = true
			assert.Equal(t, "Proteins", item.Name)
			assert.Equal(t, int64(250000), item.Quantity)
			assert.Equal(t, int64(500000), item.DesiredQuantity)
			assert.Equal(t, int64(-250000), item.StockpileDelta)
			// Deficit value should be 250,000 * 886.59 (buy price)
			assert.InDelta(t, 221647500.0, item.DeficitValue, 1.0)
		}

		if item.TypeID == 3645 { // Water
			foundWater = true
			assert.Equal(t, "Water", item.Name)
			assert.Equal(t, int64(300000), item.Quantity)
			assert.Equal(t, int64(500000), item.DesiredQuantity)
			assert.Equal(t, int64(-200000), item.StockpileDelta)
			// Deficit value should be 200,000 * 471.82 (buy price)
			assert.InDelta(t, 94364000.0, item.DeficitValue, 1.0)
		}
	}

	assert.True(t, foundProteins, "Should find Proteins deficit")
	assert.True(t, foundWater, "Should find Water deficit")
}
