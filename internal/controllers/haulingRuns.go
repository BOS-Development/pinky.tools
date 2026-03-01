package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type HaulingRunsRepository interface {
	CreateRun(ctx context.Context, run *models.HaulingRun) (*models.HaulingRun, error)
	GetRunByID(ctx context.Context, id int64, userID int64) (*models.HaulingRun, error)
	ListRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error)
	UpdateRun(ctx context.Context, run *models.HaulingRun) error
	UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error
	DeleteRun(ctx context.Context, id int64, userID int64) error
}

type HaulingRunItemsRepository interface {
	AddItem(ctx context.Context, item *models.HaulingRunItem) (*models.HaulingRunItem, error)
	GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error)
	UpdateItemAcquired(ctx context.Context, itemID int64, runID int64, quantityAcquired int64) error
	RemoveItem(ctx context.Context, itemID int64, runID int64) error
}

type HaulingMarketUpdater interface {
	ScanRegion(ctx context.Context, regionID int64, systemID int64) error
}

type HaulingMarketRepository interface {
	GetScannerResults(ctx context.Context, sourceRegionID int64, sourceSystemID int64, destRegionID int64) ([]*models.HaulingArbitrageRow, error)
}

type HaulingRunsController struct {
	runs    HaulingRunsRepository
	items   HaulingRunItemsRepository
	market  HaulingMarketRepository
	scanner HaulingMarketUpdater
}

func NewHaulingRuns(
	router Routerer,
	runs HaulingRunsRepository,
	items HaulingRunItemsRepository,
	market HaulingMarketRepository,
	scanner HaulingMarketUpdater,
) *HaulingRunsController {
	c := &HaulingRunsController{runs: runs, items: items, market: market, scanner: scanner}
	router.RegisterRestAPIRoute("/v1/hauling/runs", web.AuthAccessUser, c.ListRuns, "GET")
	router.RegisterRestAPIRoute("/v1/hauling/runs", web.AuthAccessUser, c.CreateRun, "POST")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}", web.AuthAccessUser, c.GetRun, "GET")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}", web.AuthAccessUser, c.UpdateRun, "PUT")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}", web.AuthAccessUser, c.DeleteRun, "DELETE")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}/status", web.AuthAccessUser, c.UpdateStatus, "PUT")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}/items", web.AuthAccessUser, c.AddItem, "POST")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}/items/{itemId}", web.AuthAccessUser, c.UpdateItemAcquired, "PUT")
	router.RegisterRestAPIRoute("/v1/hauling/runs/{id}/items/{itemId}", web.AuthAccessUser, c.RemoveItem, "DELETE")
	router.RegisterRestAPIRoute("/v1/hauling/scanner", web.AuthAccessUser, c.GetScannerResults, "GET")
	router.RegisterRestAPIRoute("/v1/hauling/scanner/scan", web.AuthAccessUser, c.TriggerScan, "POST")
	return c
}

func (c *HaulingRunsController) ListRuns(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	runs, err := c.runs.ListRunsByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to list runs")}
	}
	return runs, nil
}

func (c *HaulingRunsController) CreateRun(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	var run models.HaulingRun
	if err := json.NewDecoder(args.Request.Body).Decode(&run); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	run.UserID = *args.User
	run.Status = "PLANNING"
	created, err := c.runs.CreateRun(args.Request.Context(), &run)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to create run")}
	}
	created.Items = []*models.HaulingRunItem{}
	return created, nil
}

func (c *HaulingRunsController) GetRun(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	run, err := c.runs.GetRunByID(args.Request.Context(), id, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get run")}
	}
	if run == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("run not found")}
	}
	items, err := c.items.GetItemsByRunID(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get items")}
	}
	run.Items = items
	return run, nil
}

func (c *HaulingRunsController) UpdateRun(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	var run models.HaulingRun
	if err := json.NewDecoder(args.Request.Body).Decode(&run); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	run.ID = id
	run.UserID = *args.User
	if err := c.runs.UpdateRun(args.Request.Context(), &run); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to update run")}
	}
	return nil, nil
}

func (c *HaulingRunsController) UpdateStatus(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(args.Request.Body).Decode(&body); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	validStatuses := map[string]bool{
		"PLANNING": true, "ACCUMULATING": true, "READY": true,
		"IN_TRANSIT": true, "SELLING": true, "COMPLETE": true, "CANCELLED": true,
	}
	if !validStatuses[body.Status] {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid status")}
	}
	if err := c.runs.UpdateRunStatus(args.Request.Context(), id, *args.User, body.Status); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to update status")}
	}
	return nil, nil
}

func (c *HaulingRunsController) DeleteRun(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	if err := c.runs.DeleteRun(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to delete run")}
	}
	return nil, nil
}

func (c *HaulingRunsController) AddItem(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	// Verify user owns the run
	run, err := c.runs.GetRunByID(args.Request.Context(), id, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get run")}
	}
	if run == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("run not found")}
	}
	var item models.HaulingRunItem
	if err := json.NewDecoder(args.Request.Body).Decode(&item); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	item.RunID = id
	created, err := c.items.AddItem(args.Request.Context(), &item)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to add item")}
	}
	return created, nil
}

func (c *HaulingRunsController) UpdateItemAcquired(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	runID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	itemID, err := strconv.ParseInt(args.Params["itemId"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid item id")}
	}
	var body struct {
		QuantityAcquired int64 `json:"quantityAcquired"`
	}
	if err := json.NewDecoder(args.Request.Body).Decode(&body); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	if body.QuantityAcquired < 0 {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("quantity_acquired must be non-negative")}
	}
	if err := c.items.UpdateItemAcquired(args.Request.Context(), itemID, runID, body.QuantityAcquired); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to update item")}
	}
	return nil, nil
}

func (c *HaulingRunsController) RemoveItem(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	runID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid run id")}
	}
	itemID, err := strconv.ParseInt(args.Params["itemId"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid item id")}
	}
	// Verify user owns the run
	run, err := c.runs.GetRunByID(args.Request.Context(), runID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get run")}
	}
	if run == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("run not found")}
	}
	if err := c.items.RemoveItem(args.Request.Context(), itemID, runID); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to remove item")}
	}
	return nil, nil
}

func (c *HaulingRunsController) GetScannerResults(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	q := args.Request.URL.Query()
	sourceRegionStr := q.Get("source_region_id")
	destRegionStr := q.Get("dest_region_id")
	sourceSystemStr := q.Get("source_system_id")

	if sourceRegionStr == "" || destRegionStr == "" {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("source_region_id and dest_region_id are required")}
	}
	sourceRegionID, err := strconv.ParseInt(sourceRegionStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid source_region_id")}
	}
	destRegionID, err := strconv.ParseInt(destRegionStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid dest_region_id")}
	}
	var sourceSystemID int64
	if sourceSystemStr != "" {
		sourceSystemID, err = strconv.ParseInt(sourceSystemStr, 10, 64)
		if err != nil {
			return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid source_system_id")}
		}
	}
	results, err := c.market.GetScannerResults(args.Request.Context(), sourceRegionID, sourceSystemID, destRegionID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get scanner results")}
	}
	return results, nil
}

func (c *HaulingRunsController) TriggerScan(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	var body struct {
		RegionID int64 `json:"regionId"`
		SystemID int64 `json:"systemId"`
	}
	if err := json.NewDecoder(args.Request.Body).Decode(&body); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	if body.RegionID == 0 {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("regionId is required")}
	}
	// Run scan in background
	go func() {
		ctx := context.Background()
		if err := c.scanner.ScanRegion(ctx, body.RegionID, body.SystemID); err != nil {
			slog.Error("background hauling scan failed", "region_id", body.RegionID, "error", err)
		}
	}()
	return map[string]string{"status": "scanning"}, nil
}
