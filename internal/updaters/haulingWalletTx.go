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

// HaulingWalletTxUserRepository provides user IDs.
type HaulingWalletTxUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

// HaulingWalletTxCharacterRepository provides character access.
type HaulingWalletTxCharacterRepository interface {
	GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

// HaulingWalletTxRunsRepository provides hauling run access for the wallet tx updater.
type HaulingWalletTxRunsRepository interface {
	ListSellingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error)
}

// HaulingWalletTxItemsRepository provides item update access for the wallet tx updater.
type HaulingWalletTxItemsRepository interface {
	GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error)
	UpdateItemActualSellPrice(ctx context.Context, itemID int64, runID int64, actualSellPriceISK float64) error
}

// HaulingWalletTxEsiClient provides ESI wallet transaction access.
type HaulingWalletTxEsiClient interface {
	GetCharacterWalletTransactions(ctx context.Context, characterID int64, token string) ([]*client.WalletTransaction, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

// HaulingWalletTxUpdater matches wallet sell transactions to hauling run items for all users.
type HaulingWalletTxUpdater struct {
	userRepo     HaulingWalletTxUserRepository
	charRepo     HaulingWalletTxCharacterRepository
	runsRepo     HaulingWalletTxRunsRepository
	runItemsRepo HaulingWalletTxItemsRepository
	esiClient    HaulingWalletTxEsiClient
}

// NewHaulingWalletTx creates a new HaulingWalletTxUpdater.
func NewHaulingWalletTx(
	userRepo HaulingWalletTxUserRepository,
	charRepo HaulingWalletTxCharacterRepository,
	runsRepo HaulingWalletTxRunsRepository,
	runItemsRepo HaulingWalletTxItemsRepository,
	esiClient HaulingWalletTxEsiClient,
) *HaulingWalletTxUpdater {
	return &HaulingWalletTxUpdater{
		userRepo:     userRepo,
		charRepo:     charRepo,
		runsRepo:     runsRepo,
		runItemsRepo: runItemsRepo,
		esiClient:    esiClient,
	}
}

// UpdateAllUsers processes wallet transactions for all users.
func (u *HaulingWalletTxUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for hauling wallet tx update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserTransactions(ctx, userID); err != nil {
			log.Error("failed to update hauling wallet transactions for user", "userID", userID, "error", err)
		}
	}

	return nil
}

// UpdateUserTransactions processes wallet sell transactions for a single user and updates actual sell prices.
func (u *HaulingWalletTxUpdater) UpdateUserTransactions(ctx context.Context, userID int64) error {
	// Get SELLING runs for this user
	runs, err := u.runsRepo.ListSellingByUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to list selling runs for wallet tx")
	}
	if len(runs) == 0 {
		return nil
	}

	// Get all characters for this user
	characters, err := u.charRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get characters for user (wallet tx)")
	}
	if len(characters) == 0 {
		return nil
	}

	// Collect all sell transactions from all characters
	// Map: typeID -> most recent unit price (sell tx after run created_at)
	type txInfo struct {
		typeID    int64
		unitPrice float64
		date      time.Time
	}
	allSellTx := []txInfo{}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-wallet.read_character_wallet.v1") {
			continue
		}

		token := char.EsiToken
		if time.Now().After(char.EsiTokenExpiresOn) {
			refreshed, err := u.esiClient.RefreshAccessToken(ctx, char.EsiRefreshToken)
			if err != nil {
				log.Error("failed to refresh token for character (hauling wallet tx)", "characterID", char.ID, "error", err)
				continue
			}
			token = refreshed.AccessToken
			if err := u.charRepo.UpdateTokens(ctx, char.ID, char.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry); err != nil {
				log.Error("failed to persist refreshed token for character (hauling wallet tx)", "characterID", char.ID, "error", err)
			}
		}

		transactions, err := u.esiClient.GetCharacterWalletTransactions(ctx, char.ID, token)
		if err != nil {
			log.Error("failed to get character wallet transactions", "characterID", char.ID, "error", err)
			continue
		}

		for _, tx := range transactions {
			if tx.IsBuy {
				continue
			}
			txDate, err := time.Parse(time.RFC3339, tx.Date)
			if err != nil {
				log.Error("failed to parse wallet transaction date", "transactionID", tx.TransactionID, "error", err)
				continue
			}
			allSellTx = append(allSellTx, txInfo{
				typeID:    tx.TypeID,
				unitPrice: tx.UnitPrice,
				date:      txDate,
			})
		}
	}

	if len(allSellTx) == 0 {
		return nil
	}

	// For each run, match sell transactions to run items and update actual sell price
	for _, run := range runs {
		runCreatedAt, err := time.Parse(time.RFC3339, run.CreatedAt)
		if err != nil {
			log.Error("failed to parse run created_at (wallet tx)", "runID", run.ID, "error", err)
			continue
		}

		items, err := u.runItemsRepo.GetItemsByRunID(ctx, run.ID)
		if err != nil {
			log.Error("failed to get items for hauling run (wallet tx)", "runID", run.ID, "error", err)
			continue
		}

		for _, item := range items {
			// Find the most recent sell transaction after run.CreatedAt for this type
			var bestPrice *float64
			var bestDate time.Time
			for _, tx := range allSellTx {
				if tx.typeID != item.TypeID {
					continue
				}
				if !tx.date.After(runCreatedAt) {
					continue
				}
				if bestPrice == nil || tx.date.After(bestDate) {
					p := tx.unitPrice
					bestPrice = &p
					bestDate = tx.date
				}
			}

			if bestPrice == nil {
				continue
			}

			// Only update if price changed
			if item.ActualSellPriceISK != nil && *item.ActualSellPriceISK == *bestPrice {
				continue
			}

			if err := u.runItemsRepo.UpdateItemActualSellPrice(ctx, item.ID, run.ID, *bestPrice); err != nil {
				log.Error("failed to update actual sell price for hauling run item",
					"itemID", item.ID,
					"runID", run.ID,
					"typeID", item.TypeID,
					"error", err,
				)
				continue
			}
			log.Info("updated hauling run item actual sell price from wallet tx",
				"runID", run.ID,
				"typeID", item.TypeID,
				"actualSellPrice", *bestPrice,
			)
		}
	}

	return nil
}
