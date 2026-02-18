package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSalesAnalytics_GetSalesMetrics(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)


	ctx := context.Background()

	// Create item types first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
	}
	require.NoError(t, itemTypeRepo.UpsertItemTypes(ctx, itemTypes))

	// Create test users
	userRepo := repositories.NewUserRepository(db)
	sellerUser := &repositories.User{ID: 999001, Name: "Seller"}
	buyerUser := &repositories.User{ID: 999002, Name: "Buyer"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions
	txRepo := repositories.NewPurchaseTransactions(db)

	// Transaction 1: 100 units @ 1000 ISK each = 100,000 ISK (2 days ago)
	tx1 := &models.PurchaseTransaction{
		ForSaleItemID:     1,
		BuyerUserID:       buyerUser.ID,
		SellerUserID:      sellerUser.ID,
		TypeID:            34,
		QuantityPurchased: 100,
		PricePerUnit:      1000,
		TotalPrice:        100000,
		Status:            "completed",
	}

	// Transaction 2: 50 units @ 2000 ISK each = 100,000 ISK (1 day ago)
	tx2 := &models.PurchaseTransaction{
		ForSaleItemID:     2,
		BuyerUserID:       buyerUser.ID,
		SellerUserID:      sellerUser.ID,
		TypeID:            35,
		QuantityPurchased: 50,
		PricePerUnit:      2000,
		TotalPrice:        100000,
		Status:            "completed",
	}

	dbTx, err := db.Begin()
	require.NoError(t, err)

	require.NoError(t, txRepo.Create(ctx, dbTx, tx1))
	require.NoError(t, txRepo.Create(ctx, dbTx, tx2))
	require.NoError(t, dbTx.Commit())

	// Test GetSalesMetrics
	analyticsRepo := repositories.NewSalesAnalytics(db)

	t.Run("Get metrics for all time", func(t *testing.T) {
		metrics, err := analyticsRepo.GetSalesMetrics(ctx, sellerUser.ID, 0)
		require.NoError(t, err)

		assert.Equal(t, float64(200000), metrics.TotalRevenue)
		assert.Equal(t, int64(2), metrics.TotalTransactions)
		assert.Equal(t, int64(150), metrics.TotalQuantitySold)
		assert.Equal(t, int64(2), metrics.UniqueItemTypes)
		assert.Equal(t, int64(1), metrics.UniqueBuyers)
		assert.NotEmpty(t, metrics.TimeSeriesData)
		assert.NotEmpty(t, metrics.TopItems)
	})

	t.Run("Get metrics for 30 days", func(t *testing.T) {
		metrics, err := analyticsRepo.GetSalesMetrics(ctx, sellerUser.ID, 30)
		require.NoError(t, err)

		assert.Equal(t, float64(200000), metrics.TotalRevenue)
		assert.Equal(t, int64(2), metrics.TotalTransactions)
	})
}

func TestSalesAnalytics_GetTopItems(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)


	ctx := context.Background()

	// Create item types first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
	}
	require.NoError(t, itemTypeRepo.UpsertItemTypes(ctx, itemTypes))

	// Create test users
	userRepo := repositories.NewUserRepository(db)
	sellerUser := &repositories.User{ID: 999003, Name: "Seller2"}
	buyerUser := &repositories.User{ID: 999004, Name: "Buyer2"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions for different items
	txRepo := repositories.NewPurchaseTransactions(db)

	transactions := []*models.PurchaseTransaction{
		{
			ForSaleItemID:     10,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34, // Tritanium
			QuantityPurchased: 1000,
			PricePerUnit:      10,
			TotalPrice:        10000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     11,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            35, // Pyerite
			QuantityPurchased: 500,
			PricePerUnit:      20,
			TotalPrice:        10000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     12,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34, // Tritanium again
			QuantityPurchased: 2000,
			PricePerUnit:      15,
			TotalPrice:        30000,
			Status:            "completed",
		},
	}

	dbTx, err := db.Begin()
	require.NoError(t, err)

	for _, tx := range transactions {
		require.NoError(t, txRepo.Create(ctx, dbTx, tx))
	}
	require.NoError(t, dbTx.Commit())

	analyticsRepo := repositories.NewSalesAnalytics(db)

	t.Run("Get top items", func(t *testing.T) {
		items, err := analyticsRepo.GetTopItems(ctx, sellerUser.ID, 0, 10)
		require.NoError(t, err)

		assert.Len(t, items, 2) // 2 unique item types

		// First item should be Tritanium (higher revenue)
		assert.Equal(t, int64(34), items[0].TypeID)
		assert.Equal(t, "Tritanium", items[0].TypeName)
		assert.Equal(t, int64(3000), items[0].QuantitySold)
		assert.Equal(t, float64(40000), items[0].Revenue)
		assert.Equal(t, int64(2), items[0].TransactionCount)

		// Second item should be Pyerite
		assert.Equal(t, int64(35), items[1].TypeID)
		assert.Equal(t, "Pyerite", items[1].TypeName)
		assert.Equal(t, int64(500), items[1].QuantitySold)
		assert.Equal(t, float64(10000), items[1].Revenue)
		assert.Equal(t, int64(1), items[1].TransactionCount)
	})

	t.Run("Limit top items", func(t *testing.T) {
		items, err := analyticsRepo.GetTopItems(ctx, sellerUser.ID, 0, 1)
		require.NoError(t, err)

		assert.Len(t, items, 1)
		assert.Equal(t, int64(34), items[0].TypeID) // Only top item
	})
}

func TestSalesAnalytics_GetBuyerAnalytics(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)


	ctx := context.Background()

	// Create item types first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
	}
	require.NoError(t, itemTypeRepo.UpsertItemTypes(ctx, itemTypes))

	// Create test users
	userRepo := repositories.NewUserRepository(db)
	sellerUser := &repositories.User{ID: 999005, Name: "Seller3"}
	buyer1 := &repositories.User{ID: 999006, Name: "Buyer3"}
	buyer2 := &repositories.User{ID: 999007, Name: "Buyer4"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyer1))
	require.NoError(t, userRepo.Add(ctx, buyer2))

	// Create purchase transactions
	txRepo := repositories.NewPurchaseTransactions(db)

	transactions := []*models.PurchaseTransaction{
		// Buyer1 makes 2 purchases (repeat customer)
		{
			ForSaleItemID:     20,
			BuyerUserID:       buyer1.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34,
			QuantityPurchased: 100,
			PricePerUnit:      1000,
			TotalPrice:        100000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     21,
			BuyerUserID:       buyer1.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            35,
			QuantityPurchased: 50,
			PricePerUnit:      2000,
			TotalPrice:        100000,
			Status:            "completed",
		},
		// Buyer2 makes 1 purchase
		{
			ForSaleItemID:     22,
			BuyerUserID:       buyer2.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34,
			QuantityPurchased: 25,
			PricePerUnit:      1000,
			TotalPrice:        25000,
			Status:            "completed",
		},
	}

	dbTx, err := db.Begin()
	require.NoError(t, err)

	for _, tx := range transactions {
		require.NoError(t, txRepo.Create(ctx, dbTx, tx))
	}
	require.NoError(t, dbTx.Commit())

	// Wait a moment to ensure timestamps differ
	time.Sleep(10 * time.Millisecond)

	analyticsRepo := repositories.NewSalesAnalytics(db)

	t.Run("Get buyer analytics", func(t *testing.T) {
		buyers, err := analyticsRepo.GetBuyerAnalytics(ctx, sellerUser.ID, 0, 10)
		require.NoError(t, err)

		assert.Len(t, buyers, 2)

		// First buyer should be buyer1 (higher total spent)
		assert.Equal(t, buyer1.ID, buyers[0].BuyerUserID)
		assert.Equal(t, float64(200000), buyers[0].TotalSpent)
		assert.Equal(t, int64(2), buyers[0].TotalPurchases)
		assert.Equal(t, int64(150), buyers[0].TotalQuantity)
		assert.True(t, buyers[0].RepeatCustomer)

		// Second buyer should be buyer2
		assert.Equal(t, buyer2.ID, buyers[1].BuyerUserID)
		assert.Equal(t, float64(25000), buyers[1].TotalSpent)
		assert.Equal(t, int64(1), buyers[1].TotalPurchases)
		assert.Equal(t, int64(25), buyers[1].TotalQuantity)
		assert.False(t, buyers[1].RepeatCustomer)
	})

	t.Run("Limit buyer analytics", func(t *testing.T) {
		buyers, err := analyticsRepo.GetBuyerAnalytics(ctx, sellerUser.ID, 0, 1)
		require.NoError(t, err)

		assert.Len(t, buyers, 1)
		assert.Equal(t, buyer1.ID, buyers[0].BuyerUserID) // Only top buyer
	})
}

func TestSalesAnalytics_GetItemSalesHistory(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)


	ctx := context.Background()

	// Create item types first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: nil},
	}
	require.NoError(t, itemTypeRepo.UpsertItemTypes(ctx, itemTypes))

	// Create test users
	userRepo := repositories.NewUserRepository(db)
	sellerUser := &repositories.User{ID: 999008, Name: "Seller4"}
	buyerUser := &repositories.User{ID: 999009, Name: "Buyer5"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions for specific item
	txRepo := repositories.NewPurchaseTransactions(db)

	transactions := []*models.PurchaseTransaction{
		{
			ForSaleItemID:     30,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34, // Tritanium
			QuantityPurchased: 1000,
			PricePerUnit:      10,
			TotalPrice:        10000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     31,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34, // Tritanium
			QuantityPurchased: 2000,
			PricePerUnit:      12,
			TotalPrice:        24000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     32,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            35, // Pyerite (different item)
			QuantityPurchased: 500,
			PricePerUnit:      20,
			TotalPrice:        10000,
			Status:            "completed",
		},
	}

	dbTx, err := db.Begin()
	require.NoError(t, err)

	for _, tx := range transactions {
		require.NoError(t, txRepo.Create(ctx, dbTx, tx))
	}
	require.NoError(t, dbTx.Commit())

	analyticsRepo := repositories.NewSalesAnalytics(db)

	t.Run("Get item sales history for Tritanium", func(t *testing.T) {
		item, err := analyticsRepo.GetItemSalesHistory(ctx, sellerUser.ID, 34, 0)
		require.NoError(t, err)

		assert.Equal(t, int64(34), item.TypeID)
		assert.Equal(t, "Tritanium", item.TypeName)
		assert.Equal(t, int64(3000), item.QuantitySold)
		assert.Equal(t, float64(34000), item.Revenue)
		assert.Equal(t, int64(2), item.TransactionCount)
		assert.Equal(t, float64(11), item.AveragePricePerUnit) // (10 + 12) / 2 = 11
	})

	t.Run("Get item sales history for Pyerite", func(t *testing.T) {
		item, err := analyticsRepo.GetItemSalesHistory(ctx, sellerUser.ID, 35, 0)
		require.NoError(t, err)

		assert.Equal(t, int64(35), item.TypeID)
		assert.Equal(t, "Pyerite", item.TypeName)
		assert.Equal(t, int64(500), item.QuantitySold)
		assert.Equal(t, float64(10000), item.Revenue)
		assert.Equal(t, int64(1), item.TransactionCount)
		assert.Equal(t, float64(20), item.AveragePricePerUnit)
	})

	t.Run("Get item sales history for non-existent item", func(t *testing.T) {
		_, err := analyticsRepo.GetItemSalesHistory(ctx, sellerUser.ID, 999, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no sales data found")
	})
}
