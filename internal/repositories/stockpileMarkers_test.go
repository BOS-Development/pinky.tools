package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_StockpileMarkersShouldUpsertAndRetrieve(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create a stockpile marker for character assets
	marker1 := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          34, // Tritanium
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 10000,
		Notes:           stringPtr("For building"),
	}

	err = stockpileRepo.Upsert(context.Background(), marker1)
	assert.NoError(t, err)

	// Create a marker for corporation assets with division
	divisionNum := 3
	marker2 := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          35, // Pyerite
		OwnerType:       "corporation",
		OwnerID:         2001,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  &divisionNum,
		DesiredQuantity: 5000,
		Notes:           nil,
	}

	err = stockpileRepo.Upsert(context.Background(), marker2)
	assert.NoError(t, err)

	// Create a marker with container
	containerID := int64(5003)
	marker3 := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          36, // Mexallon
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     &containerID,
		DivisionNumber:  nil,
		DesiredQuantity: 2500,
		Notes:           stringPtr("In container"),
	}

	err = stockpileRepo.Upsert(context.Background(), marker3)
	assert.NoError(t, err)

	// Retrieve all markers for the user
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 3)

	// Verify marker1
	assert.Equal(t, testUser.ID, markers[0].UserID)
	assert.Equal(t, int64(34), markers[0].TypeID)
	assert.Equal(t, "character", markers[0].OwnerType)
	assert.Equal(t, int64(1337), markers[0].OwnerID)
	assert.Equal(t, int64(60003760), markers[0].LocationID)
	assert.Nil(t, markers[0].ContainerID)
	assert.Nil(t, markers[0].DivisionNumber)
	assert.Equal(t, int64(10000), markers[0].DesiredQuantity)
	assert.NotNil(t, markers[0].Notes)
	assert.Equal(t, "For building", *markers[0].Notes)

	// Verify marker2
	assert.Equal(t, int64(35), markers[1].TypeID)
	assert.Equal(t, "corporation", markers[1].OwnerType)
	assert.NotNil(t, markers[1].DivisionNumber)
	assert.Equal(t, 3, *markers[1].DivisionNumber)
	assert.Nil(t, markers[1].Notes)

	// Verify marker3
	assert.Equal(t, int64(36), markers[2].TypeID)
	assert.NotNil(t, markers[2].ContainerID)
	assert.Equal(t, int64(5003), *markers[2].ContainerID)
}

func Test_StockpileMarkersShouldUpdateOnConflict(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Insert initial marker
	marker := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 1000,
		Notes:           stringPtr("Initial note"),
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Update the marker with new values
	marker.DesiredQuantity = 5000
	marker.Notes = stringPtr("Updated note")

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Retrieve and verify update
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1, "Should still have only one marker after update")

	assert.Equal(t, int64(5000), markers[0].DesiredQuantity)
	assert.NotNil(t, markers[0].Notes)
	assert.Equal(t, "Updated note", *markers[0].Notes)
}

func Test_StockpileMarkersShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create two markers
	marker1 := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 1000,
		Notes:           nil,
	}

	marker2 := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          35,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 2000,
		Notes:           nil,
	}

	err = stockpileRepo.Upsert(context.Background(), marker1)
	assert.NoError(t, err)

	err = stockpileRepo.Upsert(context.Background(), marker2)
	assert.NoError(t, err)

	// Verify both exist
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 2)

	// Delete marker1
	err = stockpileRepo.Delete(context.Background(), marker1)
	assert.NoError(t, err)

	// Verify only marker2 remains
	markers, err = stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)
	assert.Equal(t, int64(35), markers[0].TypeID)
}

func Test_StockpileMarkersShouldHandleNullValues(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create markers at same location but different containers/divisions
	// This tests the COALESCE logic in the composite key
	containerID := int64(5003)
	divisionNum := 2

	markers := []*models.StockpileMarker{
		{
			UserID:          testUser.ID,
			TypeID:          34,
			OwnerType:       "character",
			OwnerID:         1337,
			LocationID:      60003760,
			ContainerID:     nil,
			DivisionNumber:  nil,
			DesiredQuantity: 1000,
			Notes:           nil,
		},
		{
			UserID:          testUser.ID,
			TypeID:          34,
			OwnerType:       "character",
			OwnerID:         1337,
			LocationID:      60003760,
			ContainerID:     &containerID,
			DivisionNumber:  nil,
			DesiredQuantity: 2000,
			Notes:           nil,
		},
		{
			UserID:          testUser.ID,
			TypeID:          34,
			OwnerType:       "corporation",
			OwnerID:         2001,
			LocationID:      60003760,
			ContainerID:     nil,
			DivisionNumber:  &divisionNum,
			DesiredQuantity: 3000,
			Notes:           nil,
		},
	}

	for _, marker := range markers {
		err = stockpileRepo.Upsert(context.Background(), marker)
		assert.NoError(t, err)
	}

	// Should have 3 distinct markers (different combinations of container/division)
	retrieved, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Delete the one with container
	err = stockpileRepo.Delete(context.Background(), markers[1])
	assert.NoError(t, err)

	// Should have 2 remaining
	retrieved, err = stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func Test_StockpileMarkersShouldIsolateByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	user1 := &repositories.User{
		ID:   42,
		Name: "User 1",
	}

	user2 := &repositories.User{
		ID:   43,
		Name: "User 2",
	}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)

	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	// Create marker for user1
	marker1 := &models.StockpileMarker{
		UserID:          user1.ID,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 1000,
		Notes:           nil,
	}

	// Create marker for user2
	marker2 := &models.StockpileMarker{
		UserID:          user2.ID,
		TypeID:          34,
		OwnerType:       "character",
		OwnerID:         1337,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 2000,
		Notes:           nil,
	}

	err = stockpileRepo.Upsert(context.Background(), marker1)
	assert.NoError(t, err)

	err = stockpileRepo.Upsert(context.Background(), marker2)
	assert.NoError(t, err)

	// User1 should only see their marker
	markers1, err := stockpileRepo.GetByUser(context.Background(), user1.ID)
	assert.NoError(t, err)
	assert.Len(t, markers1, 1)
	assert.Equal(t, int64(1000), markers1[0].DesiredQuantity)

	// User2 should only see their marker
	markers2, err := stockpileRepo.GetByUser(context.Background(), user2.ID)
	assert.NoError(t, err)
	assert.Len(t, markers2, 1)
	assert.Equal(t, int64(2000), markers2[0].DesiredQuantity)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}

func Test_StockpileMarkers_UpsertWithPricing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6100,
		Name: "Pricing Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create marker with PriceSource and PricePercentage set
	priceSource := "jita_sell"
	pricePercentage := 95.5
	marker := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          34, // Tritanium
		OwnerType:       "character",
		OwnerID:         61001,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 50000,
		Notes:           stringPtr("With pricing"),
		PriceSource:     &priceSource,
		PricePercentage: &pricePercentage,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify pricing fields persist via GetByUser
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)

	assert.Equal(t, int64(34), markers[0].TypeID)
	assert.Equal(t, int64(50000), markers[0].DesiredQuantity)
	assert.NotNil(t, markers[0].PriceSource)
	assert.Equal(t, "jita_sell", *markers[0].PriceSource)
	assert.NotNil(t, markers[0].PricePercentage)
	assert.Equal(t, 95.5, *markers[0].PricePercentage)
}

func Test_StockpileMarkers_UpsertWithNilPricing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6110,
		Name: "Nil Pricing Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create marker without pricing fields (nil)
	marker := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          35, // Pyerite
		OwnerType:       "character",
		OwnerID:         61101,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 25000,
		Notes:           nil,
		PriceSource:     nil,
		PricePercentage: nil,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify pricing fields come back as nil
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)

	assert.Equal(t, int64(35), markers[0].TypeID)
	assert.Equal(t, int64(25000), markers[0].DesiredQuantity)
	assert.Nil(t, markers[0].PriceSource)
	assert.Nil(t, markers[0].PricePercentage)
}

func Test_StockpileMarkers_UpsertWithAutoProduction(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6130,
		Name: "Auto Production Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create a production plan to link
	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        testUser.ID,
		ProductTypeID: 34,
		Name:          "Tritanium Plan",
	})
	assert.NoError(t, err)
	assert.NotNil(t, plan)

	// Create marker with auto-production fields set
	marker := &models.StockpileMarker{
		UserID:                    testUser.ID,
		TypeID:                    34, // Tritanium
		OwnerType:                 "character",
		OwnerID:                   61301,
		LocationID:                60003760,
		ContainerID:               nil,
		DivisionNumber:            nil,
		DesiredQuantity:           100000,
		Notes:                     nil,
		PlanID:                    &plan.ID,
		AutoProductionParallelism: 3,
		AutoProductionEnabled:     true,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify auto-production fields persist via GetByUser
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)

	assert.Equal(t, int64(34), markers[0].TypeID)
	assert.NotNil(t, markers[0].PlanID)
	assert.Equal(t, plan.ID, *markers[0].PlanID)
	assert.Equal(t, 3, markers[0].AutoProductionParallelism)
	assert.True(t, markers[0].AutoProductionEnabled)
}

func Test_StockpileMarkers_UpsertWithAutoProductionDisabled(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6140,
		Name: "Auto Production Disabled Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create marker without auto-production (defaults)
	marker := &models.StockpileMarker{
		UserID:                    testUser.ID,
		TypeID:                    35, // Pyerite
		OwnerType:                 "character",
		OwnerID:                   61401,
		LocationID:                60003760,
		DesiredQuantity:           50000,
		PlanID:                    nil,
		AutoProductionParallelism: 1,
		AutoProductionEnabled:     false,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)

	assert.Nil(t, markers[0].PlanID)
	assert.Equal(t, 1, markers[0].AutoProductionParallelism)
	assert.False(t, markers[0].AutoProductionEnabled)
}

func Test_StockpileMarkers_GetAutoProductionMarkers(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	user1 := &repositories.User{ID: 6150, Name: "AutoProd User 1"}
	user2 := &repositories.User{ID: 6151, Name: "AutoProd User 2"}

	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	plan1, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user1.ID,
		ProductTypeID: 34,
		Name:          "Plan A",
	})
	assert.NoError(t, err)

	plan2, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user2.ID,
		ProductTypeID: 35,
		Name:          "Plan B",
	})
	assert.NoError(t, err)

	// Marker 1: auto-production enabled with plan
	err = stockpileRepo.Upsert(context.Background(), &models.StockpileMarker{
		UserID:                    user1.ID,
		TypeID:                    34,
		OwnerType:                 "character",
		OwnerID:                   61501,
		LocationID:                60003760,
		DesiredQuantity:           10000,
		PlanID:                    &plan1.ID,
		AutoProductionParallelism: 2,
		AutoProductionEnabled:     true,
	})
	assert.NoError(t, err)

	// Marker 2: auto-production enabled with plan (different user)
	err = stockpileRepo.Upsert(context.Background(), &models.StockpileMarker{
		UserID:                    user2.ID,
		TypeID:                    35,
		OwnerType:                 "character",
		OwnerID:                   61511,
		LocationID:                60003760,
		DesiredQuantity:           20000,
		PlanID:                    &plan2.ID,
		AutoProductionParallelism: 1,
		AutoProductionEnabled:     true,
	})
	assert.NoError(t, err)

	// Marker 3: auto-production disabled (should not be returned)
	err = stockpileRepo.Upsert(context.Background(), &models.StockpileMarker{
		UserID:                    user1.ID,
		TypeID:                    36,
		OwnerType:                 "character",
		OwnerID:                   61501,
		LocationID:                60003760,
		DesiredQuantity:           5000,
		PlanID:                    nil,
		AutoProductionParallelism: 1,
		AutoProductionEnabled:     false,
	})
	assert.NoError(t, err)

	// Marker 4: auto-production enabled but no plan (should not be returned)
	err = stockpileRepo.Upsert(context.Background(), &models.StockpileMarker{
		UserID:                    user1.ID,
		TypeID:                    27,
		OwnerType:                 "character",
		OwnerID:                   61501,
		LocationID:                60003760,
		DesiredQuantity:           3000,
		PlanID:                    nil,
		AutoProductionParallelism: 1,
		AutoProductionEnabled:     true,
	})
	assert.NoError(t, err)

	// GetAutoProductionMarkers should return only markers 1 and 2
	autoMarkers, err := stockpileRepo.GetAutoProductionMarkers(context.Background())
	assert.NoError(t, err)

	// Filter to only the markers from our test users (other tests may have added data)
	var filtered []*models.StockpileMarker
	for _, m := range autoMarkers {
		if m.UserID == user1.ID || m.UserID == user2.ID {
			filtered = append(filtered, m)
		}
	}

	assert.Len(t, filtered, 2)

	// Should be ordered by user_id, plan_id — user1 first
	assert.Equal(t, user1.ID, filtered[0].UserID)
	assert.Equal(t, plan1.ID, *filtered[0].PlanID)
	assert.Equal(t, 2, filtered[0].AutoProductionParallelism)
	assert.True(t, filtered[0].AutoProductionEnabled)

	assert.Equal(t, user2.ID, filtered[1].UserID)
	assert.Equal(t, plan2.ID, *filtered[1].PlanID)
	assert.Equal(t, 1, filtered[1].AutoProductionParallelism)
	assert.True(t, filtered[1].AutoProductionEnabled)
}

func Test_StockpileMarkers_UpdateAutoProduction(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6160,
		Name: "Update AutoProd Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        testUser.ID,
		ProductTypeID: 34,
		Name:          "Update Plan",
	})
	assert.NoError(t, err)

	// Create marker without auto-production
	marker := &models.StockpileMarker{
		UserID:                    testUser.ID,
		TypeID:                    34,
		OwnerType:                 "character",
		OwnerID:                   61601,
		LocationID:                60003760,
		DesiredQuantity:           10000,
		PlanID:                    nil,
		AutoProductionParallelism: 1,
		AutoProductionEnabled:     false,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Enable auto-production via upsert (ON CONFLICT)
	marker.PlanID = &plan.ID
	marker.AutoProductionParallelism = 4
	marker.AutoProductionEnabled = true

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify update
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1, "Should still have one marker after upsert")

	assert.NotNil(t, markers[0].PlanID)
	assert.Equal(t, plan.ID, *markers[0].PlanID)
	assert.Equal(t, 4, markers[0].AutoProductionParallelism)
	assert.True(t, markers[0].AutoProductionEnabled)
}

func Test_StockpileMarkers_UpdatePricing(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	setupTestUniverse(t, db)

	userRepo := repositories.NewUserRepository(db)
	stockpileRepo := repositories.NewStockpileMarkers(db)

	testUser := &repositories.User{
		ID:   6120,
		Name: "Update Pricing Test User",
	}

	err = userRepo.Add(context.Background(), testUser)
	assert.NoError(t, err)

	// Create marker without pricing
	marker := &models.StockpileMarker{
		UserID:          testUser.ID,
		TypeID:          36, // Mexallon
		OwnerType:       "character",
		OwnerID:         61201,
		LocationID:      60003760,
		ContainerID:     nil,
		DivisionNumber:  nil,
		DesiredQuantity: 10000,
		Notes:           nil,
		PriceSource:     nil,
		PricePercentage: nil,
	}

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify initially no pricing
	markers, err := stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1)
	assert.Nil(t, markers[0].PriceSource)
	assert.Nil(t, markers[0].PricePercentage)

	// Update with pricing via Upsert (ON CONFLICT)
	marker.PriceSource = stringPtr("jita_buy")
	marker.PricePercentage = float64Ptr(110.0)
	marker.DesiredQuantity = 15000

	err = stockpileRepo.Upsert(context.Background(), marker)
	assert.NoError(t, err)

	// Verify pricing is updated
	markers, err = stockpileRepo.GetByUser(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, markers, 1, "Should still have only one marker after upsert")

	assert.Equal(t, int64(36), markers[0].TypeID)
	assert.Equal(t, int64(15000), markers[0].DesiredQuantity)
	assert.NotNil(t, markers[0].PriceSource)
	assert.Equal(t, "jita_buy", *markers[0].PriceSource)
	assert.NotNil(t, markers[0].PricePercentage)
	assert.Equal(t, 110.0, *markers[0].PricePercentage)
}
