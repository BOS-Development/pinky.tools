package controllers

import (
	"context"

	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type MarketPricesUpdater interface {
	UpdateJitaMarket(ctx context.Context) error
}

type MarketPrices struct {
	updater MarketPricesUpdater
}

func NewMarketPrices(router Routerer, updater MarketPricesUpdater) *MarketPrices {
	controller := &MarketPrices{
		updater: updater,
	}

	router.RegisterRestAPIRoute("/v1/market-prices/update", web.AuthAccessUser, controller.UpdateJitaMarket, "POST")

	return controller
}

func (c *MarketPrices) UpdateJitaMarket(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	err := c.updater.UpdateJitaMarket(args.Request.Context())
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to update market prices"),
		}
	}
	return nil, nil
}
