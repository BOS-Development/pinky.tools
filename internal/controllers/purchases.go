package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type PurchaseTransactionsRepository interface {
	Create(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error
	GetByBuyer(ctx context.Context, buyerUserID int64) ([]*models.PurchaseTransaction, error)
	GetBySeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error)
	GetPendingForSeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error)
	GetByID(ctx context.Context, purchaseID int64) (*models.PurchaseTransaction, error)
	UpdateStatus(ctx context.Context, purchaseID int64, newStatus string) error
	UpdateContractKeys(ctx context.Context, purchaseIDs []int64, contractKey string) error
}

type ForSaleItemsForPurchases interface {
	GetByID(ctx context.Context, itemID int64) (*models.ForSaleItem, error)
	UpdateQuantity(ctx context.Context, tx *sql.Tx, itemID int64, newQuantity int64) error
}

type PurchaseNotifierInterface interface {
	NotifyPurchase(ctx context.Context, purchase *models.PurchaseTransaction)
}

type UsersForPurchases interface {
	GetUserName(ctx context.Context, userID int64) (string, error)
}

type Purchases struct {
	db                     *sql.DB
	repository             PurchaseTransactionsRepository
	forSaleRepository      ForSaleItemsForPurchases
	permissionsRepository  ContactPermissionsRepository
	usersRepository        UsersForPurchases
	notifier               PurchaseNotifierInterface
}

func NewPurchases(router Routerer, db *sql.DB, repository PurchaseTransactionsRepository, forSaleRepository ForSaleItemsForPurchases, permissionsRepository ContactPermissionsRepository, usersRepository UsersForPurchases, notifier PurchaseNotifierInterface) *Purchases {
	controller := &Purchases{
		db:                    db,
		repository:            repository,
		forSaleRepository:     forSaleRepository,
		permissionsRepository: permissionsRepository,
		usersRepository:       usersRepository,
		notifier:              notifier,
	}

	router.RegisterRestAPIRoute("/v1/purchases", web.AuthAccessUser, controller.PurchaseItem, "POST")
	router.RegisterRestAPIRoute("/v1/purchases/buyer", web.AuthAccessUser, controller.GetBuyerHistory, "GET")
	router.RegisterRestAPIRoute("/v1/purchases/seller", web.AuthAccessUser, controller.GetSellerHistory, "GET")
	router.RegisterRestAPIRoute("/v1/purchases/pending-sales", web.AuthAccessUser, controller.GetPendingSales, "GET")
	router.RegisterRestAPIRoute("/v1/purchases/{id}/mark-contract-created", web.AuthAccessUser, controller.MarkContractCreated, "POST")
	router.RegisterRestAPIRoute("/v1/purchases/{id}/complete", web.AuthAccessUser, controller.CompletePurchase, "POST")
	router.RegisterRestAPIRoute("/v1/purchases/{id}/cancel", web.AuthAccessUser, controller.CancelPurchase, "POST")

	return controller
}

// getUserID returns the session user directly (session user IS the user ID)
func (c *Purchases) getUserID(ctx context.Context, sessionUser int64) (int64, error) {
	return sessionUser, nil
}

type PurchaseRequest struct {
	ForSaleItemID     int64  `json:"forSaleItemId"`
	QuantityPurchased int64  `json:"quantityPurchased"`
	Notes             string `json:"notes,omitempty"`
}

// PurchaseItem handles purchasing an item from the marketplace
func (c *Purchases) PurchaseItem(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	buyerUserID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	var req PurchaseRequest
	err = json.NewDecoder(args.Request.Body).Decode(&req)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate request
	if req.QuantityPurchased <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("quantity must be positive")}
	}

	// 1. Get for-sale item
	item, err := c.forSaleRepository.GetByID(args.Request.Context(), req.ForSaleItemID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.Wrap(err, "for-sale item not found")}
	}

	// 2. Verify buyer has permission to browse from this seller
	hasPermission, err := c.permissionsRepository.CheckPermission(args.Request.Context(), item.UserID, buyerUserID, "for_sale_browse")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to check permission")}
	}

	if !hasPermission {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("you do not have permission to purchase from this seller")}
	}

	// 3. Prevent self-purchase
	if buyerUserID == item.UserID {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("cannot purchase your own items")}
	}

	// 4. Validate quantity
	if req.QuantityPurchased > item.QuantityAvailable {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("requested quantity exceeds available quantity")}
	}

	// 5. Begin transaction
	tx, err := c.db.BeginTx(args.Request.Context(), nil)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to begin transaction")}
	}
	defer tx.Rollback()

	// 6. Update quantity
	newQuantity := item.QuantityAvailable - req.QuantityPurchased
	err = c.forSaleRepository.UpdateQuantity(args.Request.Context(), tx, item.ID, newQuantity)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update quantity")}
	}

	// 7. Create purchase record
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	purchase := &models.PurchaseTransaction{
		ForSaleItemID:     item.ID,
		BuyerUserID:       buyerUserID,
		SellerUserID:      item.UserID,
		TypeID:            item.TypeID,
		QuantityPurchased: req.QuantityPurchased,
		PricePerUnit:      item.PricePerUnit,
		TotalPrice:        float64(req.QuantityPurchased) * item.PricePerUnit,
		Status:            "pending",
		TransactionNotes:  notes,
	}

	err = c.repository.Create(args.Request.Context(), tx, purchase)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create purchase transaction")}
	}

	// 8. Commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to commit transaction")}
	}

	// 9. Populate fields for notification display
	purchase.TypeName = item.TypeName
	purchase.LocationName = item.LocationName
	purchase.LocationID = item.LocationID
	if buyerName, err := c.usersRepository.GetUserName(args.Request.Context(), buyerUserID); err == nil {
		purchase.BuyerName = buyerName
	}

	// 10. Send notifications (non-blocking, never fails the purchase)
	if c.notifier != nil {
		go c.notifier.NotifyPurchase(context.Background(), purchase)
	}

	return purchase, nil
}

// GetBuyerHistory returns purchase history for the authenticated buyer
func (c *Purchases) GetBuyerHistory(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	transactions, err := c.repository.GetByBuyer(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get buyer history")}
	}

	return transactions, nil
}

// GetSellerHistory returns sales history for the authenticated seller
func (c *Purchases) GetSellerHistory(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	transactions, err := c.repository.GetBySeller(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get seller history")}
	}

	return transactions, nil
}

// GetPendingSales returns pending purchase requests for the authenticated seller
func (c *Purchases) GetPendingSales(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	transactions, err := c.repository.GetPendingForSeller(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get pending sales")}
	}

	return transactions, nil
}

type MarkContractCreatedRequest struct {
	ContractKey *string `json:"contractKey,omitempty"`
}

// MarkContractCreated marks a purchase as contract_created (seller action)
func (c *Purchases) MarkContractCreated(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	purchaseIDStr := args.Params["id"]
	purchaseID, err := strconv.ParseInt(purchaseIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid purchase ID")}
	}

	// Parse request body for optional contract key
	var req MarkContractCreatedRequest
	if args.Request.Body != nil {
		err = json.NewDecoder(args.Request.Body).Decode(&req)
		if err != nil && err.Error() != "EOF" {
			return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
		}
	}

	// Get purchase to verify seller
	purchase, err := c.repository.GetByID(args.Request.Context(), purchaseID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.Wrap(err, "purchase not found")}
	}

	// Verify user is the seller
	if purchase.SellerUserID != userID {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("you are not the seller of this purchase")}
	}

	// Verify status is pending
	if purchase.Status != "pending" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("purchase must be in pending status")}
	}

	// Update status
	err = c.repository.UpdateStatus(args.Request.Context(), purchaseID, "contract_created")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update purchase status")}
	}

	// Auto-generate contract_key if not provided
	contractKey := fmt.Sprintf("PT-%d", purchaseID)
	if req.ContractKey != nil && *req.ContractKey != "" {
		contractKey = *req.ContractKey
	}

	err = c.repository.UpdateContractKeys(args.Request.Context(), []int64{purchaseID}, contractKey)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update contract key")}
	}

	return map[string]string{"status": "contract_created", "contractKey": contractKey}, nil
}

// CompletePurchase marks a purchase as completed (buyer action)
func (c *Purchases) CompletePurchase(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	purchaseIDStr := args.Params["id"]
	purchaseID, err := strconv.ParseInt(purchaseIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid purchase ID")}
	}

	// Get purchase to verify buyer
	purchase, err := c.repository.GetByID(args.Request.Context(), purchaseID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.Wrap(err, "purchase not found")}
	}

	// Verify user is the buyer
	if purchase.BuyerUserID != userID {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("you are not the buyer of this purchase")}
	}

	// Verify status is contract_created
	if purchase.Status != "contract_created" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("purchase must be in contract_created status")}
	}

	// Update status
	err = c.repository.UpdateStatus(args.Request.Context(), purchaseID, "completed")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update purchase status")}
	}

	return map[string]string{"status": "completed"}, nil
}

// CancelPurchase cancels a purchase (either party can cancel)
func (c *Purchases) CancelPurchase(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	purchaseIDStr := args.Params["id"]
	purchaseID, err := strconv.ParseInt(purchaseIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid purchase ID")}
	}

	// Get purchase to verify user is involved
	purchase, err := c.repository.GetByID(args.Request.Context(), purchaseID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.Wrap(err, "purchase not found")}
	}

	// Verify user is either buyer or seller
	if purchase.BuyerUserID != userID && purchase.SellerUserID != userID {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("you are not involved in this purchase")}
	}

	// Verify status is not already completed
	if purchase.Status == "completed" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("cannot cancel completed purchase")}
	}

	// If cancelling, restore the quantity to the for-sale item
	tx, err := c.db.BeginTx(args.Request.Context(), nil)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to begin transaction")}
	}
	defer tx.Rollback()

	// Get current for-sale item
	item, err := c.forSaleRepository.GetByID(args.Request.Context(), purchase.ForSaleItemID)
	if err == nil && item.IsActive {
		// Item still exists and is active, restore quantity
		newQuantity := item.QuantityAvailable + purchase.QuantityPurchased
		err = c.forSaleRepository.UpdateQuantity(args.Request.Context(), tx, item.ID, newQuantity)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to restore quantity")}
		}
	}
	// If item doesn't exist or is inactive, that's OK - just cancel the purchase

	// Update purchase status
	err = c.repository.UpdateStatus(args.Request.Context(), purchaseID, "cancelled")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to cancel purchase")}
	}

	err = tx.Commit()
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to commit transaction")}
	}

	return map[string]string{"status": "cancelled"}, nil
}
