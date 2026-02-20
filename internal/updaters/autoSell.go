package updaters

import (
	"context"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type AutoSellContainersRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.AutoSellContainer, error)
	GetAllActive(ctx context.Context) ([]*models.AutoSellContainer, error)
	GetItemsInContainer(ctx context.Context, ownerType string, ownerID int64, containerID int64) ([]*models.ContainerItem, error)
}

type AutoSellForSaleRepository interface {
	GetActiveAutoSellListings(ctx context.Context, autoSellContainerID int64) ([]*models.ForSaleItem, error)
	Upsert(ctx context.Context, item *models.ForSaleItem) error
	DeactivateAutoSellListings(ctx context.Context, autoSellContainerID int64) error
}

type AutoSellMarketPricesRepository interface {
	GetPricesForTypes(ctx context.Context, typeIDs []int64, regionID int64) (map[int64]*models.MarketPrice, error)
}

type AutoSell struct {
	autoSellRepo AutoSellContainersRepository
	forSaleRepo  AutoSellForSaleRepository
	marketRepo   AutoSellMarketPricesRepository
}

func NewAutoSell(
	autoSellRepo AutoSellContainersRepository,
	forSaleRepo AutoSellForSaleRepository,
	marketRepo AutoSellMarketPricesRepository,
) *AutoSell {
	return &AutoSell{
		autoSellRepo: autoSellRepo,
		forSaleRepo:  forSaleRepo,
		marketRepo:   marketRepo,
	}
}

// SyncForUser syncs auto-sell listings for a specific user after asset refresh
func (u *AutoSell) SyncForUser(ctx context.Context, userID int64) error {
	containers, err := u.autoSellRepo.GetByUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get auto-sell containers for user")
	}

	if len(containers) == 0 {
		return nil
	}

	for _, container := range containers {
		if err := u.syncContainer(ctx, container); err != nil {
			log.Error("failed to sync auto-sell container",
				"containerID", container.ID,
				"userID", container.UserID,
				"error", err)
		}
	}

	return nil
}

// SyncForAllUsers syncs auto-sell listings for all users after market price update
func (u *AutoSell) SyncForAllUsers(ctx context.Context) error {
	containers, err := u.autoSellRepo.GetAllActive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all active auto-sell containers")
	}

	if len(containers) == 0 {
		return nil
	}

	// Group by user for logging
	userContainers := make(map[int64][]*models.AutoSellContainer)
	for _, c := range containers {
		userContainers[c.UserID] = append(userContainers[c.UserID], c)
	}

	for userID, uContainers := range userContainers {
		for _, container := range uContainers {
			if err := u.syncContainer(ctx, container); err != nil {
				log.Error("failed to sync auto-sell container",
					"containerID", container.ID,
					"userID", userID,
					"error", err)
			}
		}
	}

	return nil
}

// resolveBasePrice returns the appropriate base price for the given price source.
func resolveBasePrice(price *models.MarketPrice, priceSource string) *float64 {
	switch priceSource {
	case "jita_sell":
		return price.SellPrice
	case "jita_split":
		if price.BuyPrice != nil && price.SellPrice != nil {
			split := (*price.BuyPrice + *price.SellPrice) / 2.0
			return &split
		}
		return nil
	default: // "jita_buy"
		return price.BuyPrice
	}
}

func (u *AutoSell) syncContainer(ctx context.Context, container *models.AutoSellContainer) error {
	// Get items currently in the container from asset tables
	items, err := u.autoSellRepo.GetItemsInContainer(ctx, container.OwnerType, container.OwnerID, container.ContainerID)
	if err != nil {
		return errors.Wrap(err, "failed to get items in container")
	}

	// Collect type IDs for market price lookup
	typeIDs := make([]int64, 0, len(items))
	for _, item := range items {
		typeIDs = append(typeIDs, item.TypeID)
	}

	// Get Jita buy prices
	prices := map[int64]*models.MarketPrice{}
	if len(typeIDs) > 0 {
		prices, err = u.marketRepo.GetPricesForTypes(ctx, typeIDs, JitaRegionID)
		if err != nil {
			return errors.Wrap(err, "failed to get market prices")
		}
	}

	// Get existing auto-sell listings for this container
	existingListings, err := u.forSaleRepo.GetActiveAutoSellListings(ctx, container.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get existing auto-sell listings")
	}

	// Build map of existing listings by type_id for quick lookup
	existingByType := make(map[int64]*models.ForSaleItem)
	for _, listing := range existingListings {
		existingByType[listing.TypeID] = listing
	}

	// Track which types are still in the container
	activeTypes := make(map[int64]bool)

	// For each item in the container, upsert a for-sale listing
	for _, item := range items {
		price, hasPrice := prices[item.TypeID]
		basePrice := (*float64)(nil)
		if hasPrice {
			basePrice = resolveBasePrice(price, container.PriceSource)
		}
		if basePrice == nil || *basePrice <= 0 {
			// No usable price — deactivate existing listing if any
			if existing, ok := existingByType[item.TypeID]; ok {
				existing.IsActive = false
				if err := u.forSaleRepo.Upsert(ctx, existing); err != nil {
					log.Error("failed to deactivate auto-sell listing with no price",
						"typeID", item.TypeID, "error", err)
				}
			}
			continue
		}

		activeTypes[item.TypeID] = true
		computedPrice := *basePrice * container.PricePercentage / 100.0

		forSaleItem := &models.ForSaleItem{
			UserID:              container.UserID,
			TypeID:              item.TypeID,
			OwnerType:           container.OwnerType,
			OwnerID:             container.OwnerID,
			LocationID:          container.LocationID,
			ContainerID:         &container.ContainerID,
			DivisionNumber:      container.DivisionNumber,
			QuantityAvailable:   item.Quantity,
			PricePerUnit:        computedPrice,
			AutoSellContainerID: &container.ID,
			IsActive:            true,
		}

		if err := u.forSaleRepo.Upsert(ctx, forSaleItem); err != nil {
			log.Error("failed to upsert auto-sell listing",
				"typeID", item.TypeID, "containerID", container.ID, "error", err)
		}
	}

	// Deactivate listings for items no longer in the container
	for _, listing := range existingListings {
		if !activeTypes[listing.TypeID] {
			listing.IsActive = false
			if err := u.forSaleRepo.Upsert(ctx, listing); err != nil {
				log.Error("failed to deactivate removed auto-sell listing",
					"typeID", listing.TypeID, "containerID", container.ID, "error", err)
			}
		}
	}

	return nil
}
