package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type AutoSellContainersRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.AutoSellContainer, error)
	Upsert(ctx context.Context, container *models.AutoSellContainer) error
	Delete(ctx context.Context, id int64, userID int64) error
	GetByID(ctx context.Context, id int64) (*models.AutoSellContainer, error)
}

type AutoSellSyncer interface {
	SyncForUser(ctx context.Context, userID int64) error
}

type ForSaleItemsDeactivator interface {
	DeactivateAutoSellListings(ctx context.Context, autoSellContainerID int64) error
}

var allowedPriceSources = map[string]bool{
	"jita_buy":   true,
	"jita_sell":  true,
	"jita_split": true,
}

type AutoSellContainers struct {
	repository     AutoSellContainersRepository
	syncer         AutoSellSyncer
	forSaleDeactor ForSaleItemsDeactivator
}

func NewAutoSellContainers(
	router Routerer,
	repository AutoSellContainersRepository,
	syncer AutoSellSyncer,
	forSaleDeactor ForSaleItemsDeactivator,
) *AutoSellContainers {
	controller := &AutoSellContainers{
		repository:     repository,
		syncer:         syncer,
		forSaleDeactor: forSaleDeactor,
	}

	router.RegisterRestAPIRoute("/v1/auto-sell", web.AuthAccessUser, controller.GetMyConfigs, "GET")
	router.RegisterRestAPIRoute("/v1/auto-sell", web.AuthAccessUser, controller.CreateConfig, "POST")
	router.RegisterRestAPIRoute("/v1/auto-sell/{id}", web.AuthAccessUser, controller.UpdateConfig, "PUT")
	router.RegisterRestAPIRoute("/v1/auto-sell/{id}", web.AuthAccessUser, controller.DeleteConfig, "DELETE")

	return controller
}

// GetMyConfigs returns all active auto-sell containers for the authenticated user
func (c *AutoSellContainers) GetMyConfigs(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	items, err := c.repository.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get auto-sell containers")}
	}

	return items, nil
}

// CreateConfig creates a new auto-sell container configuration
func (c *AutoSellContainers) CreateConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	var req struct {
		OwnerType       string  `json:"ownerType"`
		OwnerID         int64   `json:"ownerId"`
		LocationID      int64   `json:"locationId"`
		ContainerID     *int64  `json:"containerId"`
		DivisionNumber  *int    `json:"divisionNumber"`
		PricePercentage float64 `json:"pricePercentage"`
		PriceSource     string  `json:"priceSource"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
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
	if req.ContainerID == nil && req.DivisionNumber == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("either containerId or divisionNumber is required")}
	}
	if req.PricePercentage <= 0 || req.PricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricePercentage must be between 0 and 200")}
	}

	if req.PriceSource == "" {
		req.PriceSource = "jita_buy"
	}
	if !allowedPriceSources[req.PriceSource] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Errorf("invalid priceSource: %s", req.PriceSource)}
	}

	container := &models.AutoSellContainer{
		UserID:          *args.User,
		OwnerType:       req.OwnerType,
		OwnerID:         req.OwnerID,
		LocationID:      req.LocationID,
		ContainerID:     req.ContainerID,
		DivisionNumber:  req.DivisionNumber,
		PricePercentage: req.PricePercentage,
		PriceSource:     req.PriceSource,
	}

	if err := c.repository.Upsert(args.Request.Context(), container); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create auto-sell config")}
	}

	// Trigger immediate sync
	go func() {
		ctx := context.Background()
		_ = c.syncer.SyncForUser(ctx, container.UserID)
	}()

	return container, nil
}

// UpdateConfig updates an existing auto-sell container configuration
func (c *AutoSellContainers) UpdateConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	idStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("id is required")}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid id")}
	}

	existing, err := c.repository.GetByID(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get auto-sell container")}
	}
	if existing == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("auto-sell container not found")}
	}
	if existing.UserID != *args.User {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("not authorized to update this config")}
	}

	var req struct {
		PricePercentage float64 `json:"pricePercentage"`
		PriceSource     string  `json:"priceSource"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.PricePercentage <= 0 || req.PricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("pricePercentage must be between 0 and 200")}
	}

	if req.PriceSource != "" {
		if !allowedPriceSources[req.PriceSource] {
			return nil, &web.HttpError{StatusCode: 400, Error: errors.Errorf("invalid priceSource: %s", req.PriceSource)}
		}
		existing.PriceSource = req.PriceSource
	}

	existing.PricePercentage = req.PricePercentage
	if err := c.repository.Upsert(args.Request.Context(), existing); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update auto-sell config")}
	}

	// Trigger immediate sync
	go func() {
		ctx := context.Background()
		_ = c.syncer.SyncForUser(ctx, existing.UserID)
	}()

	return existing, nil
}

// DeleteConfig soft-deletes an auto-sell container and deactivates all associated listings
func (c *AutoSellContainers) DeleteConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	idStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("id is required")}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid id")}
	}

	// Deactivate all associated for-sale listings first
	if err := c.forSaleDeactor.DeactivateAutoSellListings(args.Request.Context(), id); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to deactivate auto-sell listings")}
	}

	if err := c.repository.Delete(args.Request.Context(), id, *args.User); err != nil {
		if err.Error() == "auto-sell container not found or user is not the owner" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete auto-sell config")}
	}

	return nil, nil
}
