package controllers

import (
	"context"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
)

type StockpilesRepository interface {
	GetStockpileDeficits(ctx context.Context, user int64) (*repositories.StockpilesResponse, error)
}

type Stockpiles struct {
	stockpilesRepo StockpilesRepository
}

func NewStockpiles(router Routerer, stockpilesRepo StockpilesRepository) *Stockpiles {
	controller := &Stockpiles{
		stockpilesRepo: stockpilesRepo,
	}

	router.RegisterRestAPIRoute("/v1/stockpiles/deficits", web.AuthAccessUser, controller.GetDeficits, "GET")

	return controller
}

func (c *Stockpiles) GetDeficits(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	stockpiles, err := c.stockpilesRepo.GetStockpileDeficits(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      err,
		}
	}

	return stockpiles, nil
}
