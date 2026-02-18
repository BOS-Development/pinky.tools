package controllers_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsController_GetSalesMetrics(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)
	defer db.Close()

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
	sellerUser := &repositories.User{ID: 998001, Name: "SellerTest"}
	buyerUser := &repositories.User{ID: 998002, Name: "BuyerTest"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions
	txRepo := repositories.NewPurchaseTransactions(db)
	dbTx, err := db.Begin()
	require.NoError(t, err)

	tx1 := &models.PurchaseTransaction{
		ForSaleItemID:     100,
		BuyerUserID:       buyerUser.ID,
		SellerUserID:      sellerUser.ID,
		TypeID:            34,
		QuantityPurchased: 100,
		PricePerUnit:      1000,
		TotalPrice:        100000,
		Status:            "completed",
	}
	require.NoError(t, txRepo.Create(ctx, dbTx, tx1))
	require.NoError(t, dbTx.Commit())

	// Setup controller
	analyticsRepo := repositories.NewSalesAnalytics(db)
	controller := controllers.NewAnalytics(&MockRouter{}, analyticsRepo)

	t.Run("Get sales metrics - all time", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/sales", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetSalesMetrics(args)
		require.Nil(t, httpErr)
		require.NotNil(t, result)

		metrics := result.(*models.SalesMetrics)
		assert.Equal(t, float64(100000), metrics.TotalRevenue)
		assert.Equal(t, int64(1), metrics.TotalTransactions)
		assert.Equal(t, int64(100), metrics.TotalQuantitySold)
	})

	t.Run("Get sales metrics - with period", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/sales?period=30d", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetSalesMetrics(args)
		require.Nil(t, httpErr)
		require.NotNil(t, result)

		metrics := result.(*models.SalesMetrics)
		assert.Equal(t, float64(100000), metrics.TotalRevenue)
	})
}

func TestAnalyticsController_GetTopItems(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)
	defer db.Close()

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
	sellerUser := &repositories.User{ID: 998003, Name: "SellerTest2"}
	buyerUser := &repositories.User{ID: 998004, Name: "BuyerTest2"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions
	txRepo := repositories.NewPurchaseTransactions(db)
	dbTx, err := db.Begin()
	require.NoError(t, err)

	transactions := []*models.PurchaseTransaction{
		{
			ForSaleItemID:     101,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            34,
			QuantityPurchased: 1000,
			PricePerUnit:      10,
			TotalPrice:        10000,
			Status:            "completed",
		},
		{
			ForSaleItemID:     102,
			BuyerUserID:       buyerUser.ID,
			SellerUserID:      sellerUser.ID,
			TypeID:            35,
			QuantityPurchased: 500,
			PricePerUnit:      20,
			TotalPrice:        10000,
			Status:            "completed",
		},
	}

	for _, tx := range transactions {
		require.NoError(t, txRepo.Create(ctx, dbTx, tx))
	}
	require.NoError(t, dbTx.Commit())

	// Setup controller
	analyticsRepo := repositories.NewSalesAnalytics(db)
	controller := controllers.NewAnalytics(&MockRouter{}, analyticsRepo)

	t.Run("Get top items", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/top-items", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetTopItems(args)
		require.Nil(t, httpErr)
		require.NotNil(t, result)

		items := result.([]models.ItemSalesData)
		assert.Len(t, items, 2)
	})

	t.Run("Get top items with limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/top-items?limit=1", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetTopItems(args)
		require.Nil(t, httpErr)
		require.NotNil(t, result)

		items := result.([]models.ItemSalesData)
		assert.Len(t, items, 1)
	})
}

func TestAnalyticsController_GetItemSalesHistory(t *testing.T) {
	db, err := setupDatabase()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Create item types first
	itemTypeRepo := repositories.NewItemTypeRepository(db)
	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: nil},
	}
	require.NoError(t, itemTypeRepo.UpsertItemTypes(ctx, itemTypes))

	// Create test users
	userRepo := repositories.NewUserRepository(db)
	sellerUser := &repositories.User{ID: 998005, Name: "SellerTest4"}
	buyerUser := &repositories.User{ID: 998006, Name: "BuyerTest5"}
	require.NoError(t, userRepo.Add(ctx, sellerUser))
	require.NoError(t, userRepo.Add(ctx, buyerUser))

	// Create purchase transactions
	txRepo := repositories.NewPurchaseTransactions(db)
	dbTx, err := db.Begin()
	require.NoError(t, err)

	tx1 := &models.PurchaseTransaction{
		ForSaleItemID:     105,
		BuyerUserID:       buyerUser.ID,
		SellerUserID:      sellerUser.ID,
		TypeID:            34,
		QuantityPurchased: 1000,
		PricePerUnit:      10,
		TotalPrice:        10000,
		Status:            "completed",
	}
	require.NoError(t, txRepo.Create(ctx, dbTx, tx1))
	require.NoError(t, dbTx.Commit())

	// Setup controller
	analyticsRepo := repositories.NewSalesAnalytics(db)
	controller := controllers.NewAnalytics(&MockRouter{}, analyticsRepo)

	t.Run("Get item sales history", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/item-history?typeId=34", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetItemSalesHistory(args)
		require.Nil(t, httpErr)
		require.NotNil(t, result)

		item := result.(*models.ItemSalesData)
		assert.Equal(t, int64(34), item.TypeID)
		assert.Equal(t, "Tritanium", item.TypeName)
		assert.Equal(t, int64(1000), item.QuantitySold)
	})

	t.Run("Get item sales history - missing typeId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/item-history", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetItemSalesHistory(args)
		require.NotNil(t, httpErr)
		assert.Nil(t, result)
		assert.Equal(t, 400, httpErr.StatusCode)
	})

	t.Run("Get item sales history - invalid typeId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/analytics/item-history?typeId=invalid", nil)

		testUserID := sellerUser.ID
		args := &web.HandlerArgs{
			Request: req,
			User:    &testUserID,
		}

		result, httpErr := controller.GetItemSalesHistory(args)
		require.NotNil(t, httpErr)
		assert.Nil(t, result)
		assert.Equal(t, 400, httpErr.StatusCode)
	})
}
