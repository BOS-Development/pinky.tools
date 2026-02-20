package updaters

import (
	"context"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type AutoBuyConfigsRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.AutoBuyConfig, error)
	GetAllActive(ctx context.Context) ([]*models.AutoBuyConfig, error)
	GetStockpileDeficitsForConfig(ctx context.Context, config *models.AutoBuyConfig) ([]*models.StockpileDeficitItem, error)
}

type AutoBuyOrdersRepository interface {
	GetActiveAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) ([]*models.BuyOrder, error)
	UpsertAutoBuy(ctx context.Context, order *models.BuyOrder) error
	DeactivateAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) error
	DeactivateAutoBuyOrder(ctx context.Context, orderID int64) error
}

type AutoBuyMarketPricesRepository interface {
	GetPricesForTypes(ctx context.Context, typeIDs []int64, regionID int64) (map[int64]*models.MarketPrice, error)
}

type AutoBuy struct {
	configRepo   AutoBuyConfigsRepository
	buyOrderRepo AutoBuyOrdersRepository
	marketRepo   AutoBuyMarketPricesRepository
}

func NewAutoBuy(
	configRepo AutoBuyConfigsRepository,
	buyOrderRepo AutoBuyOrdersRepository,
	marketRepo AutoBuyMarketPricesRepository,
) *AutoBuy {
	return &AutoBuy{
		configRepo:   configRepo,
		buyOrderRepo: buyOrderRepo,
		marketRepo:   marketRepo,
	}
}

// SyncForUser syncs auto-buy orders for a specific user after asset refresh
func (u *AutoBuy) SyncForUser(ctx context.Context, userID int64) error {
	configs, err := u.configRepo.GetByUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get auto-buy configs for user")
	}

	if len(configs) == 0 {
		return nil
	}

	for _, config := range configs {
		if err := u.syncConfig(ctx, config); err != nil {
			log.Error("failed to sync auto-buy config",
				"configID", config.ID,
				"userID", config.UserID,
				"error", err)
		}
	}

	return nil
}

// SyncForAllUsers syncs auto-buy orders for all users after market price update
func (u *AutoBuy) SyncForAllUsers(ctx context.Context) error {
	configs, err := u.configRepo.GetAllActive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all active auto-buy configs")
	}

	if len(configs) == 0 {
		return nil
	}

	userConfigs := make(map[int64][]*models.AutoBuyConfig)
	for _, c := range configs {
		userConfigs[c.UserID] = append(userConfigs[c.UserID], c)
	}

	for userID, uConfigs := range userConfigs {
		for _, config := range uConfigs {
			if err := u.syncConfig(ctx, config); err != nil {
				log.Error("failed to sync auto-buy config",
					"configID", config.ID,
					"userID", userID,
					"error", err)
			}
		}
	}

	return nil
}

func (u *AutoBuy) syncConfig(ctx context.Context, config *models.AutoBuyConfig) error {
	// Get stockpile deficits for this config's container context
	deficits, err := u.configRepo.GetStockpileDeficitsForConfig(ctx, config)
	if err != nil {
		return errors.Wrap(err, "failed to get stockpile deficits")
	}

	// Collect type IDs for market price lookup
	typeIDs := make([]int64, 0, len(deficits))
	for _, d := range deficits {
		if d.Deficit > 0 {
			typeIDs = append(typeIDs, d.TypeID)
		}
	}

	// Get Jita prices
	prices := map[int64]*models.MarketPrice{}
	if len(typeIDs) > 0 {
		prices, err = u.marketRepo.GetPricesForTypes(ctx, typeIDs, JitaRegionID)
		if err != nil {
			return errors.Wrap(err, "failed to get market prices")
		}
	}

	// Get existing auto-buy orders for this config
	existingOrders, err := u.buyOrderRepo.GetActiveAutoBuyOrders(ctx, config.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get existing auto-buy orders")
	}

	existingByType := make(map[int64]*models.BuyOrder)
	for _, order := range existingOrders {
		existingByType[order.TypeID] = order
	}

	// Track which types still have a deficit
	activeTypes := make(map[int64]bool)

	for _, deficit := range deficits {
		if deficit.Deficit <= 0 {
			continue
		}

		// Resolve pricing: per-item override takes priority over config default
		priceSource := config.PriceSource
		maxPricePercentage := config.MaxPricePercentage
		minPricePercentage := config.MinPricePercentage
		if deficit.PriceSource != nil {
			priceSource = *deficit.PriceSource
		}
		if deficit.PricePercentage != nil {
			maxPricePercentage = *deficit.PricePercentage
		}

		price, hasPrice := prices[deficit.TypeID]
		basePrice := (*float64)(nil)
		if hasPrice {
			basePrice = resolveBasePrice(price, priceSource)
		}
		if basePrice == nil || *basePrice <= 0 {
			// No usable price — deactivate existing order if any
			if existing, ok := existingByType[deficit.TypeID]; ok {
				if err := u.buyOrderRepo.DeactivateAutoBuyOrder(ctx, existing.ID); err != nil {
					log.Error("failed to deactivate auto-buy order with no price",
						"typeID", deficit.TypeID, "error", err)
				}
			}
			continue
		}

		activeTypes[deficit.TypeID] = true
		computedMaxPrice := *basePrice * maxPricePercentage / 100.0
		computedMinPrice := *basePrice * minPricePercentage / 100.0

		order := &models.BuyOrder{
			BuyerUserID:     config.UserID,
			TypeID:          deficit.TypeID,
			LocationID:      config.LocationID,
			QuantityDesired: deficit.Deficit,
			MinPricePerUnit: computedMinPrice,
			MaxPricePerUnit: computedMaxPrice,
			AutoBuyConfigID: &config.ID,
			IsActive:        true,
		}

		if err := u.buyOrderRepo.UpsertAutoBuy(ctx, order); err != nil {
			log.Error("failed to upsert auto-buy order",
				"typeID", deficit.TypeID, "configID", config.ID, "error", err)
		}
	}

	// Deactivate orders for items no longer in deficit
	for _, order := range existingOrders {
		if !activeTypes[order.TypeID] {
			if err := u.buyOrderRepo.DeactivateAutoBuyOrder(ctx, order.ID); err != nil {
				log.Error("failed to deactivate removed auto-buy order",
					"typeID", order.TypeID, "configID", config.ID, "error", err)
			}
		}
	}

	return nil
}
