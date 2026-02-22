package repositories_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_AutoFulfill_NoDuplicatesOnRepeatedCycles verifies that running the full
// auto-sell → auto-buy → auto-fulfill cycle multiple times does NOT create duplicate
// purchases. Uses SyncForAllUsers to match the market price updater path.
//
// Scenario:
//   - Seller has 100 Tritanium in a character hangar, auto-sell configured
//   - Buyer has a stockpile deficit of 100 Tritanium, auto-buy configured
//   - Run 5 cycles of SyncForAllUsers — should only ever have 1 pending purchase
func Test_AutoFulfill_NoDuplicatesOnRepeatedCycles(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	// --- Setup users ---
	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	buyerUserID := int64(8010)
	sellerUserID := int64(8020)
	buyerCharID := int64(80101)
	sellerCharID := int64(80201)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller Char", UserID: sellerUserID}))

	// --- Setup mutual permissions ---
	_, err = db.ExecContext(ctx, `
		INSERT INTO contacts (requester_user_id, recipient_user_id, status)
		VALUES ($1, $2, 'accepted')
	`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{
		ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID,
		ServiceType: "for_sale_browse", CanAccess: true,
	}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{
		ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID,
		ServiceType: "for_sale_browse", CanAccess: true,
	}))

	// --- Setup seller: character assets with 100 Tritanium in hangar ---
	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 500001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	// --- Setup auto-sell container for seller ---
	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	// --- Setup buyer: stockpile marker wanting 100 Tritanium ---
	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	// --- Setup auto-buy config for buyer ---
	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	// --- Setup Jita market prices ---
	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	// --- Create updaters ---
	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === Run 5 cycles using SyncForAllUsers (market price updater path) ===
	for cycle := 1; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx), "cycle %d: auto-sell failed", cycle)
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx), "cycle %d: auto-buy failed", cycle)
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx), "cycle %d: auto-fulfill failed", cycle)

		// Count pending purchases
		purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)

		pendingCount := 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}

		if cycle == 1 {
			require.Equal(t, 1, pendingCount, "Cycle 1: expected exactly 1 pending purchase")
			assert.Equal(t, int64(100), purchases[0].QuantityPurchased)
			assert.True(t, purchases[0].IsAutoFulfilled)
		} else {
			assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending purchase (got %d total, %d pending)", cycle, len(purchases), pendingCount)
		}

		// Log table state for debugging
		t.Logf("=== Cycle %d: %d total purchases, %d pending ===", cycle, len(purchases), pendingCount)
		for _, p := range purchases {
			t.Logf("  purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d",
				p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID)
		}

		var fsCount, fsActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM for_sale_items`).Scan(&fsCount, &fsActiveCount)
		require.NoError(t, err)
		t.Logf("  for_sale_items: %d total, %d active", fsCount, fsActiveCount)

		var boCount, boActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM buy_orders`).Scan(&boCount, &boActiveCount)
		require.NoError(t, err)
		t.Logf("  buy_orders: %d total, %d active", boCount, boActiveCount)
	}
}

// Test_AutoFulfill_SyncForAllUsers_ThenSyncForUser verifies that running
// SyncForAllUsers followed by SyncForUser (simulating market price update
// followed by asset refresh) does NOT create duplicates.
func Test_AutoFulfill_SyncForAllUsers_ThenSyncForUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8010)
	sellerUserID := int64(8020)
	buyerCharID := int64(80101)
	sellerCharID := int64(80201)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 500001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === Cycle 1: SyncForAllUsers (market price update) ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected exactly 1 purchase")

	// === Cycle 2: SyncForUser (asset refresh for seller, then auto-buy+fulfill for buyer) ===
	require.NoError(t, autoSellUpdater.SyncForUser(ctx, sellerUserID))
	require.NoError(t, autoBuyUpdater.SyncForUser(ctx, buyerUserID))
	require.NoError(t, autoFulfillUpdater.SyncForUser(ctx, buyerUserID))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
		}
	}
	assert.Equal(t, 1, pendingCount, "Cycle 2 (SyncForUser after SyncForAllUsers): expected still 1 pending")

	// === Cycle 3: SyncForAllUsers again ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount = 0
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
		}
	}
	assert.Equal(t, 1, pendingCount, "Cycle 3 (SyncForAllUsers again): expected still 1 pending")
}

// Test_AutoFulfill_NoDuplicateAfterSellerGetsMoreInventory verifies that when the
// seller gets more inventory (auto-sell creates a NEW for-sale item because the old
// one was deactivated at qty=0), auto-fulfill does NOT create a duplicate purchase.
func Test_AutoFulfill_NoDuplicateAfterSellerGetsMoreInventory(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8030)
	sellerUserID := int64(8040)
	buyerCharID := int64(80301)
	sellerCharID := int64(80401)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer2"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller2"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer2 Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller2 Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 600001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: creates purchase for 100 (consumes entire for-sale item → deactivated) ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected exactly 1 purchase")
	t.Logf("Cycle 1: purchase id=%d qty=%d forSaleItem=%d buyOrder=%v",
		purchases[0].ID, purchases[0].QuantityPurchased, purchases[0].ForSaleItemID, purchases[0].BuyOrderID)

	// === Seller gets more inventory (e.g., from mining) ===
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 600001, LocationID: 60003760, LocationType: "other", Quantity: 200, TypeID: 34, LocationFlag: "Hangar"},
	}))

	// === CYCLE 2: seller has more — for-sale item gets NEW ID (old was deactivated) ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
		}
		t.Logf("Cycle 2: purchase id=%d qty=%d status=%s forSaleItem=%d buyOrder=%v",
			p.ID, p.QuantityPurchased, p.Status, p.ForSaleItemID, p.BuyOrderID)
	}
	assert.Equal(t, 1, pendingCount, "Cycle 2 (more inventory): expected still 1 pending purchase")

	// === CYCLES 3-5: repeated runs ===
	for cycle := 3; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount = 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}
		assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending purchase", cycle)
	}
}

// Test_AutoFulfill_NoDuplicateAfterCompletedPurchase verifies behavior after a
// purchase is completed and the buyer's assets haven't been refreshed yet.
func Test_AutoFulfill_NoDuplicateAfterCompletedPurchase(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8050)
	sellerUserID := int64(8060)
	buyerCharID := int64(80501)
	sellerCharID := int64(80601)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer3"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller3"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer3 Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller3 Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 700001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: creates purchase ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1)

	// === Simulate purchase completion (contract accepted) ===
	_, err = db.ExecContext(ctx, `UPDATE purchase_transactions SET status = 'completed' WHERE id = $1`, purchases[0].ID)
	require.NoError(t, err)

	// === CYCLE 2: completed purchase, buyer assets NOT refreshed ===
	// Pending counts are now 0, deficit is still 100 → system creates a new purchase.
	// This IS expected (the old purchase completed, buyer still needs items).
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
		}
		t.Logf("After completion: purchase id=%d qty=%d status=%s", p.ID, p.QuantityPurchased, p.Status)
	}
	assert.Equal(t, 1, pendingCount, "After completed: expected exactly 1 new pending purchase")

	// === CYCLES 3-5: should NOT keep creating more ===
	for cycle := 3; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount = 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}
		assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending after completed+re-purchase", cycle)
	}
}

// Test_AutoFulfill_PartialFillNoDuplicate verifies that when the seller has LESS
// than the buyer wants, a partial purchase is created and subsequent cycles don't
// create duplicates. Also tests what happens when the seller restocks.
func Test_AutoFulfill_PartialFillNoDuplicate(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8070)
	sellerUserID := int64(8080)
	buyerCharID := int64(80701)
	sellerCharID := int64(80801)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer4"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller4"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer4 Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller4 Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	// Seller has only 50, buyer wants 100
	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 800001, LocationID: 60003760, LocationType: "other", Quantity: 50, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: partial fill — buys 50 out of 100 wanted ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected 1 purchase")
	assert.Equal(t, int64(50), purchases[0].QuantityPurchased, "Should buy only what seller has")
	t.Logf("Cycle 1: purchase id=%d qty=%d forSaleItem=%d buyOrder=%v",
		purchases[0].ID, purchases[0].QuantityPurchased, purchases[0].ForSaleItemID, purchases[0].BuyOrderID)

	// === CYCLES 2-3: no more inventory, should NOT create duplicates ===
	for cycle := 2; cycle <= 3; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount := 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}
		assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending purchase (partial fill)", cycle)
	}

	// === Seller restocks: now has 100 total (50 committed + 50 new) ===
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 800001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	// === CYCLE 4: seller restocked — system should have at most 100 total pending qty ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	totalPendingQty := int64(0)
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
			totalPendingQty += p.QuantityPurchased
		}
		t.Logf("Cycle 4: purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d",
			p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID)
	}
	t.Logf("Cycle 4: %d pending purchases, total qty=%d", pendingCount, totalPendingQty)
	assert.LessOrEqual(t, totalPendingQty, int64(100), "Total pending should not exceed desired quantity of 100")
}

// Test_AutoFulfill_MarketPriceChange_NoDuplicate verifies that when market prices
// change between cycles (causing price recalculation), duplicates are NOT created.
func Test_AutoFulfill_MarketPriceChange_NoDuplicate(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8090)
	sellerUserID := int64(8100)
	buyerCharID := int64(80901)
	sellerCharID := int64(81001)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "Buyer5"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "Seller5"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "Buyer5 Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "Seller5 Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 900001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	initialPrice := 5.0
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &initialPrice, SellPrice: &initialPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: initial price ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected 1 purchase")

	// === Change market price ===
	newPrice := 6.0
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &newPrice, SellPrice: &newPrice},
	}))

	// === CYCLE 2: price changed ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
		}
	}
	assert.Equal(t, 1, pendingCount, "Cycle 2 (price change): expected still 1 pending")

	// === Change price drastically ===
	bigPrice := 10.0
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &bigPrice, SellPrice: &bigPrice},
	}))

	// === CYCLES 3-5: repeated price changes ===
	for cycle := 3; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount = 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}
		assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending after price changes", cycle)
	}
}

// Test_AutoFulfill_ConcurrentPipeline verifies that running the full pipeline
// concurrently (simulating market price update + asset refresh overlap) does NOT
// create duplicate purchases. This models the real-world scenario where:
//   - Market price runner calls SyncForAllUsers on auto-sell/buy/fulfill
//   - Asset refresh calls SyncForUser on auto-sell/buy/fulfill for a specific user
//   - Both can run simultaneously
func Test_AutoFulfill_ConcurrentPipeline(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8110)
	sellerUserID := int64(8120)
	buyerCharID := int64(81101)
	sellerCharID := int64(81201)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "ConcBuyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "ConcSeller"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "ConcBuyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "ConcSeller Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 1100001, LocationID: 60003760, LocationType: "other", Quantity: 200, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === Cycle 1: sequential setup to establish initial state ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected exactly 1 purchase")
	t.Logf("Cycle 1: purchase id=%d qty=%d", purchases[0].ID, purchases[0].QuantityPurchased)

	// === Concurrent cycles: simulate overlapping market price + asset refresh ===
	for round := 1; round <= 10; round++ {
		var wg sync.WaitGroup

		// Goroutine 1: SyncForAllUsers (market price runner path)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = autoSellUpdater.SyncForAllUsers(ctx)
			_ = autoBuyUpdater.SyncForAllUsers(ctx)
			_ = autoFulfillUpdater.SyncForAllUsers(ctx)
		}()

		// Goroutine 2: SyncForUser for buyer (asset refresh path)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = autoSellUpdater.SyncForUser(ctx, sellerUserID)
			_ = autoBuyUpdater.SyncForUser(ctx, buyerUserID)
			_ = autoFulfillUpdater.SyncForUser(ctx, buyerUserID)
		}()

		wg.Wait()

		// Check for duplicates after each concurrent round
		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)

		pendingCount := 0
		totalPendingQty := int64(0)
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
				totalPendingQty += p.QuantityPurchased
			}
		}

		t.Logf("Concurrent round %d: %d total purchases, %d pending, %d total pending qty",
			round, len(purchases), pendingCount, totalPendingQty)
		for _, p := range purchases {
			t.Logf("  purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d auto=%v",
				p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID, p.IsAutoFulfilled)
		}

		assert.LessOrEqual(t, totalPendingQty, int64(100),
			"Round %d: total pending qty should not exceed desired (100), got %d across %d purchases",
			round, totalPendingQty, pendingCount)
	}
}

// Test_AutoFulfill_ForSaleItemDeactivatedAndRecreated verifies the scenario where
// auto-fulfill fully consumes a for-sale item (qty→0, is_active=false), and then
// auto-sell creates a NEW for-sale item (new ID) because the old one is inactive.
// The pending query in auto-sell must still account for the purchase linked to the
// old for-sale item, otherwise auto-sell resets the quantity too high.
func Test_AutoFulfill_ForSaleItemDeactivatedAndRecreated(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8130)
	sellerUserID := int64(8140)
	buyerCharID := int64(81301)
	sellerCharID := int64(81401)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "FSBuyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "FSSeller"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "FSBuyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "FSSeller Char", UserID: sellerUserID}))

	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	// Seller has exactly what buyer wants — for-sale will be fully consumed (qty→0)
	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, sellerCharID, sellerUserID, []*models.EveAsset{
		{ItemID: 1300001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "character", OwnerID: sellerCharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: purchase consumes entire for-sale item ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1)
	assert.Equal(t, int64(100), purchases[0].QuantityPurchased)
	origForSaleID := purchases[0].ForSaleItemID

	// Verify the for-sale item is now inactive
	// Note: UpdateQuantity preserves quantity_available when deactivating (positive qty constraint)
	var fsActive bool
	var fsQty int64
	err = db.QueryRowContext(ctx, `SELECT is_active, quantity_available FROM for_sale_items WHERE id = $1`, origForSaleID).Scan(&fsActive, &fsQty)
	require.NoError(t, err)
	assert.False(t, fsActive, "For-sale item should be inactive after full consumption")

	t.Logf("Cycle 1: purchase id=%d, for_sale_item=%d (now inactive, qty=%d)", purchases[0].ID, origForSaleID, fsQty)

	// === CYCLES 2-5: auto-sell will try to create NEW for-sale item (old is inactive) ===
	// but pending purchases should prevent over-listing
	for cycle := 2; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)

		pendingCount := 0
		totalPendingQty := int64(0)
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
				totalPendingQty += p.QuantityPurchased
			}
		}

		// Dump state
		var fsTotalCount, fsActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM for_sale_items`).Scan(&fsTotalCount, &fsActiveCount)
		require.NoError(t, err)

		var boTotalCount, boActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM buy_orders`).Scan(&boTotalCount, &boActiveCount)
		require.NoError(t, err)

		t.Logf("Cycle %d: %d purchases (%d pending, qty=%d) | for_sale: %d total/%d active | buy_orders: %d total/%d active",
			cycle, len(purchases), pendingCount, totalPendingQty, fsTotalCount, fsActiveCount, boTotalCount, boActiveCount)
		for _, p := range purchases {
			t.Logf("  purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d",
				p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID)
		}

		assert.Equal(t, 1, pendingCount,
			"Cycle %d: expected 1 pending purchase, got %d", cycle, pendingCount)
		assert.Equal(t, int64(100), totalPendingQty,
			"Cycle %d: total pending qty should be 100, got %d", cycle, totalPendingQty)
	}
}

// Test_AutoFulfill_CorporationAssetsWithContainer verifies the full pipeline with
// corporation assets stored in a container inside a corp hangar division.
// This matches the production setup where:
//   - Seller has items in a container within a corp hangar (CorpSAG1)
//   - Auto-sell is configured with ContainerID + DivisionNumber
//   - Auto-buy + auto-fulfill create purchases from these listings
func Test_AutoFulfill_CorporationAssetsWithContainer(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	// --- Users ---
	buyerUserID := int64(8150)
	sellerUserID := int64(8160)
	buyerCharID := int64(81501)
	sellerCharID := int64(81601)
	sellerCorpID := int64(98001)
	buyerCorpID := int64(98002)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "CorpBuyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: sellerUserID, Name: "CorpSeller"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "CorpBuyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: sellerCharID, Name: "CorpSeller Char", UserID: sellerUserID}))

	// --- Player Corporations ---
	corpRepo := repositories.NewPlayerCorporations(db)
	require.NoError(t, corpRepo.Upsert(ctx, repositories.PlayerCorporation{
		ID: sellerCorpID, UserID: sellerUserID, Name: "Seller Corp",
		EsiToken: "t", EsiRefreshToken: "r", EsiExpiresOn: require_time(t), EsiScopes: "s",
	}))
	require.NoError(t, corpRepo.Upsert(ctx, repositories.PlayerCorporation{
		ID: buyerCorpID, UserID: buyerUserID, Name: "Buyer Corp",
		EsiToken: "t", EsiRefreshToken: "r", EsiExpiresOn: require_time(t), EsiScopes: "s",
	}))

	// --- Permissions ---
	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, sellerUserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: sellerUserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: sellerUserID, ServiceType: "for_sale_browse", CanAccess: true}))

	// --- Seller: Corp assets in container within corp hangar division ---
	// Structure: Station(60003760) → Office(officeItemID) → CorpSAG1 → Container(containerItemID) → Items
	stationID := int64(60003760)
	officeItemID := int64(9000001)
	containerItemID := int64(9000002)
	tritaniumItemID := int64(9000003)

	corpAssetsRepo := repositories.NewCorporationAssets(db)
	require.NoError(t, corpAssetsRepo.Upsert(ctx, sellerCorpID, sellerUserID, []*models.EveAsset{
		// The corp office at the station
		{ItemID: officeItemID, LocationID: stationID, LocationType: "station", Quantity: 1, TypeID: 27, LocationFlag: "OfficeFolder"},
		// The container in CorpSAG1
		{ItemID: containerItemID, LocationID: officeItemID, LocationType: "item", Quantity: 1, TypeID: 17366, LocationFlag: "CorpSAG1"},
		// Tritanium inside the container
		{ItemID: tritaniumItemID, LocationID: containerItemID, LocationType: "item", Quantity: 100, TypeID: 34, LocationFlag: "CorpSAG1"},
	}))

	// --- Auto-sell container: selling from corp container ---
	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: sellerUserID, OwnerType: "corporation", OwnerID: sellerCorpID,
		LocationID: stationID, ContainerID: &containerItemID, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	// --- Buyer: stockpile in corp hangar wants 100 Tritanium ---
	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "corporation", OwnerID: buyerCorpID,
		LocationID: stationID, DesiredQuantity: 100,
	}))

	// --- Auto-buy config for buyer's corp ---
	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "corporation", OwnerID: buyerCorpID,
		LocationID: stationID, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	// --- Market prices ---
	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	// --- Create updaters ---
	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === Run 5 cycles of the full pipeline ===
	for cycle := 1; cycle <= 5; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx), "cycle %d: auto-sell failed", cycle)
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx), "cycle %d: auto-buy failed", cycle)
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx), "cycle %d: auto-fulfill failed", cycle)

		purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)

		pendingCount := 0
		totalPendingQty := int64(0)
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
				totalPendingQty += p.QuantityPurchased
			}
		}

		// Dump state
		var fsTotalCount, fsActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM for_sale_items`).Scan(&fsTotalCount, &fsActiveCount)
		require.NoError(t, err)

		var boTotalCount, boActiveCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active) FROM buy_orders`).Scan(&boTotalCount, &boActiveCount)
		require.NoError(t, err)

		t.Logf("Cycle %d: %d purchases (%d pending, qty=%d) | for_sale: %d total/%d active | buy_orders: %d total/%d active",
			cycle, len(purchases), pendingCount, totalPendingQty, fsTotalCount, fsActiveCount, boTotalCount, boActiveCount)
		for _, p := range purchases {
			t.Logf("  purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d auto=%v",
				p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID, p.IsAutoFulfilled)
		}

		if cycle == 1 {
			require.Equal(t, 1, pendingCount, "Cycle 1: expected exactly 1 purchase")
			assert.Equal(t, int64(100), purchases[0].QuantityPurchased)
		} else {
			assert.Equal(t, 1, pendingCount,
				"Cycle %d: expected still 1 pending purchase, got %d", cycle, pendingCount)
			assert.LessOrEqual(t, totalPendingQty, int64(100),
				"Cycle %d: total pending qty should not exceed 100", cycle)
		}
	}
}

// Test_AutoFulfill_GrowingDeficit_NoBuyOrderIDCycling verifies the production scenario
// where the buyer's deficit grows between cycles (buyer consumes items). Previously,
// auto-buy would deactivate the buy order when pending covered the deficit, then
// recreate it with a NEW ID when deficit grew — resetting the pending count in
// auto-fulfill and causing duplicate purchases.
//
// With the fix, auto-buy uses the raw deficit (no pending subtraction), keeping buy
// order IDs stable. Auto-fulfill handles pending tracking via GetPendingQuantityForBuyOrder.
//
// Two sellers are used so that when the deficit grows, the second purchase comes from
// a different for_sale_item (the unique index correctly prevents two pending purchases
// for the same buy_order + for_sale_item pair).
func Test_AutoFulfill_GrowingDeficit_NoBuyOrderIDCycling(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	setupTestUniverse(t, db)

	ctx := context.Background()

	buyerUserID := int64(8200)
	seller1UserID := int64(8210)
	seller2UserID := int64(8220)
	buyerCharID := int64(82001)
	seller1CharID := int64(82101)
	seller2CharID := int64(82201)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)

	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: buyerUserID, Name: "GrowBuyer"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: seller1UserID, Name: "GrowSeller1"}))
	require.NoError(t, userRepo.Add(ctx, &repositories.User{ID: seller2UserID, Name: "GrowSeller2"}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: buyerCharID, Name: "GrowBuyer Char", UserID: buyerUserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: seller1CharID, Name: "GrowSeller1 Char", UserID: seller1UserID}))
	require.NoError(t, charRepo.Add(ctx, &repositories.Character{ID: seller2CharID, Name: "GrowSeller2 Char", UserID: seller2UserID}))

	// Contact + permissions: buyer ↔ seller1
	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, seller1UserID)
	require.NoError(t, err)
	// Contact + permissions: buyer ↔ seller2
	_, err = db.ExecContext(ctx, `INSERT INTO contacts (requester_user_id, recipient_user_id, status) VALUES ($1, $2, 'accepted')`, buyerUserID, seller2UserID)
	require.NoError(t, err)

	permRepo := repositories.NewContactPermissions(db)
	// buyer ↔ seller1 mutual permissions
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: seller1UserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 1, GrantingUserID: buyerUserID, ReceivingUserID: seller1UserID, ServiceType: "for_sale_browse", CanAccess: true}))
	// buyer ↔ seller2 mutual permissions
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 2, GrantingUserID: seller2UserID, ReceivingUserID: buyerUserID, ServiceType: "for_sale_browse", CanAccess: true}))
	require.NoError(t, permRepo.Upsert(ctx, &models.ContactPermission{ContactID: 2, GrantingUserID: buyerUserID, ReceivingUserID: seller2UserID, ServiceType: "for_sale_browse", CanAccess: true}))

	// Seller1 has 100 Tritanium, Seller2 has 100 Tritanium
	charAssetsRepo := repositories.NewCharacterAssets(db)
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, seller1CharID, seller1UserID, []*models.EveAsset{
		{ItemID: 2000001, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, seller2CharID, seller2UserID, []*models.EveAsset{
		{ItemID: 2000002, LocationID: 60003760, LocationType: "other", Quantity: 100, TypeID: 34, LocationFlag: "Hangar"},
	}))

	autoSellRepo := repositories.NewAutoSellContainers(db)
	jitaSellPrice := 5.0
	divisionNum := 1
	// Seller1 auto-sell config
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: seller1UserID, OwnerType: "character", OwnerID: seller1CharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))
	// Seller2 auto-sell config
	require.NoError(t, autoSellRepo.Upsert(ctx, &models.AutoSellContainer{
		UserID: seller2UserID, OwnerType: "character", OwnerID: seller2CharID,
		LocationID: 60003760, DivisionNumber: &divisionNum,
		PricePercentage: 100.0, PriceSource: "jita_sell", IsActive: true,
	}))

	// Buyer's stockpile wants 100 Tritanium (starts with 0 in inventory)
	stockpileRepo := repositories.NewStockpileMarkers(db)
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 100,
	}))

	// Buyer has 0 Tritanium initially → deficit = 100
	require.NoError(t, charAssetsRepo.UpdateAssets(ctx, buyerCharID, buyerUserID, []*models.EveAsset{}))

	autoBuyConfigsRepo := repositories.NewAutoBuyConfigs(db)
	require.NoError(t, autoBuyConfigsRepo.Upsert(ctx, &models.AutoBuyConfig{
		UserID: buyerUserID, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, MinPricePercentage: 50.0, MaxPricePercentage: 200.0,
		PriceSource: "jita_sell", IsActive: true,
	}))

	marketRepo := repositories.NewMarketPrices(db)
	require.NoError(t, marketRepo.UpsertPrices(ctx, []models.MarketPrice{
		{TypeID: 34, RegionID: 10000002, BuyPrice: &jitaSellPrice, SellPrice: &jitaSellPrice},
	}))

	forSaleRepo := repositories.NewForSaleItems(db)
	buyOrdersRepo := repositories.NewBuyOrders(db)
	purchaseRepo := repositories.NewPurchaseTransactions(db)

	autoSellUpdater := updaters.NewAutoSell(autoSellRepo, forSaleRepo, marketRepo, stockpileRepo, purchaseRepo)
	autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepo, buyOrdersRepo, marketRepo, purchaseRepo)
	autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepo, forSaleRepo, purchaseRepo, permRepo, userRepo, nil)

	// === CYCLE 1: deficit=100, creates purchase for 100 from seller1 ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	purchases, err := purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	require.Len(t, purchases, 1, "Cycle 1: expected exactly 1 purchase")
	assert.Equal(t, int64(100), purchases[0].QuantityPurchased)

	// Record the buy order ID — should stay stable
	var buyOrderID int64
	err = db.QueryRowContext(ctx, `SELECT id FROM buy_orders WHERE is_active = true AND buyer_user_id = $1`, buyerUserID).Scan(&buyOrderID)
	require.NoError(t, err)
	t.Logf("Cycle 1: buy_order_id=%d, purchase qty=100", buyOrderID)

	// === CYCLES 2-3: deficit stays at 100 (items not delivered), no new purchases ===
	for cycle := 2; cycle <= 3; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		// Verify buy order ID is stable (not recreated)
		var currentBuyOrderID int64
		err = db.QueryRowContext(ctx, `SELECT id FROM buy_orders WHERE is_active = true AND buyer_user_id = $1`, buyerUserID).Scan(&currentBuyOrderID)
		require.NoError(t, err)
		assert.Equal(t, buyOrderID, currentBuyOrderID, "Cycle %d: buy order ID should be stable", cycle)

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount := 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
			}
		}
		assert.Equal(t, 1, pendingCount, "Cycle %d: expected still 1 pending purchase", cycle)
	}

	// === Simulate deficit growing: buyer consumes items (stockpile wants 200 now) ===
	require.NoError(t, stockpileRepo.Upsert(ctx, &models.StockpileMarker{
		UserID: buyerUserID, TypeID: 34, OwnerType: "character", OwnerID: buyerCharID,
		LocationID: 60003760, DesiredQuantity: 200,
	}))

	// === CYCLE 4: deficit grew to 200, should create 1 more purchase (100 from seller2) ===
	require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
	require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

	// Verify buy order ID is STILL stable
	var cycle4BuyOrderID int64
	err = db.QueryRowContext(ctx, `SELECT id FROM buy_orders WHERE is_active = true AND buyer_user_id = $1`, buyerUserID).Scan(&cycle4BuyOrderID)
	require.NoError(t, err)
	assert.Equal(t, buyOrderID, cycle4BuyOrderID, "Cycle 4: buy order ID should still be stable after deficit growth")

	purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
	require.NoError(t, err)
	pendingCount := 0
	totalPendingQty := int64(0)
	for _, p := range purchases {
		if p.Status == "pending" || p.Status == "contract_created" {
			pendingCount++
			totalPendingQty += p.QuantityPurchased
		}
		t.Logf("Cycle 4: purchase id=%d qty=%d status=%s buyOrder=%v forSaleItem=%d",
			p.ID, p.QuantityPurchased, p.Status, p.BuyOrderID, p.ForSaleItemID)
	}
	assert.Equal(t, 2, pendingCount, "Cycle 4: expected exactly 2 pending purchases (100 from each seller)")
	assert.Equal(t, int64(200), totalPendingQty, "Cycle 4: total pending qty should be 200")

	// === CYCLES 5-7: no more growth, should stay at 2 pending purchases ===
	for cycle := 5; cycle <= 7; cycle++ {
		require.NoError(t, autoSellUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoBuyUpdater.SyncForAllUsers(ctx))
		require.NoError(t, autoFulfillUpdater.SyncForAllUsers(ctx))

		// Verify buy order ID stability
		var currentID int64
		err = db.QueryRowContext(ctx, `SELECT id FROM buy_orders WHERE is_active = true AND buyer_user_id = $1`, buyerUserID).Scan(&currentID)
		require.NoError(t, err)
		assert.Equal(t, buyOrderID, currentID, "Cycle %d: buy order ID should be stable", cycle)

		purchases, err = purchaseRepo.GetByBuyer(ctx, buyerUserID)
		require.NoError(t, err)
		pendingCount = 0
		totalPendingQty = 0
		for _, p := range purchases {
			if p.Status == "pending" || p.Status == "contract_created" {
				pendingCount++
				totalPendingQty += p.QuantityPurchased
			}
		}
		assert.Equal(t, 2, pendingCount, "Cycle %d: expected still 2 pending purchases", cycle)
		assert.Equal(t, int64(200), totalPendingQty, "Cycle %d: total pending qty should still be 200", cycle)
	}
}

// require_time returns a time.Time for use in test setup
func require_time(t *testing.T) time.Time {
	t.Helper()
	return time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
}
