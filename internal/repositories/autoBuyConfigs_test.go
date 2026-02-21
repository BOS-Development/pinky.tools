package repositories_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func setupAutoBuyTestData(t *testing.T, db *sql.DB, userID, charID int64) {
	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)

	user := &repositories.User{ID: userID, Name: "Test User"}
	err := userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: charID, Name: "Test Character", UserID: userID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)
}

func int64Ptr(v int64) *int64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func Test_AutoBuyConfigsShouldUpsertAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7000, 70001)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:             7000,
		OwnerType:          "character",
		OwnerID:            70001,
		LocationID:         60003760,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)
	assert.NotZero(t, config.ID)
	assert.True(t, config.IsActive)
	assert.False(t, config.CreatedAt.IsZero())
	assert.False(t, config.UpdatedAt.IsZero())

	// GetByUser should return the config
	configs, err := repo.GetByUser(context.Background(), 7000)
	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, config.ID, configs[0].ID)
	assert.Equal(t, "jita_sell", configs[0].PriceSource)
	assert.Equal(t, 0.9, configs[0].MinPricePercentage)
	assert.Equal(t, 1.1, configs[0].MaxPricePercentage)
}

func Test_AutoBuyConfigsShouldUpsertUpdate(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7010, 70011)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:             7010,
		OwnerType:          "character",
		OwnerID:            70011,
		LocationID:         60003760,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)
	originalID := config.ID
	originalCreatedAt := config.CreatedAt

	// Upsert again with different price settings
	config2 := &models.AutoBuyConfig{
		UserID:             7010,
		OwnerType:          "character",
		OwnerID:            70011,
		LocationID:         60003760,
		MinPricePercentage: 0.8,
		MaxPricePercentage: 1.2,
		PriceSource:        "jita_buy",
	}

	err = repo.Upsert(context.Background(), config2)
	assert.NoError(t, err)
	assert.Equal(t, originalID, config2.ID, "should update same record")
	assert.Equal(t, originalCreatedAt.Unix(), config2.CreatedAt.Unix(), "created_at should not change")

	// Verify only one config exists
	configs, err := repo.GetByUser(context.Background(), 7010)
	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, 0.8, configs[0].MinPricePercentage)
	assert.Equal(t, "jita_buy", configs[0].PriceSource)
}

func Test_AutoBuyConfigsShouldGetByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7020, 70021)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:             7020,
		OwnerType:          "character",
		OwnerID:            70021,
		LocationID:         60003760,
		ContainerID:        int64Ptr(99999),
		MinPricePercentage: 0.95,
		MaxPricePercentage: 1.05,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// GetByID should return the config
	result, err := repo.GetByID(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, config.ID, result.ID)
	assert.Equal(t, int64Ptr(99999), result.ContainerID)

	// Non-existent ID returns nil without error
	result, err = repo.GetByID(context.Background(), 999999)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func Test_AutoBuyConfigsShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7030, 70031)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:             7030,
		OwnerType:          "character",
		OwnerID:            70031,
		LocationID:         60003760,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// Delete should soft-delete
	err = repo.Delete(context.Background(), config.ID, 7030)
	assert.NoError(t, err)

	// GetByUser should return empty (soft-deleted)
	configs, err := repo.GetByUser(context.Background(), 7030)
	assert.NoError(t, err)
	assert.Len(t, configs, 0)

	// GetByID still returns it (no is_active filter)
	result, err := repo.GetByID(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsActive)

	// Double-delete returns error
	err = repo.Delete(context.Background(), config.ID, 7030)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_AutoBuyConfigsDeleteWrongUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7040, 70041)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:             7040,
		OwnerType:          "character",
		OwnerID:            70041,
		LocationID:         60003760,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// Delete with wrong user should fail
	err = repo.Delete(context.Background(), config.ID, 9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_AutoBuyConfigsShouldGetAllActive(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7050, 70051)

	// Create a second user
	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	user2 := &repositories.User{ID: 7051, Name: "User 2"}
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	char2 := &repositories.Character{ID: 70052, Name: "Char 2", UserID: 7051}
	err = charRepo.Add(context.Background(), char2)
	assert.NoError(t, err)

	repo := repositories.NewAutoBuyConfigs(db)

	config1 := &models.AutoBuyConfig{
		UserID:             7050,
		OwnerType:          "character",
		OwnerID:            70051,
		LocationID:         60003760,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	config2 := &models.AutoBuyConfig{
		UserID:             7051,
		OwnerType:          "character",
		OwnerID:            70052,
		LocationID:         60003760,
		MinPricePercentage: 0.85,
		MaxPricePercentage: 1.15,
		PriceSource:        "jita_buy",
	}

	err = repo.Upsert(context.Background(), config1)
	assert.NoError(t, err)
	err = repo.Upsert(context.Background(), config2)
	assert.NoError(t, err)

	// Soft-delete one
	err = repo.Delete(context.Background(), config1.ID, 7050)
	assert.NoError(t, err)

	// GetAllActive should only return the active one
	active, err := repo.GetAllActive(context.Background())
	assert.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, config2.ID, active[0].ID)
}

func Test_AutoBuyConfigsDeficitsCharacterHangar(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7060, 70061)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	charAssetsRepo := repositories.NewCharacterAssets(db)
	repo := repositories.NewAutoBuyConfigs(db)

	// Create stockpile marker: want 10000 Tritanium
	marker := &models.StockpileMarker{
		UserID:          7060,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         70061,
		LocationID:      60003760,
		DesiredQuantity: 10000,
		PriceSource:     stringPtr("jita_sell"),
		PricePercentage: float64Ptr(1.0),
	}
	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Add 3000 Tritanium in character hangar
	assets := []*models.EveAsset{
		{
			ItemID:       900001,
			LocationID:   60003760,
			LocationType: "other",
			Quantity:     3000,
			TypeID:       34,
			LocationFlag: "Hangar",
		},
	}
	err = charAssetsRepo.UpdateAssets(context.Background(), 70061, 7060, assets)
	assert.NoError(t, err)

	// Query deficits
	config := &models.AutoBuyConfig{
		UserID:     7060,
		OwnerType:  "character",
		OwnerID:    70061,
		LocationID: 60003760,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(context.Background(), config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 1)
	assert.Equal(t, int64(34), deficits[0].TypeID)
	assert.Equal(t, int64(10000), deficits[0].DesiredQuantity)
	assert.Equal(t, int64(3000), deficits[0].CurrentQuantity)
	assert.Equal(t, int64(7000), deficits[0].Deficit)
	assert.Equal(t, stringPtr("jita_sell"), deficits[0].PriceSource)
	assert.Equal(t, float64Ptr(1.0), deficits[0].PricePercentage)
}

func Test_AutoBuyConfigsDeficitsCharacterContainer(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7070, 70071)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	charAssetsRepo := repositories.NewCharacterAssets(db)
	repo := repositories.NewAutoBuyConfigs(db)

	containerID := int64(800001)

	// Create stockpile marker for items in a container
	marker := &models.StockpileMarker{
		UserID:          7070,
		TypeID:          35, // Pyerite
		OwnerType:       "character",
		OwnerID:         70071,
		LocationID:      60003760,
		ContainerID:     &containerID,
		DesiredQuantity: 5000,
	}
	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Add assets in a container (location_type = 'item', location_id = container item_id)
	assets := []*models.EveAsset{
		{
			ItemID:       900010,
			LocationID:   containerID,
			LocationType: "item",
			Quantity:     2000,
			TypeID:       35,
			LocationFlag: "Hangar",
		},
	}
	err = charAssetsRepo.UpdateAssets(context.Background(), 70071, 7070, assets)
	assert.NoError(t, err)

	config := &models.AutoBuyConfig{
		UserID:      7070,
		OwnerType:   "character",
		OwnerID:     70071,
		LocationID:  60003760,
		ContainerID: &containerID,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(context.Background(), config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 1)
	assert.Equal(t, int64(35), deficits[0].TypeID)
	assert.Equal(t, int64(5000), deficits[0].DesiredQuantity)
	assert.Equal(t, int64(2000), deficits[0].CurrentQuantity)
	assert.Equal(t, int64(3000), deficits[0].Deficit)
}

func Test_AutoBuyConfigsDeficitsCorpDivision(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7080, 70081)

	ctx := context.Background()

	// Set up corporation
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corp := repositories.PlayerCorporation{
		ID:              70082,
		UserID:          7080,
		Name:            "Test Corp",
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}
	err = playerCorpsRepo.Upsert(ctx, corp)
	assert.NoError(t, err)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	repo := repositories.NewAutoBuyConfigs(db)

	divNum := 3

	// Create stockpile marker for corp division
	marker := &models.StockpileMarker{
		UserID:          7080,
		TypeID:          34,
		OwnerType:       "corporation",
		OwnerID:         70082,
		LocationID:      60003760,
		DivisionNumber:  &divNum,
		DesiredQuantity: 20000,
	}
	err = stockpileRepo.Upsert(ctx, marker)
	assert.NoError(t, err)

	// Set up corp assets: office folder at station, then items in CorpSAG3
	// 1. Office at station
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(70082, 7080, 1000001, false, true, 60003760, 'item', 1, 27, 'OfficeFolder', NOW())
	`)
	assert.NoError(t, err)

	// 2. Tritanium in CorpSAG3 (inside the office)
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(70082, 7080, 1000002, false, false, 1000001, 'item', 8000, 34, 'CorpSAG3', NOW())
	`)
	assert.NoError(t, err)

	config := &models.AutoBuyConfig{
		UserID:         7080,
		OwnerType:      "corporation",
		OwnerID:        70082,
		LocationID:     60003760,
		DivisionNumber: &divNum,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(ctx, config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 1)
	assert.Equal(t, int64(34), deficits[0].TypeID)
	assert.Equal(t, int64(20000), deficits[0].DesiredQuantity)
	assert.Equal(t, int64(8000), deficits[0].CurrentQuantity)
	assert.Equal(t, int64(12000), deficits[0].Deficit)
}

func Test_AutoBuyConfigsDeficitsNoStockpileMarkers(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7090, 70091)

	repo := repositories.NewAutoBuyConfigs(db)

	config := &models.AutoBuyConfig{
		UserID:     7090,
		OwnerType:  "character",
		OwnerID:    70091,
		LocationID: 60003760,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(context.Background(), config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 0)
}

func Test_AutoBuyConfigsDeficitsCorpContainer(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7100, 71001)

	ctx := context.Background()

	// Set up corporation
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corp := repositories.PlayerCorporation{
		ID:              71002,
		UserID:          7100,
		Name:            "Test Corp",
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}
	err = playerCorpsRepo.Upsert(ctx, corp)
	assert.NoError(t, err)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	repo := repositories.NewAutoBuyConfigs(db)

	containerID := int64(2000001)

	// Stockpile marker for corp container
	marker := &models.StockpileMarker{
		UserID:          7100,
		TypeID:          36, // Mexallon
		OwnerType:       "corporation",
		OwnerID:         71002,
		LocationID:      60003760,
		ContainerID:     &containerID,
		DesiredQuantity: 15000,
	}
	err = stockpileRepo.Upsert(ctx, marker)
	assert.NoError(t, err)

	// Corp assets in a container
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(71002, 7100, 2000002, false, false, 2000001, 'item', 6000, 36, 'CorpSAG1', NOW())
	`)
	assert.NoError(t, err)

	config := &models.AutoBuyConfig{
		UserID:      7100,
		OwnerType:   "corporation",
		OwnerID:     71002,
		LocationID:  60003760,
		ContainerID: &containerID,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(ctx, config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 1)
	assert.Equal(t, int64(36), deficits[0].TypeID)
	assert.Equal(t, int64(15000), deficits[0].DesiredQuantity)
	assert.Equal(t, int64(6000), deficits[0].CurrentQuantity)
	assert.Equal(t, int64(9000), deficits[0].Deficit)
}

func Test_AutoBuyConfigsWithDivisionNumber(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7110, 71101)

	repo := repositories.NewAutoBuyConfigs(db)

	divNum := 5
	config := &models.AutoBuyConfig{
		UserID:             7110,
		OwnerType:          "character",
		OwnerID:            71101,
		LocationID:         60003760,
		DivisionNumber:     &divNum,
		MinPricePercentage: 0.9,
		MaxPricePercentage: 1.1,
		PriceSource:        "jita_sell",
	}

	err = repo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	result, err := repo.GetByID(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, intPtr(5), result.DivisionNumber)
}

func Test_AutoBuyConfigsDeficitsCorpDivisionIsolation(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7120, 71201)

	ctx := context.Background()

	// Set up corporation
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corp := repositories.PlayerCorporation{
		ID:              71202,
		UserID:          7120,
		Name:            "Test Corp Isolation",
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}
	err = playerCorpsRepo.Upsert(ctx, corp)
	assert.NoError(t, err)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	repo := repositories.NewAutoBuyConfigs(db)

	divNum1 := 1
	divNum3 := 3

	// Stockpile marker for division 1: want 10000 Tritanium
	marker1 := &models.StockpileMarker{
		UserID:          7120,
		TypeID:          34,
		OwnerType:       "corporation",
		OwnerID:         71202,
		LocationID:      60003760,
		DivisionNumber:  &divNum1,
		DesiredQuantity: 10000,
	}
	err = stockpileRepo.Upsert(ctx, marker1)
	assert.NoError(t, err)

	// Stockpile marker for division 3: want 5000 Tritanium
	marker3 := &models.StockpileMarker{
		UserID:          7120,
		TypeID:          34,
		OwnerType:       "corporation",
		OwnerID:         71202,
		LocationID:      60003760,
		DivisionNumber:  &divNum3,
		DesiredQuantity: 5000,
	}
	err = stockpileRepo.Upsert(ctx, marker3)
	assert.NoError(t, err)

	// Office at station
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(71202, 7120, 3000001, false, true, 60003760, 'item', 1, 27, 'OfficeFolder', NOW())
	`)
	assert.NoError(t, err)

	// Tritanium only in CorpSAG1 (division 1) — division 3 has nothing
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(71202, 7120, 3000002, false, false, 3000001, 'item', 4000, 34, 'CorpSAG1', NOW())
	`)
	assert.NoError(t, err)

	// Query division 1: should see 4000 current, 6000 deficit
	config1 := &models.AutoBuyConfig{
		UserID:         7120,
		OwnerType:      "corporation",
		OwnerID:        71202,
		LocationID:     60003760,
		DivisionNumber: &divNum1,
	}
	deficits1, err := repo.GetStockpileDeficitsForConfig(ctx, config1)
	assert.NoError(t, err)
	assert.Len(t, deficits1, 1)
	assert.Equal(t, int64(34), deficits1[0].TypeID)
	assert.Equal(t, int64(4000), deficits1[0].CurrentQuantity)
	assert.Equal(t, int64(6000), deficits1[0].Deficit)

	// Query division 3: should see 0 current, 5000 deficit (no assets in CorpSAG3)
	config3 := &models.AutoBuyConfig{
		UserID:         7120,
		OwnerType:      "corporation",
		OwnerID:        71202,
		LocationID:     60003760,
		DivisionNumber: &divNum3,
	}
	deficits3, err := repo.GetStockpileDeficitsForConfig(ctx, config3)
	assert.NoError(t, err)
	assert.Len(t, deficits3, 1)
	assert.Equal(t, int64(34), deficits3[0].TypeID)
	assert.Equal(t, int64(0), deficits3[0].CurrentQuantity)
	assert.Equal(t, int64(5000), deficits3[0].Deficit)
}

func Test_AutoBuyConfigsDeficitsCorpDivisionMultipleItems(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)
	setupAutoBuyTestData(t, db, 7130, 71301)

	ctx := context.Background()

	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corp := repositories.PlayerCorporation{
		ID:              71302,
		UserID:          7130,
		Name:            "Test Corp Multi",
		EsiToken:        "token",
		EsiRefreshToken: "refresh",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}
	err = playerCorpsRepo.Upsert(ctx, corp)
	assert.NoError(t, err)

	stockpileRepo := repositories.NewStockpileMarkers(db)
	repo := repositories.NewAutoBuyConfigs(db)

	divNum := 2

	// Want 10000 Tritanium and 5000 Pyerite in division 2
	for _, m := range []*models.StockpileMarker{
		{UserID: 7130, TypeID: 34, OwnerType: "corporation", OwnerID: 71302, LocationID: 60003760, DivisionNumber: &divNum, DesiredQuantity: 10000},
		{UserID: 7130, TypeID: 35, OwnerType: "corporation", OwnerID: 71302, LocationID: 60003760, DivisionNumber: &divNum, DesiredQuantity: 5000},
	} {
		err = stockpileRepo.Upsert(ctx, m)
		assert.NoError(t, err)
	}

	// Office + assets in CorpSAG2
	_, err = db.ExecContext(ctx, `
		INSERT INTO corporation_assets
		(corporation_id, user_id, item_id, is_blueprint_copy, is_singleton,
		 location_id, location_type, quantity, type_id, location_flag, update_key)
		VALUES
			(71302, 7130, 4000001, false, true, 60003760, 'item', 1, 27, 'OfficeFolder', NOW()),
			(71302, 7130, 4000002, false, false, 4000001, 'item', 3000, 34, 'CorpSAG2', NOW()),
			(71302, 7130, 4000003, false, false, 4000001, 'item', 5000, 35, 'CorpSAG2', NOW())
	`)
	assert.NoError(t, err)

	config := &models.AutoBuyConfig{
		UserID:         7130,
		OwnerType:      "corporation",
		OwnerID:        71302,
		LocationID:     60003760,
		DivisionNumber: &divNum,
	}

	deficits, err := repo.GetStockpileDeficitsForConfig(ctx, config)
	assert.NoError(t, err)
	assert.Len(t, deficits, 2)

	// Build a map for order-independent assertions
	deficitMap := map[int64]*models.StockpileDeficitItem{}
	for _, d := range deficits {
		deficitMap[d.TypeID] = d
	}

	// Tritanium: 3000 of 10000
	assert.Equal(t, int64(10000), deficitMap[34].DesiredQuantity)
	assert.Equal(t, int64(3000), deficitMap[34].CurrentQuantity)
	assert.Equal(t, int64(7000), deficitMap[34].Deficit)

	// Pyerite: 5000 of 5000 (fully stocked, deficit 0)
	assert.Equal(t, int64(5000), deficitMap[35].DesiredQuantity)
	assert.Equal(t, int64(5000), deficitMap[35].CurrentQuantity)
	assert.Equal(t, int64(0), deficitMap[35].Deficit)
}
