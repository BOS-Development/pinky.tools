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

func setupTestUniverse(t *testing.T, db any) {
	regionsRepo := repositories.NewRegions(db.(*sql.DB))
	constellationsRepo := repositories.NewConstellations(db.(*sql.DB))
	solarSystemsRepo := repositories.NewSolarSystems(db.(*sql.DB))
	stationsRepo := repositories.NewStations(db.(*sql.DB))
	itemTypeRepo := repositories.NewItemTypeRepository(db.(*sql.DB))

	regions := []models.Region{
		{ID: 10000002, Name: "The Forge"},
	}
	err := regionsRepo.Upsert(context.Background(), regions)
	assert.NoError(t, err)

	constellations := []models.Constellation{
		{ID: 20000020, Name: "Kimotoro", RegionID: 10000002},
	}
	err = constellationsRepo.Upsert(context.Background(), constellations)
	assert.NoError(t, err)

	solarSystems := []models.SolarSystem{
		{ID: 30000142, Name: "Jita", ConstellationID: 20000020, Security: 0.9},
	}
	err = solarSystemsRepo.Upsert(context.Background(), solarSystems)
	assert.NoError(t, err)

	stations := []models.Station{
		{
			ID:            60003760,
			Name:          "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
			SolarSystemID: 30000142,
			CorporationID: 1000035,
			IsNPC:         true,
		},
		{
			ID:            60003761,
			Name:          "Jita IV - Moon 4 - Some Other Station",
			SolarSystemID: 30000142,
			CorporationID: 1000035,
			IsNPC:         false,
		},
	}
	err = stationsRepo.Upsert(context.Background(), stations)
	assert.NoError(t, err)

	itemTypes := []models.EveInventoryType{
		{TypeID: 27, TypeName: "Office", Volume: 0.0, IconID: nil},
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
		{TypeID: 36, TypeName: "Mexallon", Volume: 0.01, IconID: nil},
		{TypeID: 3293, TypeName: "Medium Standard Container", Volume: 33.0, IconID: nil},
	}
	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)
}

func Test_AssetsShouldGetCharacterAssetsInStation(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)
	characterAssetsRepository := repositories.NewCharacterAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Crushim deez nuts",
		UserID: 42,
	}

	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	characterAssets := []*models.EveAsset{
		{
			ItemID:          1001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        100,
			TypeID:          34,
			LocationFlag:    "Hangar",
		},
	}

	err = characterAssetsRepository.UpdateAssets(context.Background(), testCharacter.ID, testUser.ID, characterAssets)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)
	assert.Equal(t, int64(60003760), response.Structures[0].ID)
	assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", response.Structures[0].Name)
	assert.Equal(t, "Jita", response.Structures[0].SolarSystem)
	assert.Equal(t, "The Forge", response.Structures[0].Region)

	expectedQuantity := int64(100)
	expectedAssets := []*repositories.Asset{
		{
			Name:            "Tritanium",
			TypeID:          34,
			Quantity:        100,
			Volume:          1.0,
			OwnerType:       "character",
			OwnerName:       testCharacter.Name,
			OwnerID:         testCharacter.ID,
			DesiredQuantity: nil,
			StockpileDelta:  &expectedQuantity,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
		},
	}

	assert.Equal(t, expectedAssets, response.Structures[0].HangarAssets)
}

func Test_AssetsShouldGetCharacterAssetsInContainers(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)
	characterAssetsRepository := repositories.NewCharacterAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Crushim deez nuts",
		UserID: 42,
	}

	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	characterAssets := []*models.EveAsset{
		{
			ItemID:          2001,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "Hangar",
		},
		{
			ItemID:          3001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      2001,
			LocationType:    "item",
			Quantity:        50,
			TypeID:          34,
			LocationFlag:    "Hangar",
		},
	}

	err = characterAssetsRepository.UpdateAssets(context.Background(), testCharacter.ID, testUser.ID, characterAssets)
	assert.NoError(t, err)

	containerNames := map[int64]string{
		2001: "My Container",
	}

	err = characterAssetsRepository.UpsertContainerNames(context.Background(), testCharacter.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)

	expectedQuantity50 := int64(50)
	expectedContainers := []*repositories.AssetContainer{
		{
			ID:        2001,
			Name:      "My Container",
			OwnerType: "character",
			OwnerName: testCharacter.Name,
			OwnerID:   testCharacter.ID,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        50,
					Volume:          0.5,
					OwnerType:       "character",
					OwnerName:       testCharacter.Name,
					OwnerID:         testCharacter.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity50,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
		},
	}

	assert.Equal(t, expectedContainers, response.Structures[0].HangarContainers)
}

func Test_AssetsShouldGetCorporationAssetsInDivisions(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Main Hangar",
			2: "Secondary Hangar",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	corpAssets := []*models.EveAsset{
		{
			ItemID:          5000,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          34,
			LocationFlag:    "OfficeFolder",
		},
		{
			ItemID:          6000,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003761,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          34,
			LocationFlag:    "OfficeFolder",
		},
		{
			ItemID:          5001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        200,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
		{
			ItemID:          5002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        150,
			TypeID:          34,
			LocationFlag:    "CorpSAG2",
		},
		// Add container in CorpSAG1
		{
			ItemID:          5003,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293, // Small Secure Container
			LocationFlag:    "CorpSAG1",
		},
		// Add items inside the container
		{
			ItemID:          5004,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      5003, // Inside container 5003
			LocationType:    "item",
			Quantity:        50,
			TypeID:          35, // Pyerite
			LocationFlag:    "Unlocked",
		},
		{
			ItemID:          5005,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      5003, // Inside container 5003
			LocationType:    "item",
			Quantity:        25,
			TypeID:          36, // Mexallon
			LocationFlag:    "Unlocked",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	// Add container names
	containerNames := map[int64]string{
		5003: "Materials Container",
	}
	err = corpAssetsRepository.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// We should have 2 structures now (2 stations with assets)
	assert.Len(t, response.Structures, 2)

	// Find the main station
	var mainStation *repositories.AssetStructure
	for _, structure := range response.Structures {
		if structure.Name == "Jita IV - Moon 4 - Caldari Navy Assembly Plant" {
			mainStation = structure
			break
		}
	}
	assert.NotNil(t, mainStation, "Main station should exist")

	expectedQuantity200 := int64(200)
	expectedQuantity150 := int64(150)
	expectedQuantity50 := int64(50)
	expectedQuantity25 := int64(25)
	expectedHangers := []*repositories.CorporationHanger{
		{
			ID:              1,
			Name:            "Main Hangar",
			CorporationID:   testCorp.ID,
			CorporationName: testCorp.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        200,
					Volume:          2.0,
					OwnerType:       "corporation",
					OwnerName:       testCorp.Name,
					OwnerID:         testCorp.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity200,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{
				{
					ID:        5003,
					Name:      "Materials Container",
					OwnerType: "corporation",
					OwnerName: testCorp.Name,
					OwnerID:   testCorp.ID,
					Assets: []*repositories.Asset{
						{
							Name:            "Pyerite",
							TypeID:          35,
							Quantity:        50,
							Volume:          0.16,
							OwnerType:       "corporation",
							OwnerName:       testCorp.Name,
							OwnerID:         testCorp.ID,
							DesiredQuantity: nil,
							StockpileDelta:  &expectedQuantity50,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
						},
						{
							Name:            "Mexallon",
							TypeID:          36,
							Quantity:        25,
							Volume:          0.25,
							OwnerType:       "corporation",
							OwnerName:       testCorp.Name,
							OwnerID:         testCorp.ID,
							DesiredQuantity: nil,
							StockpileDelta:  &expectedQuantity25,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
						},
					},
				},
			},
		},
		{
			ID:              2,
			Name:            "Secondary Hangar",
			CorporationID:   testCorp.ID,
			CorporationName: testCorp.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        150,
					Volume:          1.5,
					OwnerType:       "corporation",
					OwnerName:       testCorp.Name,
					OwnerID:         testCorp.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity150,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
	}

	assert.ElementsMatch(t, expectedHangers, mainStation.CorporationHangers)

	// CRITICAL: Check that the OTHER station doesn't have the same divisions duplicated
	var otherStation *repositories.AssetStructure
	for _, structure := range response.Structures {
		if structure.Name != "Jita IV - Moon 4 - Caldari Navy Assembly Plant" {
			otherStation = structure
			break
		}
	}

	if otherStation != nil {
		// The other station has corp assets (OfficeFolder), so all corp divisions should appear there
		// But the divisions should be EMPTY since no CorpSAG1/2 assets are there
		t.Logf("Other station '%s' has %d corp hangers (expected: 2)", otherStation.Name, len(otherStation.CorporationHangers))
		assert.Len(t, otherStation.CorporationHangers, 2, "Other station should show all divisions (corp has presence there)")

		// All divisions at the other station should be empty
		for _, hanger := range otherStation.CorporationHangers {
			assert.Len(t, hanger.Assets, 0, "Division '%s' at other station should have no assets", hanger.Name)
			assert.Len(t, hanger.HangarContainers, 0, "Division '%s' at other station should have no containers", hanger.Name)
		}
	}
}

func Test_AssetsShouldHandleNestedContainersInOffices(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Mining Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	// Set up corp divisions
	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Hangar 1",
			2: "Hangar 2",
			3: "PI Materials",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	// Create the scenario that caused the bug:
	// 1. Office at station
	// 2. Container nested INSIDE the Office (location_type='item', location_id=office item_id)
	// 3. Items inside that nested container
	corpAssets := []*models.EveAsset{
		// Office at station
		{
			ItemID:          7001,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          27, // Office
			LocationFlag:    "OfficeFolder",
		},
		// Container INSIDE the Office (nested) - this is the key test case
		{
			ItemID:          7002,
			IsSingleton:     true,
			LocationID:      7001, // Inside Office
			LocationType:    "item",
			Quantity:        1,
			TypeID:          3293, // Medium Standard Container
			LocationFlag:    "CorpSAG3", // PI Materials division
		},
		// Items inside the nested container
		{
			ItemID:          7003,
			IsSingleton:     false,
			LocationID:      7002, // Inside container 7002
			LocationType:    "item",
			Quantity:        100,
			TypeID:          34, // Tritanium
			LocationFlag:    "Unlocked",
		},
		{
			ItemID:          7004,
			IsSingleton:     false,
			LocationID:      7002, // Inside container 7002
			LocationType:    "item",
			Quantity:        50,
			TypeID:          35, // Pyerite
			LocationFlag:    "Unlocked",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	// Add container names
	containerNames := map[int64]string{
		7002: "Nested Materials",
	}
	err = corpAssetsRepository.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	// Get assets - this is where the bug occurred (recursive query returning NULL station_id)
	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Find the station
	var station *repositories.AssetStructure
	for _, s := range response.Structures {
		if s.ID == 60003760 {
			station = s
			break
		}
	}
	assert.NotNil(t, station, "Station should exist")

	// Find PI Materials division
	var piMaterials *repositories.CorporationHanger
	for _, hanger := range station.CorporationHangers {
		if hanger.Name == "PI Materials" {
			piMaterials = hanger
			break
		}
	}
	assert.NotNil(t, piMaterials, "PI Materials division should exist")

	// CRITICAL: The nested container should appear in the division
	// This was the bug - containers nested in offices weren't showing because station_id was NULL
	assert.Len(t, piMaterials.HangarContainers, 1, "Should have 1 container in PI Materials")

	nestedContainer := piMaterials.HangarContainers[0]
	assert.Equal(t, int64(7002), nestedContainer.ID)
	assert.Equal(t, "Nested Materials", nestedContainer.Name)

	// CRITICAL: Items should be inside the nested container
	assert.Len(t, nestedContainer.Assets, 2, "Nested container should have 2 items")

	// Verify the items
	itemNames := []string{}
	for _, asset := range nestedContainer.Assets {
		itemNames = append(itemNames, asset.Name)
	}
	assert.Contains(t, itemNames, "Tritanium")
	assert.Contains(t, itemNames, "Pyerite")
}

func Test_NestedContainersShouldNotDuplicateAcrossStations(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Multi-Station Corp",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Main Hangar",
			3: "Materials",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	// Create scenario: Corp has offices at TWO stations, but nested container should only appear at ONE
	corpAssets := []*models.EveAsset{
		// Office at station 1
		{
			ItemID:       8001,
			IsSingleton:  true,
			LocationID:   60003760, // Jita IV - Caldari Navy Assembly Plant
			LocationType: "item",
			Quantity:     1,
			TypeID:       27, // Office
			LocationFlag: "OfficeFolder",
		},
		// Office at station 2
		{
			ItemID:       8002,
			IsSingleton:  true,
			LocationID:   60003761, // Jita IV - Some Other Station
			LocationType: "item",
			Quantity:     1,
			TypeID:       27, // Office
			LocationFlag: "OfficeFolder",
		},
		// Container nested in Office at station 1 ONLY
		{
			ItemID:       8003,
			IsSingleton:  true,
			LocationID:   8001, // Inside first Office
			LocationType: "item",
			Quantity:     1,
			TypeID:       3293, // Medium Standard Container
			LocationFlag: "CorpSAG3", // Materials division
		},
		// Items inside the nested container
		{
			ItemID:       8004,
			IsSingleton:  false,
			LocationID:   8003, // Inside container
			LocationType: "item",
			Quantity:     500,
			TypeID:       34, // Tritanium
			LocationFlag: "Unlocked",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	containerNames := map[int64]string{
		8003: "Materials Container",
	}
	err = corpAssetsRepository.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	// Get assets
	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Find both stations
	var station1 *repositories.AssetStructure
	var station2 *repositories.AssetStructure
	for _, s := range response.Structures {
		if s.ID == 60003760 {
			station1 = s
		} else if s.ID == 60003761 {
			station2 = s
		}
	}
	assert.NotNil(t, station1, "Station 1 should exist")
	assert.NotNil(t, station2, "Station 2 should exist")

	// Find Materials division at station 1
	var materials1 *repositories.CorporationHanger
	for _, h := range station1.CorporationHangers {
		if h.Name == "Materials" {
			materials1 = h
			break
		}
	}
	assert.NotNil(t, materials1, "Materials division should exist at station 1")

	// CRITICAL: Container should appear at station 1 ONLY (where its Office is)
	assert.Len(t, materials1.HangarContainers, 1, "Should have 1 container at station 1")
	assert.Equal(t, "Materials Container", materials1.HangarContainers[0].Name)
	assert.Len(t, materials1.HangarContainers[0].Assets, 1, "Container should have 1 item")

	// Find Materials division at station 2
	var materials2 *repositories.CorporationHanger
	for _, h := range station2.CorporationHangers {
		if h.Name == "Materials" {
			materials2 = h
			break
		}
	}
	assert.NotNil(t, materials2, "Materials division should exist at station 2")

	// CRITICAL: Container should NOT appear at station 2 (prevents duplication bug)
	assert.Len(t, materials2.HangarContainers, 0, "Should have NO containers at station 2 (no duplication)")
	assert.Len(t, materials2.Assets, 0, "Should have NO assets at station 2")
}

func Test_AssetsShouldGetCorporationAssetsInContainers(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Main Hangar",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	corpAssets := []*models.EveAsset{
		{
			ItemID:          6001,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG1",
		},
		{
			ItemID:          6002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      6001,
			LocationType:    "item",
			Quantity:        75,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	containerNames := map[int64]string{
		6001: "Corp Container",
	}

	err = corpAssetsRepository.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)

	expectedQuantity75 := int64(75)
	expectedHangers := []*repositories.CorporationHanger{
		{
			ID:              1,
			Name:            "Main Hangar",
			CorporationID:   testCorp.ID,
			CorporationName: testCorp.Name,
			Assets:          []*repositories.Asset{},
			HangarContainers: []*repositories.AssetContainer{
				{
					ID:        6001,
					Name:      "Corp Container",
					OwnerType: "corporation",
					OwnerName: testCorp.Name,
					OwnerID:   testCorp.ID,
					Assets: []*repositories.Asset{
						{
							Name:            "Tritanium",
							TypeID:          34,
							Quantity:        75,
							Volume:          0.75,
							OwnerType:       "corporation",
							OwnerName:       testCorp.Name,
							OwnerID:         testCorp.ID,
							DesiredQuantity: nil,
							StockpileDelta:  &expectedQuantity75,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expectedHangers, response.Structures[0].CorporationHangers)
}

func Test_AssetsShouldShowEmptyCorporationDivisions(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Empty Hangar",
			2: "Another Empty Hangar",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	corpAssets := []*models.EveAsset{
		{
			ItemID:          7001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)

	expectedQuantity1 := int64(1)
	expectedHangers := []*repositories.CorporationHanger{
		{
			ID:              1,
			Name:            "Empty Hangar",
			CorporationID:   testCorp.ID,
			CorporationName: testCorp.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        1,
					Volume:          0.01,
					OwnerType:       "corporation",
					OwnerName:       testCorp.Name,
					OwnerID:         testCorp.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity1,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
		{
			ID:               2,
			Name:             "Another Empty Hangar",
			CorporationID:    testCorp.ID,
			CorporationName:  testCorp.Name,
			Assets:           []*repositories.Asset{},
			HangarContainers: []*repositories.AssetContainer{},
		},
	}

	assert.ElementsMatch(t, expectedHangers, response.Structures[0].CorporationHangers)
}

func Test_AssetsShouldGetMixedCharacterAndCorporationAssets(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)
	characterAssetsRepository := repositories.NewCharacterAssets(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Crushim deez nuts",
		UserID: 42,
	}

	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Corp Hangar",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	characterAssets := []*models.EveAsset{
		{
			ItemID:          8001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        100,
			TypeID:          34,
			LocationFlag:    "Hangar",
		},
	}

	err = characterAssetsRepository.UpdateAssets(context.Background(), testCharacter.ID, testUser.ID, characterAssets)
	assert.NoError(t, err)

	corpAssets := []*models.EveAsset{
		{
			ItemID:          8002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        200,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)

	expectedQuantityChar := int64(100)
	expectedQuantityCorp := int64(200)
	expectedCharacterAssets := []*repositories.Asset{
		{
			Name:            "Tritanium",
			TypeID:          34,
			Quantity:        100,
			Volume:          1.0,
			OwnerType:       "character",
			OwnerName:       testCharacter.Name,
			OwnerID:         testCharacter.ID,
			DesiredQuantity: nil,
			StockpileDelta:  &expectedQuantityChar,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
		},
	}

	expectedCorpHangers := []*repositories.CorporationHanger{
		{
			ID:              1,
			Name:            "Corp Hangar",
			CorporationID:   testCorp.ID,
			CorporationName: testCorp.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        200,
					Volume:          2.0,
					OwnerType:       "corporation",
					OwnerName:       testCorp.Name,
					OwnerID:         testCorp.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantityCorp,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
	}

	assert.Equal(t, expectedCharacterAssets, response.Structures[0].HangarAssets)
	assert.Equal(t, expectedCorpHangers, response.Structures[0].CorporationHangers)
}

func Test_AssetsShouldAddCorporationHangersToExistingStations(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)
	characterAssetsRepository := repositories.NewCharacterAssets(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Crushim deez nuts",
		UserID: 42,
	}

	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	testCorp1 := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "First Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp1)
	assert.NoError(t, err)

	testCorp2 := repositories.PlayerCorporation{
		ID:              2002,
		UserID:          42,
		Name:            "Second Corporation",
		EsiToken:        "token789",
		EsiRefreshToken: "refresh101",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp2)
	assert.NoError(t, err)

	characterAssets := []*models.EveAsset{
		{
			ItemID:          9001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        500,
			TypeID:          34,
			LocationFlag:    "Hangar",
		},
	}

	err = characterAssetsRepository.UpdateAssets(context.Background(), testCharacter.ID, testUser.ID, characterAssets)
	assert.NoError(t, err)

	divisions1 := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Corp1 Main",
			2: "Corp1 Secondary",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp1.ID, testUser.ID, divisions1)
	assert.NoError(t, err)

	divisions2 := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Corp2 Hangar",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp2.ID, testUser.ID, divisions2)
	assert.NoError(t, err)

	corp1Assets := []*models.EveAsset{
		{
			ItemID:          9002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        300,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
		{
			ItemID:          9003,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        150,
			TypeID:          34,
			LocationFlag:    "CorpSAG2",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp1.ID, testUser.ID, corp1Assets)
	assert.NoError(t, err)

	corp2Assets := []*models.EveAsset{
		{
			ItemID:          9004,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        250,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp2.ID, testUser.ID, corp2Assets)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)
	assert.Equal(t, int64(60003760), response.Structures[0].ID)
	assert.Equal(t, "Jita IV - Moon 4 - Caldari Navy Assembly Plant", response.Structures[0].Name)

	expectedQuantity500 := int64(500)
	expectedQuantity300 := int64(300)
	expectedQuantity150v2 := int64(150)
	expectedQuantity250 := int64(250)
	expectedCharacterAssets := []*repositories.Asset{
		{
			Name:            "Tritanium",
			TypeID:          34,
			Quantity:        500,
			Volume:          5.0,
			OwnerType:       "character",
			OwnerName:       testCharacter.Name,
			OwnerID:         testCharacter.ID,
			DesiredQuantity: nil,
			StockpileDelta:  &expectedQuantity500,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
		},
	}

	assert.Equal(t, expectedCharacterAssets, response.Structures[0].HangarAssets)

	expectedCorpHangers := []*repositories.CorporationHanger{
		{
			ID:              1,
			Name:            "Corp1 Main",
			CorporationID:   testCorp1.ID,
			CorporationName: testCorp1.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        300,
					Volume:          3.0,
					OwnerType:       "corporation",
					OwnerName:       testCorp1.Name,
					OwnerID:         testCorp1.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity300,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
		{
			ID:              2,
			Name:            "Corp1 Secondary",
			CorporationID:   testCorp1.ID,
			CorporationName: testCorp1.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        150,
					Volume:          1.5,
					OwnerType:       "corporation",
					OwnerName:       testCorp1.Name,
					OwnerID:         testCorp1.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity150v2,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
		{
			ID:              1,
			Name:            "Corp2 Hangar",
			CorporationID:   testCorp2.ID,
			CorporationName: testCorp2.Name,
			Assets: []*repositories.Asset{
				{
					Name:            "Tritanium",
					TypeID:          34,
					Quantity:        250,
					Volume:          2.5,
					OwnerType:       "corporation",
					OwnerName:       testCorp2.Name,
					OwnerID:         testCorp2.ID,
					DesiredQuantity: nil,
					StockpileDelta:  &expectedQuantity250,
			UnitPrice:    nil,
			TotalValue:   ptrFloat64(0),
			DeficitValue: ptrFloat64(0),
				},
			},
			HangarContainers: []*repositories.AssetContainer{},
		},
	}

	assert.ElementsMatch(t, expectedCorpHangers, response.Structures[0].CorporationHangers)
}

func Test_AssetsShouldNotDuplicateContainersAcrossDivisions(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	playerCorpsRepository := repositories.NewPlayerCorporations(db)
	corpAssetsRepository := repositories.NewCorporationAssets(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              2001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = playerCorpsRepository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Division 1",
			2: "Division 2",
			3: "Division 3",
		},
		Wallet: map[int]string{},
	}

	err = playerCorpsRepository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	corpAssets := []*models.EveAsset{
		// Container in Division 1
		{
			ItemID:          10001,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG1",
		},
		// Items in the Division 1 container
		{
			ItemID:          10002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      10001,
			LocationType:    "item",
			Quantity:        100,
			TypeID:          34,
			LocationFlag:    "CorpSAG1",
		},
		// Container in Division 2
		{
			ItemID:          10003,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG2",
		},
		// Items in the Division 2 container
		{
			ItemID:          10004,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      10003,
			LocationType:    "item",
			Quantity:        200,
			TypeID:          34,
			LocationFlag:    "CorpSAG2",
		},
		// Container in Division 3
		{
			ItemID:          10005,
			IsBlueprintCopy: false,
			IsSingleton:     true,
			LocationID:      60003760,
			LocationType:    "station",
			Quantity:        1,
			TypeID:          3293,
			LocationFlag:    "CorpSAG3",
		},
		// Items in the Division 3 container
		{
			ItemID:          10006,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      10005,
			LocationType:    "item",
			Quantity:        300,
			TypeID:          34,
			LocationFlag:    "CorpSAG3",
		},
	}

	err = corpAssetsRepository.Upsert(context.Background(), testCorp.ID, testUser.ID, corpAssets)
	assert.NoError(t, err)

	containerNames := map[int64]string{
		10001: "Container Alpha",
		10003: "Container Beta",
		10005: "Container Gamma",
	}

	err = corpAssetsRepository.UpsertContainerNames(context.Background(), testCorp.ID, testUser.ID, containerNames)
	assert.NoError(t, err)

	response, err := assetsRepository.GetUserAssets(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Structures, 1)

	// Verify we have exactly 3 divisions
	assert.Len(t, response.Structures[0].CorporationHangers, 3, "Should have exactly 3 divisions")

	// Find each division by name
	var division1, division2, division3 *repositories.CorporationHanger
	for _, hanger := range response.Structures[0].CorporationHangers {
		switch hanger.Name {
		case "Division 1":
			division1 = hanger
		case "Division 2":
			division2 = hanger
		case "Division 3":
			division3 = hanger
		}
	}

	assert.NotNil(t, division1, "Division 1 should exist")
	assert.NotNil(t, division2, "Division 2 should exist")
	assert.NotNil(t, division3, "Division 3 should exist")

	// CRITICAL: Each division should have EXACTLY ONE container, not all three
	assert.Len(t, division1.HangarContainers, 1, "Division 1 should have exactly 1 container")
	assert.Len(t, division2.HangarContainers, 1, "Division 2 should have exactly 1 container")
	assert.Len(t, division3.HangarContainers, 1, "Division 3 should have exactly 1 container")

	// Verify Division 1 has only Container Alpha
	assert.Equal(t, int64(10001), division1.HangarContainers[0].ID, "Division 1 should have Container Alpha")
	assert.Equal(t, "Container Alpha", division1.HangarContainers[0].Name)
	assert.Len(t, division1.HangarContainers[0].Assets, 1)
	assert.Equal(t, int64(100), division1.HangarContainers[0].Assets[0].Quantity)

	// Verify Division 2 has only Container Beta
	assert.Equal(t, int64(10003), division2.HangarContainers[0].ID, "Division 2 should have Container Beta")
	assert.Equal(t, "Container Beta", division2.HangarContainers[0].Name)
	assert.Len(t, division2.HangarContainers[0].Assets, 1)
	assert.Equal(t, int64(200), division2.HangarContainers[0].Assets[0].Quantity)

	// Verify Division 3 has only Container Gamma
	assert.Equal(t, int64(10005), division3.HangarContainers[0].ID, "Division 3 should have Container Gamma")
	assert.Equal(t, "Container Gamma", division3.HangarContainers[0].Name)
	assert.Len(t, division3.HangarContainers[0].Assets, 1)
	assert.Equal(t, int64(300), division3.HangarContainers[0].Assets[0].Quantity)
}

func Test_AssetsShouldGetUserAssetsSummary(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)
	characterAssetsRepository := repositories.NewCharacterAssets(db)
	marketPricesRepo := repositories.NewMarketPrices(db)
	stockpileMarkersRepo := repositories.NewStockpileMarkers(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}
	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Test Character",
		UserID: 42,
	}
	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	// Add character assets
	characterAssets := []*models.EveAsset{
		{
			ItemID:       1001,
			LocationID:   60003760,
			LocationType: "station",
			Quantity:     1000, // 1000 Tritanium
			TypeID:       34,
			LocationFlag: "Hangar",
		},
		{
			ItemID:       1002,
			LocationID:   60003760,
			LocationType: "station",
			Quantity:     500, // 500 Pyerite
			TypeID:       35,
			LocationFlag: "Hangar",
		},
	}
	err = characterAssetsRepository.UpdateAssets(context.Background(), testCharacter.ID, testUser.ID, characterAssets)
	assert.NoError(t, err)

	// Add market prices
	marketPrices := []models.MarketPrice{
		{
			TypeID:    34, // Tritanium
			RegionID:  10000002,
			BuyPrice:  ptrFloat64(5.0),
			SellPrice: ptrFloat64(5.0),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
		{
			TypeID:    35, // Pyerite
			RegionID:  10000002,
			BuyPrice:  ptrFloat64(10.0),
			SellPrice: ptrFloat64(10.0),
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
	err = marketPricesRepo.UpsertPrices(context.Background(), marketPrices)
	assert.NoError(t, err)

	// Add stockpile marker for Tritanium (desired 2000, have 1000 = deficit of 1000)
	stockpileMarker := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         testCharacter.ID,
		LocationID:      60003760,
		DesiredQuantity: 2000,
	}
	err = stockpileMarkersRepo.Upsert(context.Background(), stockpileMarker)
	assert.NoError(t, err)

	// Get summary
	summary, err := assetsRepository.GetUserAssetsSummary(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	// Expected calculations:
	// Total Value:
	//   - 1000 Tritanium * 5.0 = 5000
	//   - 500 Pyerite * 10.0 = 5000
	//   Total = 10000
	// Total Deficit:
	//   - Tritanium deficit: (2000 - 1000) * 5.0 = 5000
	//   - Pyerite has no marker, so no deficit
	//   Total = 5000

	assert.Equal(t, 10000.0, summary.TotalValue, "Total value should be 10000 ISK")
	assert.Equal(t, 5000.0, summary.TotalDeficit, "Total deficit should be 5000 ISK")
}

func Test_AssetsShouldReturnZeroSummaryForNoAssets(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepository := repositories.NewUserRepository(db)
	assetsRepository := repositories.NewAssets(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}
	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Get summary with no assets
	summary, err := assetsRepository.GetUserAssetsSummary(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)

	assert.Equal(t, 0.0, summary.TotalValue, "Total value should be 0 for user with no assets")
	assert.Equal(t, 0.0, summary.TotalDeficit, "Total deficit should be 0 for user with no assets")
}

func ptrFloat64(v float64) *float64 {
	return &v
}
