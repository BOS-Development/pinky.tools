package controllers

import (
	"context"
	"encoding/json"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type StockpileMarkersRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.StockpileMarker, error)
	Upsert(ctx context.Context, marker *models.StockpileMarker) error
	Delete(ctx context.Context, marker *models.StockpileMarker) error
}

type StockpileMarkers struct {
	repository StockpileMarkersRepository
}

func NewStockpileMarkers(router Routerer, repository StockpileMarkersRepository) *StockpileMarkers {
	controller := &StockpileMarkers{
		repository: repository,
	}

	router.RegisterRestAPIRoute("/v1/stockpiles", web.AuthAccessUser, controller.GetStockpiles, "GET")
	router.RegisterRestAPIRoute("/v1/stockpiles", web.AuthAccessUser, controller.UpsertStockpile, "POST")
	router.RegisterRestAPIRoute("/v1/stockpiles", web.AuthAccessUser, controller.DeleteStockpile, "DELETE")

	return controller
}

func (c *StockpileMarkers) GetStockpiles(args *web.HandlerArgs) (any, *web.HttpError) {
	markers, err := c.repository.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get stockpile markers"),
		}
	}

	return markers, nil
}

func (c *StockpileMarkers) UpsertStockpile(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var marker models.StockpileMarker
	err := d.Decode(&marker)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.Wrap(err, "failed to decode json"),
		}
	}

	// Set user ID from auth context
	marker.UserID = *args.User

	err = c.repository.Upsert(args.Request.Context(), &marker)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to upsert stockpile marker"),
		}
	}

	return nil, nil
}

func (c *StockpileMarkers) DeleteStockpile(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var marker models.StockpileMarker
	err := d.Decode(&marker)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.Wrap(err, "failed to decode json"),
		}
	}

	// Set user ID from auth context
	marker.UserID = *args.User

	err = c.repository.Delete(args.Request.Context(), &marker)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to delete stockpile marker"),
		}
	}

	return nil, nil
}
