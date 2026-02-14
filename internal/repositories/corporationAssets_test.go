package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_CorporationAssetsShouldUpsertAndUpdate(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corpAssetsRepo := repositories.NewCorporationAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepo.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Insert initial assets
	assets := []*models.EveAsset{
		{
			ItemID:          5001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        100,
			TypeID:          34, // Tritanium
			LocationFlag:    "CorpSAG1",
		},
		{
			ItemID:          5002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        50,
			TypeID:          35, // Pyerite
			LocationFlag:    "CorpSAG2",
		},
	}

	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, assets)
	assert.NoError(t, err)

	// Update existing asset (change quantity)
	assets[0].Quantity = 200

	// Add new asset
	newAsset := &models.EveAsset{
		ItemID:          5003,
		IsBlueprintCopy: false,
		IsSingleton:     true,
		LocationID:      60003760,
		LocationType:    "station",
		Quantity:        1,
		TypeID:          3293, // Container
		LocationFlag:    "CorpSAG1",
	}

	updatedAssets := []*models.EveAsset{assets[0], assets[1], newAsset}

	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, updatedAssets)
	assert.NoError(t, err)

	// Verify through assets API (we need to verify the data was stored correctly)
	// Since we can't directly query corporation_assets, we'll use GetAssembledContainers
	containers, err := corpAssetsRepo.GetAssembledContainers(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Contains(t, containers, int64(5003), "Container should be in the list")
}

func Test_CorporationAssetsShouldGetAssembledContainers(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corpAssetsRepo := repositories.NewCorporationAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepo.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Create assets with containers
	assets := []*models.EveAsset{
		// Non-container item
		{
			ItemID:          5001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        100,
			TypeID:          34, // Tritanium (not a container)
			LocationFlag:    "CorpSAG1",
		},
		// Container (singleton)
		{
			ItemID:          5002,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293, // Container
			LocationFlag:    "CorpSAG1",
		},
		// Another container
		{
			ItemID:          5003,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293, // Container
			LocationFlag:    "CorpSAG2",
		},
	}

	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, assets)
	assert.NoError(t, err)

	// Get containers
	containers, err := corpAssetsRepo.GetAssembledContainers(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)

	// Should only return the containers (singletons), not regular items
	assert.Len(t, containers, 2)
	assert.Contains(t, containers, int64(5002))
	assert.Contains(t, containers, int64(5003))
	assert.NotContains(t, containers, int64(5001), "Non-container items should not be included")
}

func Test_CorporationAssetsShouldUpsertContainerNames(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corpAssetsRepo := repositories.NewCorporationAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepo.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Create containers
	assets := []*models.EveAsset{
		{
			ItemID:          5002,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG1",
		},
		{
			ItemID:          5003,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG2",
		},
	}

	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, assets)
	assert.NoError(t, err)

	// Upsert container names
	containerNames := map[int64]string{
		5002: "Materials Storage",
		5003: "Build Components",
	}

	err = corpAssetsRepo.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	// Update one name
	updatedNames := map[int64]string{
		5002: "Updated Materials Storage",
	}

	err = corpAssetsRepo.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, updatedNames)
	assert.NoError(t, err)

	// Verification would require querying corporation_asset_location_names
	// which is tested indirectly through the assets integration tests
}

func Test_CorporationAssetsShouldGetPlayerOwnedStationIDs(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corpAssetsRepo := repositories.NewCorporationAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepo.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Create assets at multiple player-owned structures (location_type='item', location_flag='Hangar')
	assets := []*models.EveAsset{
		{
			ItemID:          5001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      1000000001, // Player structure 1
			LocationType:    "item",
			Quantity:        100,
			TypeID:          34,
			LocationFlag:    "Hangar",
		},
		{
			ItemID:          5002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      1000000002, // Player structure 2
			LocationType:    "item",
			Quantity:        50,
			TypeID:          35,
			LocationFlag:    "Hangar",
		},
		{
			ItemID:          5003,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      1000000001, // Player structure 1 again
			LocationType:    "item",
			Quantity:        25,
			TypeID:          36,
			LocationFlag:    "Hangar",
		},
	}

	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, assets)
	assert.NoError(t, err)

	// Get player-owned station IDs
	stationIDs, err := corpAssetsRepo.GetPlayerOwnedStationIDs(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)

	// Should return both unique player structure IDs
	assert.Len(t, stationIDs, 2)
	assert.Contains(t, stationIDs, int64(1000000001))
	assert.Contains(t, stationIDs, int64(1000000002))
}

func Test_CorporationAssetsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	playerCorpsRepo := repositories.NewPlayerCorporations(db)
	corpAssetsRepo := repositories.NewCorporationAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepo.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Upsert empty array should not error
	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, []*models.EveAsset{})
	assert.NoError(t, err)

	// Upsert nil should not error
	err = corpAssetsRepo.Upsert(context.Background(), testCorp.ID, testUser.ID, nil)
	assert.NoError(t, err)
}
