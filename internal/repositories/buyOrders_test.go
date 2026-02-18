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
		QuantityDesired: 500000,
		MaxPricePerUnit: 6,
		IsActive:        true,
	}
	err = buyOrdersRepo.Create(context.Background(), order1)
	assert.NoError(t, err)

	order2 := &models.BuyOrder{
		BuyerUserID:     5040,
		TypeID:          66,
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
