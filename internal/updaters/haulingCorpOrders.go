package updaters

import (
	"context"
	"strings"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/pkg/errors"
)

// HaulingCorpUserRepository provides user IDs.
type HaulingCorpUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

// HaulingCorpCorporationRepository provides player corporations.
type HaulingCorpCorporationRepository interface {
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

// HaulingCorpRunsRepository provides hauling run access for the corp orders updater.
type HaulingCorpRunsRepository interface {
	ListAccumulatingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error)
}

// HaulingCorpRunItemsRepository provides hauling run item access for the corp orders updater.
type HaulingCorpRunItemsRepository interface {
	GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error)
}

// HaulingCorpItemsRepository provides item update access.
type HaulingCorpItemsRepository interface {
	UpdateItemAcquired(ctx context.Context, itemID int64, runID int64, quantityAcquired int64) error
}

// HaulingCorpEsiClient provides ESI corporation order access.
type HaulingCorpEsiClient interface {
	GetCorporationOrders(ctx context.Context, corporationID int64, token string) ([]*client.CorpOrder, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

// HaulingCorpOrdersUpdater matches corp buy orders to hauling run items for all users.
type HaulingCorpOrdersUpdater struct {
	userRepo     HaulingCorpUserRepository
	corpRepo     HaulingCorpCorporationRepository
	runsRepo     HaulingCorpRunsRepository
	runItemsRepo HaulingCorpRunItemsRepository
	itemsRepo    HaulingCorpItemsRepository
	esiClient    HaulingCorpEsiClient
}

// NewHaulingCorpOrders creates a new HaulingCorpOrdersUpdater.
func NewHaulingCorpOrders(
	userRepo HaulingCorpUserRepository,
	corpRepo HaulingCorpCorporationRepository,
	runsRepo HaulingCorpRunsRepository,
	runItemsRepo HaulingCorpRunItemsRepository,
	itemsRepo HaulingCorpItemsRepository,
	esiClient HaulingCorpEsiClient,
) *HaulingCorpOrdersUpdater {
	return &HaulingCorpOrdersUpdater{
		userRepo:     userRepo,
		corpRepo:     corpRepo,
		runsRepo:     runsRepo,
		runItemsRepo: runItemsRepo,
		itemsRepo:    itemsRepo,
		esiClient:    esiClient,
	}
}

// UpdateAllUsers processes corp orders for all users.
func (u *HaulingCorpOrdersUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for hauling corp orders update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserOrders(ctx, userID); err != nil {
			log.Error("failed to update hauling corp orders for user", "userID", userID, "error", err)
		}
	}

	return nil
}

// UpdateUserOrders processes corp orders for a single user and updates hauling run items.
func (u *HaulingCorpOrdersUpdater) UpdateUserOrders(ctx context.Context, userID int64) error {
	// Get ACCUMULATING runs for this user
	runs, err := u.runsRepo.ListAccumulatingByUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to list accumulating runs")
	}
	if len(runs) == 0 {
		return nil
	}

	// Get all player corporations for this user
	corporations, err := u.corpRepo.Get(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get corporations for user")
	}
	if len(corporations) == 0 {
		return nil
	}

	// Collect all buy orders from all corporations
	// Map: typeID -> list of (filled qty, characterID from issuedBy)
	type orderInfo struct {
		filledQty int64
		issuedBy  int64
	}
	// typeID -> slice of orders (multiple corps may have orders for same type)
	corpOrdersByType := map[int64][]orderInfo{}

	for _, corp := range corporations {
		if !strings.Contains(corp.EsiScopes, "esi-corporations.read_orders.v1") {
			continue
		}

		token := corp.EsiToken
		if time.Now().After(corp.EsiExpiresOn) {
			refreshed, err := u.esiClient.RefreshAccessToken(ctx, corp.EsiRefreshToken)
			if err != nil {
				log.Error("failed to refresh token for corporation (hauling corp orders)", "corporationID", corp.ID, "error", err)
				continue
			}
			token = refreshed.AccessToken
			if err := u.corpRepo.UpdateTokens(ctx, corp.ID, corp.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry); err != nil {
				log.Error("failed to persist refreshed token for corporation (hauling corp orders)", "corporationID", corp.ID, "error", err)
			}
		}

		orders, err := u.esiClient.GetCorporationOrders(ctx, corp.ID, token)
		if err != nil {
			log.Error("failed to get corporation orders for hauling", "corporationID", corp.ID, "error", err)
			continue
		}

		for _, order := range orders {
			if !order.IsBuyOrder {
				continue
			}
			filled := order.VolumeTotal - order.VolumeRemain
			corpOrdersByType[order.TypeID] = append(corpOrdersByType[order.TypeID], orderInfo{
				filledQty: filled,
				issuedBy:  order.IssuedBy,
			})
		}
	}

	if len(corpOrdersByType) == 0 {
		return nil
	}

	// For each run, match corp orders to run items and update quantity_acquired
	for _, run := range runs {
		items, err := u.runItemsRepo.GetItemsByRunID(ctx, run.ID)
		if err != nil {
			log.Error("failed to get items for hauling run", "runID", run.ID, "error", err)
			continue
		}

		for _, item := range items {
			orders, ok := corpOrdersByType[item.TypeID]
			if !ok {
				continue
			}

			// Sum filled quantity across all matching orders
			// If the item has a character_id, prefer orders from that character
			var totalFilled int64
			for _, o := range orders {
				if item.CharacterID != nil && *item.CharacterID == o.issuedBy {
					// Exact match by character
					totalFilled += o.filledQty
				} else if item.CharacterID == nil {
					// No character filter — accept any order for this type
					totalFilled += o.filledQty
				}
			}

			// If character-specific match returned 0, fall back to any order
			if totalFilled == 0 && item.CharacterID != nil {
				for _, o := range orders {
					totalFilled += o.filledQty
				}
			}

			if totalFilled != item.QuantityAcquired {
				if err := u.itemsRepo.UpdateItemAcquired(ctx, item.ID, run.ID, totalFilled); err != nil {
					log.Error("failed to update item acquired for hauling run",
						"itemID", item.ID,
						"runID", run.ID,
						"typeID", item.TypeID,
						"error", err,
					)
				} else {
					log.Info("updated hauling run item acquired from corp orders",
						"runID", run.ID,
						"typeID", item.TypeID,
						"quantityAcquired", totalFilled,
					)
				}
			}
		}
	}

	return nil
}
