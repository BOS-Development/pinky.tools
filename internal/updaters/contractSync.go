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

type ContractSyncPurchaseRepository interface {
	GetContractCreatedWithKeys(ctx context.Context) ([]*models.PurchaseTransaction, error)
	CompleteWithContractID(ctx context.Context, purchaseID int64, eveContractID int64) error
}

type ContractSyncCharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type ContractSyncEsiClient interface {
	GetCharacterContracts(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*client.EsiContract, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type ContractSync struct {
	purchaseRepo  ContractSyncPurchaseRepository
	characterRepo ContractSyncCharacterRepository
	esiClient     ContractSyncEsiClient
}

func NewContractSync(
	purchaseRepo ContractSyncPurchaseRepository,
	characterRepo ContractSyncCharacterRepository,
	esiClient ContractSyncEsiClient,
) *ContractSync {
	return &ContractSync{
		purchaseRepo:  purchaseRepo,
		characterRepo: characterRepo,
		esiClient:     esiClient,
	}
}

// SyncAll checks ESI contracts for all buyers with contract_created purchases and auto-completes matches.
func (u *ContractSync) SyncAll(ctx context.Context) error {
	purchases, err := u.purchaseRepo.GetContractCreatedWithKeys(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get contract_created purchases")
	}

	if len(purchases) == 0 {
		return nil
	}

	// Group purchases by buyer_user_id and build contract_key lookup
	buyerPurchases := make(map[int64][]*models.PurchaseTransaction)
	keyToPurchases := make(map[string][]*models.PurchaseTransaction)
	for _, p := range purchases {
		buyerPurchases[p.BuyerUserID] = append(buyerPurchases[p.BuyerUserID], p)
		if p.ContractKey != nil {
			keyToPurchases[*p.ContractKey] = append(keyToPurchases[*p.ContractKey], p)
		}
	}

	// Collect all known contract keys for matching
	allKeys := make([]string, 0, len(keyToPurchases))
	for key := range keyToPurchases {
		allKeys = append(allKeys, key)
	}

	log.Info("contract sync: checking purchases", "pendingCount", len(purchases), "buyerCount", len(buyerPurchases))

	for buyerUserID := range buyerPurchases {
		if err := u.syncBuyer(ctx, buyerUserID, allKeys, keyToPurchases); err != nil {
			log.Error("contract sync: failed for buyer", "buyerUserID", buyerUserID, "error", err)
		}
	}

	return nil
}

func (u *ContractSync) syncBuyer(ctx context.Context, buyerUserID int64, allKeys []string, keyToPurchases map[string][]*models.PurchaseTransaction) error {
	characters, err := u.characterRepo.GetAll(ctx, buyerUserID)
	if err != nil {
		return errors.Wrap(err, "failed to get buyer characters")
	}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-contracts.read_character_contracts.v1") {
			continue
		}

		if err := u.syncCharacterContracts(ctx, char, buyerUserID, allKeys, keyToPurchases); err != nil {
			log.Error("contract sync: failed for character",
				"characterID", char.ID, "buyerUserID", buyerUserID, "error", err)
		}
	}

	return nil
}

func (u *ContractSync) syncCharacterContracts(
	ctx context.Context,
	char *repositories.Character,
	userID int64,
	allKeys []string,
	keyToPurchases map[string][]*models.PurchaseTransaction,
) error {
	token, refresh, expire := char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn

	if time.Now().After(expire) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, refresh)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken
		refresh = refreshed.RefreshToken
		expire = refreshed.Expiry

		if err := u.characterRepo.UpdateTokens(ctx, char.ID, userID, token, refresh, expire); err != nil {
			log.Error("contract sync: failed to persist refreshed token", "characterID", char.ID, "error", err)
		}
	}

	contracts, err := u.esiClient.GetCharacterContracts(ctx, char.ID, token, refresh, expire)
	if err != nil {
		return errors.Wrapf(err, "failed to get contracts for character %d", char.ID)
	}

	for _, contract := range contracts {
		if contract.Status != "finished" || contract.Type != "item_exchange" {
			continue
		}

		// Check if the contract title contains any known contract_key
		for _, key := range allKeys {
			if strings.Contains(contract.Title, key) {
				u.completePurchases(ctx, keyToPurchases[key], contract.ContractID)
				break
			}
		}
	}

	return nil
}

func (u *ContractSync) completePurchases(ctx context.Context, purchases []*models.PurchaseTransaction, eveContractID int64) {
	for _, purchase := range purchases {
		if err := u.purchaseRepo.CompleteWithContractID(ctx, purchase.ID, eveContractID); err != nil {
			log.Error("contract sync: failed to auto-complete purchase",
				"purchaseID", purchase.ID, "eveContractID", eveContractID, "error", err)
		} else {
			log.Info("contract sync: auto-completed purchase",
				"purchaseID", purchase.ID, "eveContractID", eveContractID,
				"buyerUserID", purchase.BuyerUserID, "contractKey", *purchase.ContractKey)
		}
	}
}
