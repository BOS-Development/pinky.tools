package controllers

import (
	"context"
	"net/http"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ESIRefreshUpdater interface {
	ForceUpdateJitaMarket(ctx context.Context) error
}

type ESIRefreshCostIndicesUpdater interface {
	ForceUpdate(ctx context.Context) error
}

type ESIRefreshCcpPricesUpdater interface {
	ForceUpdate(ctx context.Context) error
}

type ESIRefreshResult struct {
	MarketPrices string `json:"market_prices"`
	CostIndices  string `json:"cost_indices"`
	CcpPrices    string `json:"ccp_prices"`
}

type ESIRefresh struct {
	marketPrices ESIRefreshUpdater
	costIndices  ESIRefreshCostIndicesUpdater
	ccpPrices    ESIRefreshCcpPricesUpdater
}

func NewESIRefresh(router Routerer, market ESIRefreshUpdater, costIndices ESIRefreshCostIndicesUpdater, ccp ESIRefreshCcpPricesUpdater) *ESIRefresh {
	c := &ESIRefresh{
		marketPrices: market,
		costIndices:  costIndices,
		ccpPrices:    ccp,
	}

	router.RegisterRestAPIRoute("/v1/esi/refresh", web.AuthAccessUser, c.RefreshAll, "POST")

	return c
}

func (c *ESIRefresh) RefreshAll(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	result := &ESIRefreshResult{
		MarketPrices: "ok",
		CostIndices:  "ok",
		CcpPrices:    "ok",
	}

	if err := c.marketPrices.ForceUpdateJitaMarket(ctx); err != nil {
		log.Error("esi refresh: market prices failed", "error", err)
		result.MarketPrices = "error: " + err.Error()
	}

	if err := c.costIndices.ForceUpdate(ctx); err != nil {
		log.Error("esi refresh: cost indices failed", "error", err)
		result.CostIndices = "error: " + err.Error()
	}

	if err := c.ccpPrices.ForceUpdate(ctx); err != nil {
		log.Error("esi refresh: ccp prices failed", "error", err)
		result.CcpPrices = "error: " + err.Error()
	}

	// Return 207 Multi-Status when at least one updater succeeded but others failed
	hasError := result.MarketPrices != "ok" || result.CostIndices != "ok" || result.CcpPrices != "ok"
	allFailed := result.MarketPrices != "ok" && result.CostIndices != "ok" && result.CcpPrices != "ok"

	if allFailed {
		return nil, &web.HttpError{
			StatusCode: http.StatusInternalServerError,
			Error:      errors.New("all ESI refresh operations failed"),
		}
	}

	if hasError {
		// Partial success — return 207 with the result body
		// We can't return a non-200 with a body via the standard path, so we
		// log the partial failure and return 200 with the result details.
		log.Warn("esi refresh: partial failure", "result", result)
	}

	return result, nil
}
