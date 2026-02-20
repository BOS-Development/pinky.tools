package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_BuyOrders_CreateAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5000, Name: "Test Buyer"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item type
	itemTypes := []models.EveInventoryType{
		{TypeID: 60, TypeName: "Mexallon", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create buy order
	order := &models.BuyOrder{
		BuyerUserID:     5000,
		TypeID:          60,
		LocationID:      60003760,
		QuantityDesired: 100000,
		MaxPricePerUnit: 50,
		IsActive:        true,
	}

	err = repo.Create(context.Background(), order)
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)
	assert.NotZero(t, order.CreatedAt)
	assert.NotZero(t, order.UpdatedAt)

	// Get by ID
	retrieved, err := repo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, retrieved.ID)
	assert.Equal(t, int64(5000), retrieved.BuyerUserID)
	assert.Equal(t, int64(60), retrieved.TypeID)
	assert.Equal(t, "Mexallon", retrieved.TypeName)
	assert.Equal(t, int64(100000), retrieved.QuantityDesired)
	assert.Equal(t, float64(50), retrieved.MaxPricePerUnit)
	assert.True(t, retrieved.IsActive)
}

func Test_BuyOrders_GetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5010, Name: "Test Buyer"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 61, TypeName: "Isogen", Volume: 0.01},
		{TypeID: 62, TypeName: "Nocxium", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create multiple buy orders
	for i := 0; i < 3; i++ {
		order := &models.BuyOrder{
			BuyerUserID:     5010,
			TypeID:          61 + int64(i%2),
			LocationID:      60003760,
			QuantityDesired: int64(10000 * (i + 1)),
			MaxPricePerUnit: float64(50 + i),
			IsActive:        true,
		}
		err = repo.Create(context.Background(), order)
		assert.NoError(t, err)
	}

	// Get by user
	orders, err := repo.GetByUser(context.Background(), 5010)
	assert.NoError(t, err)
	assert.Len(t, orders, 3)

	// Verify ordering (DESC by created_at)
	assert.GreaterOrEqual(t, orders[0].CreatedAt, orders[1].CreatedAt)
	assert.GreaterOrEqual(t, orders[1].CreatedAt, orders[2].CreatedAt)

	// Verify type names populated
	assert.NotEmpty(t, orders[0].TypeName)
}

func Test_BuyOrders_Update(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5020, Name: "Test Buyer"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item type
	itemTypes := []models.EveInventoryType{
		{TypeID: 63, TypeName: "Zydrine", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create buy order
	order := &models.BuyOrder{
		BuyerUserID:     5020,
		TypeID:          63,
		LocationID:      60003760,
		QuantityDesired: 50000,
		MaxPricePerUnit: 100,
		IsActive:        true,
	}
	err = repo.Create(context.Background(), order)
	assert.NoError(t, err)

	// Update order
	order.QuantityDesired = 75000
	order.MaxPricePerUnit = 120
	notes := "Urgent order"
	order.Notes = &notes

	err = repo.Update(context.Background(), order)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(75000), retrieved.QuantityDesired)
	assert.Equal(t, float64(120), retrieved.MaxPricePerUnit)
	assert.NotNil(t, retrieved.Notes)
	assert.Equal(t, "Urgent order", *retrieved.Notes)
}

func Test_BuyOrders_Delete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5030, Name: "Test Buyer"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item type
	itemTypes := []models.EveInventoryType{
		{TypeID: 64, TypeName: "Megacyte", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create buy order
	order := &models.BuyOrder{
		BuyerUserID:     5030,
		TypeID:          64,
		LocationID:      60003760,
		QuantityDesired: 25000,
		MaxPricePerUnit: 150,
		IsActive:        true,
	}
	err = repo.Create(context.Background(), order)
	assert.NoError(t, err)

	// Delete order
	err = repo.Delete(context.Background(), order.ID, 5030)
	assert.NoError(t, err)

	// Verify soft delete (is_active = false)
	retrieved, err := repo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.False(t, retrieved.IsActive)
}

func Test_BuyOrders_GetDemandForSeller(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	contactsRepo := repositories.NewContacts(db)
	permRepo := repositories.NewContactPermissions(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)

	// Create buyer and seller
	buyer := &repositories.User{ID: 5040, Name: "Buyer"}
	err = userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	seller := &repositories.User{ID: 5041, Name: "Seller"}
	err = userRepo.Add(context.Background(), seller)
	assert.NoError(t, err)

	// Create characters
	buyerChar := &repositories.Character{ID: 50400, Name: "Buyer Character", UserID: 5040}
	err = charRepo.Add(context.Background(), buyerChar)
	assert.NoError(t, err)

	sellerChar := &repositories.Character{ID: 50410, Name: "Seller Character", UserID: 5041}
	err = charRepo.Add(context.Background(), sellerChar)
	assert.NoError(t, err)

	// Create contact relationship
	contact, err := contactsRepo.Create(context.Background(), 5040, 5041)
	assert.NoError(t, err)

	// Accept contact
	_, err = contactsRepo.UpdateStatus(context.Background(), contact.ID, 5041, "accepted")
	assert.NoError(t, err)

	// Grant permission from buyer to seller (buyer grants seller permission to see buyer's buy orders)
	perm := &models.ContactPermission{
		ContactID:       contact.ID,
		GrantingUserID:  5040,
		ReceivingUserID: 5041,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	}
	err = permRepo.Upsert(context.Background(), perm)
	assert.NoError(t, err)

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 65, TypeName: "Tritanium", Volume: 0.01},
		{TypeID: 66, TypeName: "Pyerite", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create buy orders from buyer
	order1 := &models.BuyOrder{
		BuyerUserID:     5040,
		TypeID:          65,
		LocationID:      60003760,
		QuantityDesired: 500000,
		MaxPricePerUnit: 6,
		IsActive:        true,
	}
	err = buyOrdersRepo.Create(context.Background(), order1)
	assert.NoError(t, err)

	order2 := &models.BuyOrder{
		BuyerUserID:     5040,
		TypeID:          66,
		LocationID:      60003760,
		QuantityDesired: 250000,
		MaxPricePerUnit: 15,
		IsActive:        true,
	}
	err = buyOrdersRepo.Create(context.Background(), order2)
	assert.NoError(t, err)

	// Create inactive order (should not appear)
	order3 := &models.BuyOrder{
		BuyerUserID:     5040,
		TypeID:          65,
		LocationID:      60003760,
		QuantityDesired: 100000,
		MaxPricePerUnit: 10,
		IsActive:        false,
	}
	err = buyOrdersRepo.Create(context.Background(), order3)
	assert.NoError(t, err)

	// Get demand for seller
	demand, err := buyOrdersRepo.GetDemandForSeller(context.Background(), 5041)
	assert.NoError(t, err)
	assert.Len(t, demand, 2) // Only active orders

	// Verify orders are from buyer and have type names
	for _, order := range demand {
		assert.Equal(t, int64(5040), order.BuyerUserID)
		assert.NotEmpty(t, order.TypeName)
		assert.True(t, order.IsActive)
	}
}

func Test_BuyOrders_GetByID_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewBuyOrders(db)

	_, err = repo.GetByID(context.Background(), 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buy order not found")
}

func Test_BuyOrders_Update_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewBuyOrders(db)

	order := &models.BuyOrder{
		ID:              999999,
		LocationID:      60003760,
		QuantityDesired: 100,
		MaxPricePerUnit: 50,
		IsActive:        true,
	}

	err = repo.Update(context.Background(), order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buy order not found")
}

func Test_BuyOrders_Delete_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewBuyOrders(db)

	err = repo.Delete(context.Background(), 999999, 5000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buy order not found")
}

func Test_BuyOrders_UpsertAutoBuy(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	autoBuyRepo := repositories.NewAutoBuyConfigs(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5100, Name: "Auto Buyer"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item type
	itemTypes := []models.EveInventoryType{
		{TypeID: 70, TypeName: "Morphite", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create auto_buy_config
	config := &models.AutoBuyConfig{
		UserID:          5100,
		OwnerType:       "character",
		OwnerID:         12345,
		LocationID:      60003760,
		PricePercentage: 100.0,
		PriceSource:     "jita_sell",
	}
	err = autoBuyRepo.Upsert(context.Background(), config)
	assert.NoError(t, err)
	assert.NotZero(t, config.ID)

	// Test UpsertAutoBuy creates an order with auto_buy_config_id set
	order := &models.BuyOrder{
		BuyerUserID:     5100,
		TypeID:          70,
		LocationID:      60003760,
		QuantityDesired: 5000,
		MaxPricePerUnit: 80.0,
		AutoBuyConfigID: &config.ID,
	}

	err = repo.UpsertAutoBuy(context.Background(), order)
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)
	assert.True(t, order.IsActive)
	assert.NotZero(t, order.CreatedAt)
	assert.NotZero(t, order.UpdatedAt)

	// Verify via GetByID
	retrieved, err := repo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(5100), retrieved.BuyerUserID)
	assert.Equal(t, int64(70), retrieved.TypeID)
	assert.Equal(t, int64(5000), retrieved.QuantityDesired)
	assert.Equal(t, 80.0, retrieved.MaxPricePerUnit)
	assert.NotNil(t, retrieved.AutoBuyConfigID)
	assert.Equal(t, config.ID, *retrieved.AutoBuyConfigID)
	assert.True(t, retrieved.IsActive)

	// Test UpsertAutoBuy updates existing order on conflict
	order.QuantityDesired = 10000
	order.MaxPricePerUnit = 90.0
	notes := "Updated by auto-buy"
	order.Notes = &notes

	err = repo.UpsertAutoBuy(context.Background(), order)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(context.Background(), order.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(10000), updated.QuantityDesired)
	assert.Equal(t, 90.0, updated.MaxPricePerUnit)
	assert.NotNil(t, updated.Notes)
	assert.Equal(t, "Updated by auto-buy", *updated.Notes)

	// Verify it's the same record (same ID, not a new row)
	assert.Equal(t, order.ID, updated.ID)
}

func Test_BuyOrders_GetActiveAutoBuyOrders(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	autoBuyRepo := repositories.NewAutoBuyConfigs(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5110, Name: "Auto Buyer 2"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 71, TypeName: "Fullerite-C50", Volume: 1.0},
		{TypeID: 72, TypeName: "Fullerite-C60", Volume: 1.0},
		{TypeID: 73, TypeName: "Fullerite-C70", Volume: 1.0},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create auto_buy_config
	config := &models.AutoBuyConfig{
		UserID:          5110,
		OwnerType:       "character",
		OwnerID:         12346,
		LocationID:      60003760,
		PricePercentage: 95.0,
		PriceSource:     "jita_sell",
	}
	err = autoBuyRepo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// Create multiple auto-buy orders
	typeIDs := []int64{71, 72, 73}
	for _, typeID := range typeIDs {
		order := &models.BuyOrder{
			BuyerUserID:     5110,
			TypeID:          typeID,
			LocationID:      60003760,
			QuantityDesired: 1000,
			MaxPricePerUnit: 50.0,
			AutoBuyConfigID: &config.ID,
		}
		err = repo.UpsertAutoBuy(context.Background(), order)
		assert.NoError(t, err)
	}

	// Verify GetActiveAutoBuyOrders returns all 3 active orders
	active, err := repo.GetActiveAutoBuyOrders(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.Len(t, active, 3)

	for _, order := range active {
		assert.Equal(t, int64(5110), order.BuyerUserID)
		assert.True(t, order.IsActive)
		assert.NotNil(t, order.AutoBuyConfigID)
		assert.Equal(t, config.ID, *order.AutoBuyConfigID)
	}
}

func Test_BuyOrders_DeactivateAutoBuyOrders(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	autoBuyRepo := repositories.NewAutoBuyConfigs(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5120, Name: "Auto Buyer 3"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 74, TypeName: "Fullerite-C320", Volume: 1.0},
		{TypeID: 75, TypeName: "Fullerite-C540", Volume: 1.0},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create auto_buy_config
	config := &models.AutoBuyConfig{
		UserID:          5120,
		OwnerType:       "character",
		OwnerID:         12347,
		LocationID:      60003760,
		PricePercentage: 100.0,
		PriceSource:     "jita_sell",
	}
	err = autoBuyRepo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// Create auto-buy orders
	for _, typeID := range []int64{74, 75} {
		order := &models.BuyOrder{
			BuyerUserID:     5120,
			TypeID:          typeID,
			LocationID:      60003760,
			QuantityDesired: 2000,
			MaxPricePerUnit: 100.0,
			AutoBuyConfigID: &config.ID,
		}
		err = repo.UpsertAutoBuy(context.Background(), order)
		assert.NoError(t, err)
	}

	// Verify orders are active
	active, err := repo.GetActiveAutoBuyOrders(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.Len(t, active, 2)

	// Deactivate all orders for the config
	err = repo.DeactivateAutoBuyOrders(context.Background(), config.ID)
	assert.NoError(t, err)

	// Verify all orders are now inactive
	active, err = repo.GetActiveAutoBuyOrders(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.Len(t, active, 0)

	// Verify orders still exist but are inactive via GetByUser
	allOrders, err := repo.GetByUser(context.Background(), 5120)
	assert.NoError(t, err)
	assert.Len(t, allOrders, 2)
	for _, order := range allOrders {
		assert.False(t, order.IsActive)
	}
}

func Test_BuyOrders_DeactivateAutoBuyOrder(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	autoBuyRepo := repositories.NewAutoBuyConfigs(db)
	repo := repositories.NewBuyOrders(db)

	// Create user
	user := &repositories.User{ID: 5130, Name: "Auto Buyer 4"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create item types
	itemTypes := []models.EveInventoryType{
		{TypeID: 76, TypeName: "Carbon", Volume: 0.01},
		{TypeID: 77, TypeName: "Silicon", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create auto_buy_config
	config := &models.AutoBuyConfig{
		UserID:          5130,
		OwnerType:       "character",
		OwnerID:         12348,
		LocationID:      60003760,
		PricePercentage: 100.0,
		PriceSource:     "jita_sell",
	}
	err = autoBuyRepo.Upsert(context.Background(), config)
	assert.NoError(t, err)

	// Create two auto-buy orders
	order1 := &models.BuyOrder{
		BuyerUserID:     5130,
		TypeID:          76,
		LocationID:      60003760,
		QuantityDesired: 3000,
		MaxPricePerUnit: 25.0,
		AutoBuyConfigID: &config.ID,
	}
	err = repo.UpsertAutoBuy(context.Background(), order1)
	assert.NoError(t, err)

	order2 := &models.BuyOrder{
		BuyerUserID:     5130,
		TypeID:          77,
		LocationID:      60003760,
		QuantityDesired: 4000,
		MaxPricePerUnit: 30.0,
		AutoBuyConfigID: &config.ID,
	}
	err = repo.UpsertAutoBuy(context.Background(), order2)
	assert.NoError(t, err)

	// Deactivate only order1
	err = repo.DeactivateAutoBuyOrder(context.Background(), order1.ID)
	assert.NoError(t, err)

	// Verify order1 is inactive
	retrieved1, err := repo.GetByID(context.Background(), order1.ID)
	assert.NoError(t, err)
	assert.False(t, retrieved1.IsActive)

	// Verify order2 is still active
	retrieved2, err := repo.GetByID(context.Background(), order2.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved2.IsActive)

	// Verify GetActiveAutoBuyOrders returns only order2
	active, err := repo.GetActiveAutoBuyOrders(context.Background(), config.ID)
	assert.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, order2.ID, active[0].ID)
}
