package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type AutoBuyConfigsRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.AutoBuyConfig, error)
	Upsert(ctx context.Context, config *models.AutoBuyConfig) error
	Delete(ctx context.Context, id int64, userID int64) error
	GetByID(ctx context.Context, id int64) (*models.AutoBuyConfig, error)
}

type AutoBuyConfigsSyncer interface {
	SyncForUser(ctx context.Context, userID int64) error
}

type BuyOrdersDeactivator interface {
	DeactivateAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) error
}

type AutoBuyConfigsAutoFulfillSyncer interface {
	SyncForUser(ctx context.Context, userID int64) error
}

type AutoBuyConfigs struct {
	repository        AutoBuyConfigsRepository
	syncer            AutoBuyConfigsSyncer
	buyOrderDeactor   BuyOrdersDeactivator
	autoFulfillSyncer AutoBuyConfigsAutoFulfillSyncer
}

func NewAutoBuyConfigs(
	router Routerer,
	repository AutoBuyConfigsRepository,
	syncer AutoBuyConfigsSyncer,
	buyOrderDeactor BuyOrdersDeactivator,
	autoFulfillSyncer AutoBuyConfigsAutoFulfillSyncer,
) *AutoBuyConfigs {
	controller := &AutoBuyConfigs{
		repository:        repository,
		syncer:            syncer,
		buyOrderDeactor:   buyOrderDeactor,
		autoFulfillSyncer: autoFulfillSyncer,
	}

	router.RegisterRestAPIRoute("/v1/auto-buy", web.AuthAccessUser, controller.GetMyConfigs, "GET")
	router.RegisterRestAPIRoute("/v1/auto-buy", web.AuthAccessUser, controller.CreateConfig, "POST")
	router.RegisterRestAPIRoute("/v1/auto-buy/{id}", web.AuthAccessUser, controller.UpdateConfig, "PUT")
	router.RegisterRestAPIRoute("/v1/auto-buy/{id}", web.AuthAccessUser, controller.DeleteConfig, "DELETE")

	return controller
}

// GetMyConfigs returns all active auto-buy configs for the authenticated user
func (c *AutoBuyConfigs) GetMyConfigs(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	items, err := c.repository.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get auto-buy configs")}
	}

	return items, nil
}

// CreateConfig creates a new auto-buy configuration
func (c *AutoBuyConfigs) CreateConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	var req struct {
		OwnerType          string  `json:"ownerType"`
		OwnerID            int64   `json:"ownerId"`
		LocationID         int64   `json:"locationId"`
		ContainerID        *int64  `json:"containerId"`
		DivisionNumber     *int    `json:"divisionNumber"`
		MinPricePercentage float64 `json:"minPricePercentage"`
		MaxPricePercentage float64 `json:"maxPricePercentage"`
		PriceSource        string  `json:"priceSource"`
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
	if req.MinPricePercentage < 0 || req.MinPricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("minPricePercentage must be between 0 and 200")}
	}
	if req.MaxPricePercentage <= 0 || req.MaxPricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("maxPricePercentage must be between 0 and 200")}
	}
	if req.MinPricePercentage > req.MaxPricePercentage {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("minPricePercentage must not exceed maxPricePercentage")}
	}

	if req.PriceSource == "" {
		req.PriceSource = "jita_sell"
	}
	if !allowedPriceSources[req.PriceSource] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Errorf("invalid priceSource: %s", req.PriceSource)}
	}

	config := &models.AutoBuyConfig{
		UserID:             *args.User,
		OwnerType:          req.OwnerType,
		OwnerID:            req.OwnerID,
		LocationID:         req.LocationID,
		ContainerID:        req.ContainerID,
		DivisionNumber:     req.DivisionNumber,
		MinPricePercentage: req.MinPricePercentage,
		MaxPricePercentage: req.MaxPricePercentage,
		PriceSource:        req.PriceSource,
	}

	if err := c.repository.Upsert(args.Request.Context(), config); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create auto-buy config")}
	}

	// Trigger immediate sync (auto-buy then auto-fulfill)
	go func() {
		ctx := context.Background()
		_ = c.syncer.SyncForUser(ctx, config.UserID)
		if c.autoFulfillSyncer != nil {
			_ = c.autoFulfillSyncer.SyncForUser(ctx, config.UserID)
		}
	}()

	return config, nil
}

// UpdateConfig updates an existing auto-buy configuration
func (c *AutoBuyConfigs) UpdateConfig(args *web.HandlerArgs) (any, *web.HttpError) {
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
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get auto-buy config")}
	}
	if existing == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("auto-buy config not found")}
	}
	if existing.UserID != *args.User {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("not authorized to update this config")}
	}

	var req struct {
		MinPricePercentage float64 `json:"minPricePercentage"`
		MaxPricePercentage float64 `json:"maxPricePercentage"`
		PriceSource        string  `json:"priceSource"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.MinPricePercentage < 0 || req.MinPricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("minPricePercentage must be between 0 and 200")}
	}
	if req.MaxPricePercentage <= 0 || req.MaxPricePercentage > 200 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("maxPricePercentage must be between 0 and 200")}
	}
	if req.MinPricePercentage > req.MaxPricePercentage {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("minPricePercentage must not exceed maxPricePercentage")}
	}

	if req.PriceSource != "" {
		if !allowedPriceSources[req.PriceSource] {
			return nil, &web.HttpError{StatusCode: 400, Error: errors.Errorf("invalid priceSource: %s", req.PriceSource)}
		}
		existing.PriceSource = req.PriceSource
	}

	existing.MinPricePercentage = req.MinPricePercentage
	existing.MaxPricePercentage = req.MaxPricePercentage
	if err := c.repository.Upsert(args.Request.Context(), existing); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update auto-buy config")}
	}

	// Trigger immediate sync (auto-buy then auto-fulfill)
	go func() {
		ctx := context.Background()
		_ = c.syncer.SyncForUser(ctx, existing.UserID)
		if c.autoFulfillSyncer != nil {
			_ = c.autoFulfillSyncer.SyncForUser(ctx, existing.UserID)
		}
	}()

	return existing, nil
}

// DeleteConfig soft-deletes an auto-buy config and deactivates all associated buy orders
func (c *AutoBuyConfigs) DeleteConfig(args *web.HandlerArgs) (any, *web.HttpError) {
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

	// Deactivate all associated buy orders first
	if err := c.buyOrderDeactor.DeactivateAutoBuyOrders(args.Request.Context(), id); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to deactivate auto-buy orders")}
	}

	if err := c.repository.Delete(args.Request.Context(), id, *args.User); err != nil {
		if err.Error() == "auto-buy config not found or user is not the owner" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete auto-buy config")}
	}

	return nil, nil
}
