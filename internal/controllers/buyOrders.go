package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type BuyOrdersRepository interface {
	Create(ctx context.Context, order *models.BuyOrder) error
	GetByID(ctx context.Context, id int64) (*models.BuyOrder, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.BuyOrder, error)
	GetDemandForSeller(ctx context.Context, sellerUserID int64) ([]*models.BuyOrder, error)
	Update(ctx context.Context, order *models.BuyOrder) error
	Delete(ctx context.Context, id int64, userID int64) error
}

type BuyOrdersController struct {
	repository BuyOrdersRepository
	permRepo   ContactPermissionsRepository
}

func NewBuyOrders(router Routerer, repository BuyOrdersRepository, permRepo ContactPermissionsRepository) *BuyOrdersController {
	controller := &BuyOrdersController{
		repository: repository,
		permRepo:   permRepo,
	}

	router.RegisterRestAPIRoute("/v1/buy-orders", web.AuthAccessUser, controller.GetMyOrders, "GET")
	router.RegisterRestAPIRoute("/v1/buy-orders", web.AuthAccessUser, controller.CreateOrder, "POST")
	router.RegisterRestAPIRoute("/v1/buy-orders/{id}", web.AuthAccessUser, controller.UpdateOrder, "PUT")
	router.RegisterRestAPIRoute("/v1/buy-orders/{id}", web.AuthAccessUser, controller.DeleteOrder, "DELETE")
	router.RegisterRestAPIRoute("/v1/buy-orders/demand", web.AuthAccessUser, controller.GetDemand, "GET")

	log.Info("registering buy orders controller", "endpoints", []string{
		"/v1/buy-orders GET",
		"/v1/buy-orders POST",
		"/v1/buy-orders/{id} PUT",
		"/v1/buy-orders/{id} DELETE",
		"/v1/buy-orders/demand GET",
	})

	return controller
}

// GetMyOrders returns all buy orders for the authenticated user
func (c *BuyOrdersController) GetMyOrders(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	orders, err := c.repository.GetByUser(ctx, *args.User)
	if err != nil {
		log.Error("failed to get buy orders", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
		}
	}

	return orders, nil
}

// CreateOrder creates a new buy order
func (c *BuyOrdersController) CreateOrder(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	var req struct {
		TypeID          int64   `json:"typeId"`
		LocationID      int64   `json:"locationId"`
		QuantityDesired int64   `json:"quantityDesired"`
		MaxPricePerUnit float64 `json:"maxPricePerUnit"`
		Notes           *string `json:"notes"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		log.Error("failed to decode buy order request", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      err,
		}
	}

	// Validate required fields
	if req.TypeID == 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("typeId is required"),
		}
	}

	if req.LocationID == 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("locationId is required"),
		}
	}

	if req.QuantityDesired <= 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("quantityDesired must be positive"),
		}
	}

	if req.MaxPricePerUnit < 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("maxPricePerUnit must be non-negative"),
		}
	}

	order := &models.BuyOrder{
		BuyerUserID:     *args.User,
		TypeID:          req.TypeID,
		LocationID:      req.LocationID,
		QuantityDesired: req.QuantityDesired,
		MaxPricePerUnit: req.MaxPricePerUnit,
		Notes:           req.Notes,
		IsActive:        true,
	}

	if err := c.repository.Create(ctx, order); err != nil {
		log.Error("failed to create buy order", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
		}
	}

	log.Info("buy order created", "orderId", order.ID, "userId", *args.User, "typeId", req.TypeID)

	return order, nil
}

// UpdateOrder updates an existing buy order
func (c *BuyOrdersController) UpdateOrder(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	idStr := args.Params["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("invalid order ID"),
		}
	}

	// Get existing order
	order, err := c.repository.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get buy order", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusNotFound,
			Error:      err,
		}
	}

	// Verify ownership
	if order.BuyerUserID != *args.User {
		return nil, &web.HttpError{
			StatusCode: http.StatusForbidden,
			Error:      errors.New("you do not own this buy order"),
		}
	}

	var req struct {
		LocationID      int64   `json:"locationId"`
		QuantityDesired int64   `json:"quantityDesired"`
		MaxPricePerUnit float64 `json:"maxPricePerUnit"`
		Notes           *string `json:"notes"`
		IsActive        *bool   `json:"isActive"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		log.Error("failed to decode update request", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      err,
		}
	}

	// Validate
	if req.LocationID == 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("locationId is required"),
		}
	}

	if req.QuantityDesired <= 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("quantityDesired must be positive"),
		}
	}

	if req.MaxPricePerUnit < 0 {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("maxPricePerUnit must be non-negative"),
		}
	}

	// Update fields
	order.LocationID = req.LocationID
	order.QuantityDesired = req.QuantityDesired
	order.MaxPricePerUnit = req.MaxPricePerUnit
	order.Notes = req.Notes
	if req.IsActive != nil {
		order.IsActive = *req.IsActive
	}

	if err := c.repository.Update(ctx, order); err != nil {
		log.Error("failed to update buy order", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
		}
	}

	log.Info("buy order updated", "orderId", id, "userId", *args.User)

	return order, nil
}

// DeleteOrder soft-deletes a buy order
func (c *BuyOrdersController) DeleteOrder(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	idStr := args.Params["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: http.StatusBadRequest,
			Error:      errors.New("invalid order ID"),
		}
	}

	if err := c.repository.Delete(ctx, id, *args.User); err != nil {
		log.Error("failed to delete buy order", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusNotFound,
			Error:      err,
		}
	}

	log.Info("buy order deleted", "orderId", id, "userId", *args.User)

	return map[string]string{"status": "deleted"}, nil
}

// GetDemand returns active buy orders from contacts who have granted permission
func (c *BuyOrdersController) GetDemand(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	orders, err := c.repository.GetDemandForSeller(ctx, *args.User)
	if err != nil {
		log.Error("failed to get demand", "error", err.Error())
		return nil, &web.HttpError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
		}
	}

	log.Info("demand retrieved", "userId", *args.User, "count", len(orders))

	return orders, nil
}
