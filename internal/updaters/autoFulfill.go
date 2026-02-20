package updaters

import (
	"context"
	"database/sql"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type AutoFulfillBuyOrdersRepository interface {
	GetAllActiveBuyOrders(ctx context.Context) ([]*models.BuyOrder, error)
	GetActiveBuyOrdersForUser(ctx context.Context, userID int64) ([]*models.BuyOrder, error)
}

type AutoFulfillForSaleRepository interface {
	GetMatchingForSaleItems(ctx context.Context, typeID int64, minPrice, maxPrice float64, excludeUserID int64) ([]*models.ForSaleItem, error)
	UpdateQuantity(ctx context.Context, tx *sql.Tx, itemID int64, newQuantity int64) error
	GetByID(ctx context.Context, itemID int64) (*models.ForSaleItem, error)
}

type AutoFulfillPurchaseRepository interface {
	CreateAutoFulfill(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error
}

type AutoFulfillPermissionsRepository interface {
	CheckPermission(ctx context.Context, grantingUserID, receivingUserID int64, serviceType string) (bool, error)
}

type AutoFulfillNotifier interface {
	NotifyPurchase(ctx context.Context, purchase *models.PurchaseTransaction)
}

type AutoFulfillUsersRepository interface {
	GetUserName(ctx context.Context, userID int64) (string, error)
}

type AutoFulfill struct {
	db              *sql.DB
	buyOrderRepo    AutoFulfillBuyOrdersRepository
	forSaleRepo     AutoFulfillForSaleRepository
	purchaseRepo    AutoFulfillPurchaseRepository
	permissionsRepo AutoFulfillPermissionsRepository
	usersRepo       AutoFulfillUsersRepository
	notifier        AutoFulfillNotifier
}

func NewAutoFulfill(
	db *sql.DB,
	buyOrderRepo AutoFulfillBuyOrdersRepository,
	forSaleRepo AutoFulfillForSaleRepository,
	purchaseRepo AutoFulfillPurchaseRepository,
	permissionsRepo AutoFulfillPermissionsRepository,
	usersRepo AutoFulfillUsersRepository,
	notifier AutoFulfillNotifier,
) *AutoFulfill {
	return &AutoFulfill{
		db:              db,
		buyOrderRepo:    buyOrderRepo,
		forSaleRepo:     forSaleRepo,
		purchaseRepo:    purchaseRepo,
		permissionsRepo: permissionsRepo,
		usersRepo:       usersRepo,
		notifier:        notifier,
	}
}

// SyncForUser matches buy orders for a specific user against available for-sale items
func (u *AutoFulfill) SyncForUser(ctx context.Context, userID int64) error {
	orders, err := u.buyOrderRepo.GetActiveBuyOrdersForUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get active buy orders for user")
	}

	if len(orders) == 0 {
		return nil
	}

	for _, order := range orders {
		if err := u.matchBuyOrder(ctx, order); err != nil {
			log.Error("failed to match buy order",
				"orderID", order.ID,
				"userID", order.BuyerUserID,
				"error", err)
		}
	}

	return nil
}

// SyncForAllUsers matches all active buy orders against available for-sale items
func (u *AutoFulfill) SyncForAllUsers(ctx context.Context) error {
	orders, err := u.buyOrderRepo.GetAllActiveBuyOrders(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all active buy orders")
	}

	if len(orders) == 0 {
		return nil
	}

	for _, order := range orders {
		if err := u.matchBuyOrder(ctx, order); err != nil {
			log.Error("failed to match buy order",
				"orderID", order.ID,
				"userID", order.BuyerUserID,
				"error", err)
		}
	}

	return nil
}

// matchBuyOrder finds for-sale items that match a buy order and creates purchase transactions
func (u *AutoFulfill) matchBuyOrder(ctx context.Context, order *models.BuyOrder) error {
	// Skip orders with no max price (can't match)
	if order.MaxPricePerUnit <= 0 {
		return nil
	}

	// Find for-sale items matching type + price range, excluding buyer's own items
	items, err := u.forSaleRepo.GetMatchingForSaleItems(ctx, order.TypeID, order.MinPricePerUnit, order.MaxPricePerUnit, order.BuyerUserID)
	if err != nil {
		return errors.Wrap(err, "failed to get matching for-sale items")
	}

	if len(items) == 0 {
		return nil
	}

	remainingQuantity := order.QuantityDesired

	for _, item := range items {
		if remainingQuantity <= 0 {
			break
		}

		// Check mutual for_sale_browse permissions
		// Seller must have granted buyer permission
		sellerGrantsBuyer, err := u.permissionsRepo.CheckPermission(ctx, item.UserID, order.BuyerUserID, "for_sale_browse")
		if err != nil {
			log.Error("failed to check seller→buyer permission",
				"sellerID", item.UserID, "buyerID", order.BuyerUserID, "error", err)
			continue
		}
		if !sellerGrantsBuyer {
			continue
		}

		// Buyer must have granted seller permission
		buyerGrantsSeller, err := u.permissionsRepo.CheckPermission(ctx, order.BuyerUserID, item.UserID, "for_sale_browse")
		if err != nil {
			log.Error("failed to check buyer→seller permission",
				"buyerID", order.BuyerUserID, "sellerID", item.UserID, "error", err)
			continue
		}
		if !buyerGrantsSeller {
			continue
		}

		// Compute quantity to purchase
		quantity := remainingQuantity
		if quantity > item.QuantityAvailable {
			quantity = item.QuantityAvailable
		}

		// Create the purchase atomically
		err = u.createAutoFulfillPurchase(ctx, order, item, quantity)
		if err != nil {
			// Unique constraint violation means duplicate — skip silently
			log.Error("failed to create auto-fulfill purchase",
				"orderID", order.ID, "itemID", item.ID, "error", err)
			continue
		}

		remainingQuantity -= quantity
	}

	return nil
}

// createAutoFulfillPurchase atomically creates a purchase transaction and reduces for-sale quantity
func (u *AutoFulfill) createAutoFulfillPurchase(ctx context.Context, order *models.BuyOrder, item *models.ForSaleItem, quantity int64) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Reduce for-sale item quantity
	newQuantity := item.QuantityAvailable - quantity
	err = u.forSaleRepo.UpdateQuantity(ctx, tx, item.ID, newQuantity)
	if err != nil {
		return errors.Wrap(err, "failed to update for-sale item quantity")
	}

	// Create purchase transaction
	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       order.BuyerUserID,
		SellerUserID:      item.UserID,
		TypeID:            order.TypeID,
		QuantityPurchased: quantity,
		PricePerUnit:      item.PricePerUnit,
		TotalPrice:        float64(quantity) * item.PricePerUnit,
		Status:            "pending",
		BuyOrderID:        &order.ID,
		IsAutoFulfilled:   true,
	}

	err = u.purchaseRepo.CreateAutoFulfill(ctx, tx, purchase)
	if err != nil {
		return errors.Wrap(err, "failed to create auto-fulfill purchase transaction")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit auto-fulfill transaction")
	}

	// Populate fields for notification
	purchase.TypeName = item.TypeName
	purchase.LocationName = item.LocationName
	purchase.LocationID = item.LocationID
	if u.usersRepo != nil {
		if buyerName, err := u.usersRepo.GetUserName(ctx, order.BuyerUserID); err == nil {
			purchase.BuyerName = buyerName
		}
	}

	// Send notification (non-blocking)
	if u.notifier != nil {
		go u.notifier.NotifyPurchase(context.Background(), purchase)
	}

	log.Info("auto-fulfill purchase created",
		"purchaseID", purchase.ID,
		"buyOrderID", order.ID,
		"forSaleItemID", item.ID,
		"typeID", order.TypeID,
		"quantity", quantity,
		"pricePerUnit", item.PricePerUnit,
		"buyerID", order.BuyerUserID,
		"sellerID", item.UserID)

	return nil
}
