package updaters

import (
	"context"
	"math"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

const JitaRegionID = 10000002
const UpdateInterval = 6 * time.Hour

type MarketPricesRepository interface {
	UpsertPrices(ctx context.Context, prices []models.MarketPrice) error
	DeleteAllForRegion(ctx context.Context, regionID int64) error
	GetLastUpdateTime(ctx context.Context, regionID int64) (*time.Time, error)
}

type MarketPricesEsiClient interface {
	GetMarketOrders(ctx context.Context, regionID int64) ([]*client.MarketOrder, error)
}

type AutoSellAllUsersSyncer interface {
	SyncForAllUsers(ctx context.Context) error
}

type AutoBuyAllUsersSyncer interface {
	SyncForAllUsers(ctx context.Context) error
}

type AutoFulfillAllUsersSyncer interface {
	SyncForAllUsers(ctx context.Context) error
}

type MarketPrices struct {
	marketPricesRepo    MarketPricesRepository
	esiClient           MarketPricesEsiClient
	autoSellSyncer      AutoSellAllUsersSyncer
	autoBuySyncer       AutoBuyAllUsersSyncer
	autoFulfillSyncer   AutoFulfillAllUsersSyncer
}

func NewMarketPrices(repo MarketPricesRepository, esiClient MarketPricesEsiClient) *MarketPrices {
	return &MarketPrices{
		marketPricesRepo: repo,
		esiClient:        esiClient,
	}
}

func (u *MarketPrices) UpdateJitaMarket(ctx context.Context) error {
	// Check when the last update was
	lastUpdate, err := u.marketPricesRepo.GetLastUpdateTime(ctx, JitaRegionID)
	if err != nil {
		return errors.Wrap(err, "failed to get last market price update time")
	}

	// If we have a recent update (within the last 6 hours), skip
	if lastUpdate != nil {
		timeSinceUpdate := time.Since(*lastUpdate)
		if timeSinceUpdate < UpdateInterval {
			log.Info("skipping market price update, last update was recent",
				"last_update", lastUpdate.Format(time.RFC3339),
				"time_since_update", timeSinceUpdate.String(),
				"next_update_in", (UpdateInterval - timeSinceUpdate).String())
			return nil
		}
	}

	log.Info("updating market prices", "region_id", JitaRegionID)

	// Fetch all market orders for Jita
	orders, err := u.esiClient.GetMarketOrders(ctx, JitaRegionID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch market orders from ESI")
	}

	// Group orders by type_id
	ordersByType := make(map[int64][]*client.MarketOrder)
	for _, order := range orders {
		ordersByType[order.TypeID] = append(ordersByType[order.TypeID], order)
	}

	// Calculate best prices for each type
	prices := make([]models.MarketPrice, 0, len(ordersByType))
	for typeID, typeOrders := range ordersByType {
		bestBuy := 0.0
		bestSell := math.MaxFloat64
		totalVolume := int64(0)

		for _, order := range typeOrders {
			totalVolume += order.VolumeRemain

			if order.IsBuyOrder {
				// For buy orders, we want the highest price (best bid)
				if order.Price > bestBuy {
					bestBuy = order.Price
				}
			} else {
				// For sell orders, we want the lowest price (best ask)
				if order.Price < bestSell {
					bestSell = order.Price
				}
			}
		}

		var buyPrice *float64
		var sellPrice *float64

		if bestBuy > 0 {
			buyPrice = &bestBuy
		}

		if bestSell < math.MaxFloat64 {
			sellPrice = &bestSell
		}

		prices = append(prices, models.MarketPrice{
			TypeID:      typeID,
			RegionID:    JitaRegionID,
			BuyPrice:    buyPrice,
			SellPrice:   sellPrice,
			DailyVolume: &totalVolume,
		})
	}

	// Delete old prices
	err = u.marketPricesRepo.DeleteAllForRegion(ctx, JitaRegionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete old market prices")
	}

	// Upsert new prices
	err = u.marketPricesRepo.UpsertPrices(ctx, prices)
	if err != nil {
		return errors.Wrap(err, "failed to upsert market prices")
	}

	if u.autoSellSyncer != nil {
		if err := u.autoSellSyncer.SyncForAllUsers(ctx); err != nil {
			log.Error("failed to sync auto-sell listings after market price update", "error", err)
		}
	}

	if u.autoBuySyncer != nil {
		if err := u.autoBuySyncer.SyncForAllUsers(ctx); err != nil {
			log.Error("failed to sync auto-buy orders after market price update", "error", err)
		}
	}

	if u.autoFulfillSyncer != nil {
		if err := u.autoFulfillSyncer.SyncForAllUsers(ctx); err != nil {
			log.Error("failed to sync auto-fulfill after market price update", "error", err)
		}
	}

	return nil
}

// WithAutoSellUpdater sets the optional auto-sell syncer
func (u *MarketPrices) WithAutoSellUpdater(syncer AutoSellAllUsersSyncer) {
	u.autoSellSyncer = syncer
}

// WithAutoBuyUpdater sets the optional auto-buy syncer
func (u *MarketPrices) WithAutoBuyUpdater(syncer AutoBuyAllUsersSyncer) {
	u.autoBuySyncer = syncer
}

// WithAutoFulfillUpdater sets the optional auto-fulfill syncer
func (u *MarketPrices) WithAutoFulfillUpdater(syncer AutoFulfillAllUsersSyncer) {
	u.autoFulfillSyncer = syncer
}
