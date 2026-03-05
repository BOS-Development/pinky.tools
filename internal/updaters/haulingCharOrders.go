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

// HaulingCharOrdersUserRepository provides user IDs.
type HaulingCharOrdersUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

// HaulingCharOrdersCharacterRepository provides character access.
type HaulingCharOrdersCharacterRepository interface {
	GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

// HaulingCharOrdersRunsRepository provides hauling run access for the char orders updater.
type HaulingCharOrdersRunsRepository interface {
	ListSellingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error)
	UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error
}

// HaulingCharOrdersRunItemsRepository provides hauling run item read access.
type HaulingCharOrdersRunItemsRepository interface {
	GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error)
	UpdateItemSold(ctx context.Context, itemID int64, runID int64, qtySold int64, sellOrderID *int64) error
}

// HaulingCharOrdersNotifier sends notifications for item sold and run complete events.
type HaulingCharOrdersNotifier interface {
	NotifyHaulingItemSold(ctx context.Context, userID int64, run *models.HaulingRun, item *models.HaulingRunItem)
	NotifyHaulingComplete(ctx context.Context, userID int64, run *models.HaulingRun, summary *models.HaulingRunPnlSummary)
}

// HaulingCharOrdersEsiClient provides ESI character order access.
type HaulingCharOrdersEsiClient interface {
	GetCharacterOrders(ctx context.Context, characterID int64, token string) ([]*client.CharacterOrder, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

// HaulingCharOrdersUpdater matches character sell orders to hauling run items for all users.
type HaulingCharOrdersUpdater struct {
	userRepo     HaulingCharOrdersUserRepository
	charRepo     HaulingCharOrdersCharacterRepository
	runsRepo     HaulingCharOrdersRunsRepository
	runItemsRepo HaulingCharOrdersRunItemsRepository
	notifier     HaulingCharOrdersNotifier
	esiClient    HaulingCharOrdersEsiClient
}

// NewHaulingCharOrders creates a new HaulingCharOrdersUpdater.
func NewHaulingCharOrders(
	userRepo HaulingCharOrdersUserRepository,
	charRepo HaulingCharOrdersCharacterRepository,
	runsRepo HaulingCharOrdersRunsRepository,
	runItemsRepo HaulingCharOrdersRunItemsRepository,
	notifier HaulingCharOrdersNotifier,
	esiClient HaulingCharOrdersEsiClient,
) *HaulingCharOrdersUpdater {
	return &HaulingCharOrdersUpdater{
		userRepo:     userRepo,
		charRepo:     charRepo,
		runsRepo:     runsRepo,
		runItemsRepo: runItemsRepo,
		notifier:     notifier,
		esiClient:    esiClient,
	}
}

// UpdateAllUsers processes character sell orders for all users.
func (u *HaulingCharOrdersUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for hauling char orders update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserOrders(ctx, userID); err != nil {
			log.Error("failed to update hauling char orders for user", "userID", userID, "error", err)
		}
	}

	return nil
}

// UpdateUserOrders processes character sell orders for a single user and updates hauling run items.
func (u *HaulingCharOrdersUpdater) UpdateUserOrders(ctx context.Context, userID int64) error {
	// Get SELLING runs for this user
	runs, err := u.runsRepo.ListSellingByUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to list selling runs")
	}
	if len(runs) == 0 {
		return nil
	}

	// Get all characters for this user
	characters, err := u.charRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get characters for user")
	}
	if len(characters) == 0 {
		return nil
	}

	// Collect all sell orders from all characters
	// Map: typeID -> list of sell order info
	type sellOrderInfo struct {
		orderID      int64
		regionID     int64
		qtySold      int64
		issued       time.Time
	}
	charSellOrdersByType := map[int64][]sellOrderInfo{}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-markets.read_character_orders.v1") {
			continue
		}

		token := char.EsiToken
		if time.Now().After(char.EsiTokenExpiresOn) {
			refreshed, err := u.esiClient.RefreshAccessToken(ctx, char.EsiRefreshToken)
			if err != nil {
				log.Error("failed to refresh token for character (hauling char orders)", "characterID", char.ID, "error", err)
				continue
			}
			token = refreshed.AccessToken
			if err := u.charRepo.UpdateTokens(ctx, char.ID, char.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry); err != nil {
				log.Error("failed to persist refreshed token for character (hauling char orders)", "characterID", char.ID, "error", err)
			}
		}

		orders, err := u.esiClient.GetCharacterOrders(ctx, char.ID, token)
		if err != nil {
			log.Error("failed to get character orders for hauling", "characterID", char.ID, "error", err)
			continue
		}

		for _, order := range orders {
			if order.IsBuyOrder {
				continue
			}
			issued, err := time.Parse(time.RFC3339, order.Issued)
			if err != nil {
				log.Error("failed to parse order issued time", "orderID", order.OrderID, "issued", order.Issued, "error", err)
				continue
			}
			qtySold := order.VolumeTotal - order.VolumeRemain
			charSellOrdersByType[order.TypeID] = append(charSellOrdersByType[order.TypeID], sellOrderInfo{
				orderID:  order.OrderID,
				regionID: order.RegionID,
				qtySold:  qtySold,
				issued:   issued,
			})
		}
	}

	if len(charSellOrdersByType) == 0 {
		return nil
	}

	// For each run, match sell orders to run items and update qty_sold
	for _, run := range runs {
		runCreatedAt, err := time.Parse(time.RFC3339, run.CreatedAt)
		if err != nil {
			log.Error("failed to parse run created_at", "runID", run.ID, "error", err)
			continue
		}

		items, err := u.runItemsRepo.GetItemsByRunID(ctx, run.ID)
		if err != nil {
			log.Error("failed to get items for hauling run (char orders)", "runID", run.ID, "error", err)
			continue
		}

		allSold := true
		for _, item := range items {
			orders, ok := charSellOrdersByType[item.TypeID]
			if !ok {
				allSold = false
				continue
			}

			// Find matching orders: must be sell, issued after run.CreatedAt, in run's destination region
			var totalSold int64
			var bestOrderID *int64
			for i := range orders {
				o := &orders[i]
				if o.regionID != run.ToRegionID {
					continue
				}
				if !o.issued.After(runCreatedAt) {
					continue
				}
				totalSold += o.qtySold
				if bestOrderID == nil {
					bestOrderID = &o.orderID
				}
			}

			if totalSold != item.QtySold {
				prevSold := item.QtySold
				if err := u.runItemsRepo.UpdateItemSold(ctx, item.ID, run.ID, totalSold, bestOrderID); err != nil {
					log.Error("failed to update item sold for hauling run",
						"itemID", item.ID,
						"runID", run.ID,
						"typeID", item.TypeID,
						"error", err,
					)
					allSold = false
					continue
				}
				log.Info("updated hauling run item sold from char orders",
					"runID", run.ID,
					"typeID", item.TypeID,
					"qtySold", totalSold,
				)

				// Notify if newly fully sold
				if prevSold < item.QuantityPlanned && totalSold >= item.QuantityPlanned && run.NotifyTier3 && u.notifier != nil {
					item.QtySold = totalSold
					go u.notifier.NotifyHaulingItemSold(ctx, run.UserID, run, item)
				}
			}

			if totalSold < item.QuantityPlanned {
				allSold = false
			}
		}

		// Auto-transition to COMPLETE if all items fully sold
		if allSold && len(items) > 0 {
			if err := u.runsRepo.UpdateRunStatus(ctx, run.ID, run.UserID, "COMPLETE"); err != nil {
				log.Error("failed to auto-complete hauling run", "runID", run.ID, "error", err)
			} else {
				log.Info("auto-completed hauling run (all items sold)", "runID", run.ID)
				if u.notifier != nil {
					go u.notifier.NotifyHaulingComplete(ctx, run.UserID, run, nil)
				}
			}
		}
	}

	return nil
}
