package repositories_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func setupPurchaseTestData(t *testing.T, db *sql.DB, buyerID, sellerID, typeID, locationID int64) (*models.ForSaleItem, error) {
	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	itemTypesRepo := repositories.NewItemTypeRepository(db)
	forSaleRepo := repositories.NewForSaleItems(db)

	// Create buyer
	buyer := &repositories.User{ID: buyerID, Name: "Test Buyer"}
	err := userRepo.Add(context.Background(), buyer)
	assert.NoError(t, err)

	buyerChar := &repositories.Character{ID: buyerID * 10, Name: "Buyer Character", UserID: buyerID}
	err = charRepo.Add(context.Background(), buyerChar)
	assert.NoError(t, err)

	// Create seller
	seller := &repositories.User{ID: sellerID, Name: "Test Seller"}
	err = userRepo.Add(context.Background(), seller)
	assert.NoError(t, err)

	sellerChar := &repositories.Character{ID: sellerID * 10, Name: "Seller Character", UserID: sellerID}
	err = charRepo.Add(context.Background(), sellerChar)
	assert.NoError(t, err)

	// Create item type
	itemTypes := []models.EveInventoryType{
		{TypeID: typeID, TypeName: "Test Item", Volume: 0.01},
	}
	err = itemTypesRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Create region, constellation, solar system
	_, err = db.ExecContext(context.Background(),
		"INSERT INTO regions (region_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		10000002, "The Forge")
	assert.NoError(t, err)

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO constellations (constellation_id, name, region_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
		20000020, "Kimotoro", 10000002)
	assert.NoError(t, err)

	_, err = db.ExecContext(context.Background(),
		"INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		locationID, "Test System", 20000020, 0.9)
	assert.NoError(t, err)

	// Create for-sale item
	item := &models.ForSaleItem{
		UserID:            sellerID,
		TypeID:            typeID,
		OwnerType:         "character",
		OwnerID:           sellerID * 10,
		LocationID:        locationID,
		QuantityAvailable: 1000,
		PricePerUnit:      100,
		IsActive:          true,
	}

	err = forSaleRepo.Upsert(context.Background(), item)
	assert.NoError(t, err)

	return item, nil
}

func Test_PurchaseTransactions_CreateAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3000, 3001, 40, 30000150)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	// Begin transaction
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	// Create purchase transaction
	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3000,
		SellerUserID:      3001,
		TypeID:            40,
		QuantityPurchased: 100,
		PricePerUnit:      100,
		TotalPrice:        10000,
		Status:            "pending",
	}

	err = repo.Create(context.Background(), tx, purchase)
	assert.NoError(t, err)
	assert.NotZero(t, purchase.ID)
	assert.NotZero(t, purchase.PurchasedAt)

	err = tx.Commit()
	assert.NoError(t, err)

	// Get by ID
	retrieved, err := repo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, purchase.ID, retrieved.ID)
	assert.Equal(t, int64(3000), retrieved.BuyerUserID)
	assert.Equal(t, int64(3001), retrieved.SellerUserID)
	assert.Equal(t, int64(100), retrieved.QuantityPurchased)
	assert.Equal(t, "pending", retrieved.Status)
}

func Test_PurchaseTransactions_UpdateStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3010, 3011, 41, 30000151)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3010,
		SellerUserID:      3011,
		TypeID:            41,
		QuantityPurchased: 50,
		PricePerUnit:      100,
		TotalPrice:        5000,
		Status:            "pending",
	}

	err = repo.Create(context.Background(), tx, purchase)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)

	// Update status to contract_created
	err = repo.UpdateStatus(context.Background(), purchase.ID, "contract_created")
	assert.NoError(t, err)

	// Verify status changed
	retrieved, err := repo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "contract_created", retrieved.Status)

	// Update to completed
	err = repo.UpdateStatus(context.Background(), purchase.ID, "completed")
	assert.NoError(t, err)

	retrieved, err = repo.GetByID(context.Background(), purchase.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrieved.Status)
}

func Test_PurchaseTransactions_UpdateContractKeys(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3020, 3021, 42, 30000152)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	// Create multiple purchases
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	purchase1 := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3020,
		SellerUserID:      3021,
		TypeID:            42,
		QuantityPurchased: 25,
		PricePerUnit:      100,
		TotalPrice:        2500,
		Status:            "pending",
	}

	purchase2 := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3020,
		SellerUserID:      3021,
		TypeID:            42,
		QuantityPurchased: 75,
		PricePerUnit:      100,
		TotalPrice:        7500,
		Status:            "pending",
	}

	err = repo.Create(context.Background(), tx, purchase1)
	assert.NoError(t, err)

	err = repo.Create(context.Background(), tx, purchase2)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Update contract keys for both purchases
	contractKey := "PT-3020-30000152-1234567890"
	err = repo.UpdateContractKeys(context.Background(), []int64{purchase1.ID, purchase2.ID}, contractKey)
	assert.NoError(t, err)

	// Verify contract keys set
	retrieved1, err := repo.GetByID(context.Background(), purchase1.ID)
	assert.NoError(t, err)
	assert.Equal(t, contractKey, *retrieved1.ContractKey)

	retrieved2, err := repo.GetByID(context.Background(), purchase2.ID)
	assert.NoError(t, err)
	assert.Equal(t, contractKey, *retrieved2.ContractKey)
}

func Test_PurchaseTransactions_GetByBuyer(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3030, 3031, 43, 30000153)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	// Create multiple purchases
	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	for i := 0; i < 3; i++ {
		purchase := &models.PurchaseTransaction{
			ForSaleItemID:     item.ID,
			BuyerUserID:       3030,
			SellerUserID:      3031,
			TypeID:            43,
			QuantityPurchased: int64(10 + i),
			PricePerUnit:      100.0,
			TotalPrice:        float64((10 + i) * 100),
			Status:            "pending",
		}
		err = repo.Create(context.Background(), tx, purchase)
		assert.NoError(t, err)
	}

	err = tx.Commit()
	assert.NoError(t, err)

	// Get buyer history
	transactions, err := repo.GetByBuyer(context.Background(), 3030)
	assert.NoError(t, err)
	assert.Len(t, transactions, 3)

	// Verify they're sorted by purchase date DESC
	assert.GreaterOrEqual(t, transactions[0].PurchasedAt, transactions[1].PurchasedAt)
	assert.GreaterOrEqual(t, transactions[1].PurchasedAt, transactions[2].PurchasedAt)

	// Verify type name is populated
	assert.Equal(t, "Test Item", transactions[0].TypeName)
}

func Test_PurchaseTransactions_GetBySeller(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3040, 3041, 44, 30000154)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	for i := 0; i < 2; i++ {
		purchase := &models.PurchaseTransaction{
			ForSaleItemID:     item.ID,
			BuyerUserID:       3040,
			SellerUserID:      3041,
			TypeID:            44,
			QuantityPurchased: int64(20 + i),
			PricePerUnit:      100.0,
			TotalPrice:        float64((20 + i) * 100),
			Status:            "completed",
		}
		err = repo.Create(context.Background(), tx, purchase)
		assert.NoError(t, err)
	}

	err = tx.Commit()
	assert.NoError(t, err)

	// Get seller history
	transactions, err := repo.GetBySeller(context.Background(), 3041)
	assert.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.Equal(t, "Test Item", transactions[0].TypeName)
}

func Test_PurchaseTransactions_GetPendingForSeller(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	item, err := setupPurchaseTestData(t, db, 3050, 3051, 45, 30000155)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	tx, err := db.BeginTx(context.Background(), nil)
	assert.NoError(t, err)
	defer tx.Rollback()

	// Create pending purchase
	pendingPurchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3050,
		SellerUserID:      3051,
		TypeID:            45,
		QuantityPurchased: 30,
		PricePerUnit:      100,
		TotalPrice:        3000,
		Status:            "pending",
	}
	err = repo.Create(context.Background(), tx, pendingPurchase)
	assert.NoError(t, err)

	// Create completed purchase (should not appear)
	completedPurchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       3050,
		SellerUserID:      3051,
		TypeID:            45,
		QuantityPurchased: 40,
		PricePerUnit:      100,
		TotalPrice:        4000,
		Status:            "completed",
	}
	err = repo.Create(context.Background(), tx, completedPurchase)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Get pending sales
	pending, err := repo.GetPendingForSeller(context.Background(), 3051)
	assert.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, "pending", pending[0].Status)
	assert.Equal(t, int64(30), pending[0].QuantityPurchased)

	// Verify buyer name is populated (from users table, not characters)
	assert.Equal(t, "Test Buyer", pending[0].BuyerName)

	// Verify location name is populated
	assert.Equal(t, "Test System", pending[0].LocationName)
}

func Test_PurchaseTransactions_GetByID_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	_, err = repo.GetByID(context.Background(), 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purchase transaction not found")
}

func Test_PurchaseTransactions_UpdateStatus_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	err = repo.UpdateStatus(context.Background(), 999999, "completed")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "purchase transaction not found")
}

func Test_PurchaseTransactions_UpdateContractKeys_EmptyArray(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewPurchaseTransactions(db)

	// Should return nil without error
	err = repo.UpdateContractKeys(context.Background(), []int64{}, "some-key")
	assert.NoError(t, err)
}
