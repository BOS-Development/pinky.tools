package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ForSaleItemsRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.ForSaleItem, error)
	GetBrowsableItems(ctx context.Context, buyerUserID int64, sellerUserIDs []int64) ([]*models.ForSaleItem, error)
	Upsert(ctx context.Context, item *models.ForSaleItem) error
	Delete(ctx context.Context, itemID int64, userID int64) error
	GetByID(ctx context.Context, itemID int64) (*models.ForSaleItem, error)
	GetUserIDByCharacterID(ctx context.Context, characterID int64) (int64, error)
}

type ForSaleItems struct {
	repository            ForSaleItemsRepository
	permissionsRepository ContactPermissionsRepository
}

func NewForSaleItems(router Routerer, repository ForSaleItemsRepository, permissionsRepository ContactPermissionsRepository) *ForSaleItems {
	controller := &ForSaleItems{
		repository:            repository,
		permissionsRepository: permissionsRepository,
	}

	router.RegisterRestAPIRoute("/v1/for-sale", web.AuthAccessUser, controller.GetMyListings, "GET")
	router.RegisterRestAPIRoute("/v1/for-sale/browse", web.AuthAccessUser, controller.BrowseListings, "GET")
	router.RegisterRestAPIRoute("/v1/for-sale", web.AuthAccessUser, controller.CreateListing, "POST")
	router.RegisterRestAPIRoute("/v1/for-sale/{id}", web.AuthAccessUser, controller.UpdateListing, "PUT")
	router.RegisterRestAPIRoute("/v1/for-sale/{id}", web.AuthAccessUser, controller.DeleteListing, "DELETE")

	return controller
}

// getUserID converts the session user (which is the main character's EVE ID) to a user ID
// Since users.id IS the main character's EVE character ID, we can use it directly
func (c *ForSaleItems) getUserID(ctx context.Context, sessionUser int64) (int64, error) {
	// The session user is the main character's EVE ID, which is also the user ID
	return sessionUser, nil
}

// GetMyListings returns all active for-sale items owned by the authenticated user
func (c *ForSaleItems) GetMyListings(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	items, err := c.repository.GetByUser(args.Request.Context(), userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get for-sale items")}
	}

	return items, nil
}

// BrowseListings returns for-sale items from contacts who granted browse permission
func (c *ForSaleItems) BrowseListings(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	// Get list of users who granted this user permission to browse their for-sale items
	sellerUserIDs, err := c.permissionsRepository.GetUserPermissionsForService(args.Request.Context(), userID, "for_sale_browse")
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get permissions")}
	}

	// Get browsable items from those sellers
	items, err := c.repository.GetBrowsableItems(args.Request.Context(), userID, sellerUserIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get browsable items")}
	}

	return items, nil
}

// CreateListing creates a new for-sale item listing
func (c *ForSaleItems) CreateListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	var req struct {
		TypeID            int64   `json:"typeId"`
		OwnerType         string  `json:"ownerType"`
		OwnerID           int64   `json:"ownerId"`
		LocationID        int64   `json:"locationId"`
		ContainerID       *int64  `json:"containerId"`
		DivisionNumber    *int    `json:"divisionNumber"`
		QuantityAvailable int64   `json:"quantityAvailable"`
		PricePerUnit      float64 `json:"pricePerUnit"`
		Notes             *string `json:"notes"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate required fields
	if req.TypeID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("typeId is required")}
	}
	if req.OwnerType == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("ownerType is required")}
	}
	if req.OwnerID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("ownerId is required")}
	}
	if req.LocationID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("locationId is required")}
	}
	if req.QuantityAvailable <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("quantityAvailable must be greater than 0")}
	}
	if req.PricePerUnit < 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricePerUnit must be non-negative")}
	}

	item := &models.ForSaleItem{
		UserID:            userID,
		TypeID:            req.TypeID,
		OwnerType:         req.OwnerType,
		OwnerID:           req.OwnerID,
		LocationID:        req.LocationID,
		ContainerID:       req.ContainerID,
		DivisionNumber:    req.DivisionNumber,
		QuantityAvailable: req.QuantityAvailable,
		PricePerUnit:      req.PricePerUnit,
		Notes:             req.Notes,
		IsActive:          true,
	}

	if err := c.repository.Upsert(args.Request.Context(), item); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create listing")}
	}

	return item, nil
}

// UpdateListing updates an existing for-sale item listing
func (c *ForSaleItems) UpdateListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	itemIDStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("item ID is required")}
	}

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid item ID")}
	}

	// Get existing item to verify ownership
	existingItem, err := c.repository.GetByID(args.Request.Context(), itemID)
	if err != nil {
		if err == sql.ErrNoRows || errors.Cause(err).Error() == "for-sale item not found" {
			return nil, &web.HttpError{StatusCode: 404, Error: errors.New("for-sale item not found")}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get for-sale item")}
	}

	if existingItem.UserID != userID {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("not authorized to update this listing")}
	}

	var req struct {
		QuantityAvailable int64   `json:"quantityAvailable"`
		PricePerUnit      float64 `json:"pricePerUnit"`
		Notes             *string `json:"notes"`
		IsActive          *bool   `json:"isActive"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate updates
	if req.QuantityAvailable <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("quantityAvailable must be greater than 0")}
	}
	if req.PricePerUnit < 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricePerUnit must be non-negative")}
	}

	// Update fields
	existingItem.QuantityAvailable = req.QuantityAvailable
	existingItem.PricePerUnit = req.PricePerUnit
	existingItem.Notes = req.Notes
	if req.IsActive != nil {
		existingItem.IsActive = *req.IsActive
	}

	if err := c.repository.Upsert(args.Request.Context(), existingItem); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update listing")}
	}

	return existingItem, nil
}

// DeleteListing soft-deletes a for-sale item listing
func (c *ForSaleItems) DeleteListing(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	userID, err := c.getUserID(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user ID")}
	}

	itemIDStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("item ID is required")}
	}

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid item ID")}
	}

	if err := c.repository.Delete(args.Request.Context(), itemID, userID); err != nil {
		if err.Error() == "for-sale item not found or user is not the owner" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete listing")}
	}

	return nil, nil
}
