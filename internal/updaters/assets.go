package updaters

import (
	"context"
	"sync"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"

	"github.com/pkg/errors"
)

type CharacterAssetsRepository interface {
	UpdateAssets(ctx context.Context, characterID, userID int64, assets []*models.EveAsset) error
	GetAssembledContainers(ctx context.Context, character, user int64) ([]int64, error)
	UpsertContainerNames(ctx context.Context, characterID, userID int64, locationNames map[int64]string) error
	GetPlayerOwnedStationIDs(ctx context.Context, character, user int64) ([]int64, error)
}

type PlayerCorporationAssetsRepository interface {
	Upsert(ctx context.Context, corp, user int64, assets []*models.EveAsset) error
	GetAssembledContainers(ctx context.Context, corp, user int64) ([]int64, error)
	UpsertContainerNames(ctx context.Context, corp, user int64, locationNames map[int64]string) error
	GetPlayerOwnedStationIDs(ctx context.Context, corp, user int64) ([]int64, error)
}

type CharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type PlayerCorporationRepository interface {
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
	UpsertDivisions(ctx context.Context, corp, user int64, divisions *models.CorporationDivisions) error
}

type StationRepository interface {
	Upsert(ctx context.Context, stations []models.Station) error
}

type UserTimestampRepository interface {
	UpdateAssetsLastUpdated(ctx context.Context, userID int64) error
}

type EsiClient interface {
	GetCharacterAssets(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error)
	GetCharacterLocationNames(ctx context.Context, characterID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error)
	GetPlayerOwnedStationInformation(ctx context.Context, token, refresh string, expire time.Time, ids []int64) ([]models.Station, error)
	GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error)
	GetCorporationLocationNames(ctx context.Context, corpID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error)
	GetCorporationDivisions(ctx context.Context, corpID int64, token, refresh string, expire time.Time) (*models.CorporationDivisions, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type AutoSellSyncer interface {
	SyncForUser(ctx context.Context, userID int64) error
}

type Assets struct {
	characterRepository               CharacterRepository
	characterAssetsRepository         CharacterAssetsRepository
	stationRepository                 StationRepository
	playerCorporationsRepository      PlayerCorporationRepository
	playerCorporationAssetsRepository PlayerCorporationAssetsRepository
	esiClient                         EsiClient
	userTimestampRepository           UserTimestampRepository
	autoSellSyncer                    AutoSellSyncer
	concurrency                       int
}

func NewAssets(
	characterAssetsRepository CharacterAssetsRepository,
	characterRepository CharacterRepository,
	stationRepository StationRepository,
	playerCorporationsRepository PlayerCorporationRepository,
	playerCorporationAssetsRepository PlayerCorporationAssetsRepository,
	esiClient EsiClient,
	userTimestampRepository UserTimestampRepository,
	concurrency int) *Assets {
	return &Assets{
		characterAssetsRepository:         characterAssetsRepository,
		characterRepository:               characterRepository,
		stationRepository:                 stationRepository,
		playerCorporationsRepository:      playerCorporationsRepository,
		playerCorporationAssetsRepository: playerCorporationAssetsRepository,
		esiClient:                         esiClient,
		userTimestampRepository:           userTimestampRepository,
		concurrency:                       concurrency,
	}
}

func (u *Assets) UpdateUserAssets(ctx context.Context, userID int64) error {
	characters, err := u.characterRepository.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user chars from repository")
	}

	corporations, err := u.playerCorporationsRepository.Get(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user player corporations")
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, u.concurrency)

	for _, char := range characters {
		char := char
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := u.UpdateCharacterAssets(ctx, char, userID); err != nil {
				log.Error("failed to update character assets", "characterID", char.ID, "error", err)
			}
		}()
	}

	for _, corp := range corporations {
		corp := corp
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := u.UpdateCorporationAssets(ctx, corp, userID); err != nil {
				log.Error("failed to update corporation assets", "corporationID", corp.ID, "error", err)
			}
		}()
	}

	wg.Wait()

	if u.autoSellSyncer != nil {
		if err := u.autoSellSyncer.SyncForUser(ctx, userID); err != nil {
			log.Error("failed to sync auto-sell listings after asset update", "userID", userID, "error", err)
		}
	}

	if err := u.userTimestampRepository.UpdateAssetsLastUpdated(ctx, userID); err != nil {
		log.Error("failed to update assets_last_updated_at", "userID", userID, "error", err)
	}

	return nil
}

// WithAutoSellUpdater sets the optional auto-sell syncer
func (u *Assets) WithAutoSellUpdater(syncer AutoSellSyncer) {
	u.autoSellSyncer = syncer
}

// UpdateCharacterAssets updates assets for a single character
func (u *Assets) UpdateCharacterAssets(ctx context.Context, char *repositories.Character, userID int64) error {
	token, refresh, expire := char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn

	if time.Now().After(expire) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, refresh)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken
		refresh = refreshed.RefreshToken
		expire = refreshed.Expiry

		err = u.characterRepository.UpdateTokens(ctx, char.ID, char.UserID, token, refresh, expire)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for character %d", char.ID)
		}
		log.Info("refreshed ESI token for character", "characterID", char.ID)
	}

	assets, err := u.esiClient.GetCharacterAssets(ctx, char.ID, token, refresh, expire)
	if err != nil {
		return errors.Wrap(err, "failed to get assets from the esi client")
	}

	err = u.characterAssetsRepository.UpdateAssets(ctx, char.ID, char.UserID, assets)
	if err != nil {
		return errors.Wrap(err, "failed to update assets in repository")
	}

	containers, err := u.characterAssetsRepository.GetAssembledContainers(ctx, char.ID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get character containers")
	}

	containerNames, err := u.esiClient.GetCharacterLocationNames(ctx, char.ID, token, refresh, expire, containers)
	if err != nil {
		return errors.Wrap(err, "failed to get character container names from esi")
	}

	err = u.characterAssetsRepository.UpsertContainerNames(ctx, char.ID, userID, containerNames)
	if err != nil {
		return errors.Wrap(err, "failed to upsert container location names")
	}

	playerOwnedStationIDs, err := u.characterAssetsRepository.GetPlayerOwnedStationIDs(ctx, char.ID, char.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve player owned station IDs")
	}

	stations, err := u.esiClient.GetPlayerOwnedStationInformation(ctx, token, refresh, expire, playerOwnedStationIDs)
	if err != nil {
		return errors.Wrap(err, "failed to get player owned station information from esi")
	}

	err = u.stationRepository.Upsert(ctx, stations)
	if err != nil {
		return errors.Wrap(err, "failed to upsert player owned stations")
	}

	return nil
}

// UpdateCorporationAssets updates assets for a single corporation
func (u *Assets) UpdateCorporationAssets(ctx context.Context, corp repositories.PlayerCorporation, userID int64) error {
	token, refresh, expire := corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn

	if time.Now().After(expire) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, refresh)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for corporation %d", corp.ID)
		}
		token = refreshed.AccessToken
		refresh = refreshed.RefreshToken
		expire = refreshed.Expiry

		err = u.playerCorporationsRepository.UpdateTokens(ctx, corp.ID, corp.UserID, token, refresh, expire)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for corporation %d", corp.ID)
		}
		log.Info("refreshed ESI token for corporation", "corporationID", corp.ID)
	}

	assets, err := u.esiClient.GetCorporationAssets(ctx, corp.ID, token, refresh, expire)
	if err != nil {
		return errors.Wrap(err, "failed to get corp assets")
	}

	err = u.playerCorporationAssetsRepository.Upsert(ctx, corp.ID, userID, assets)
	if err != nil {
		return errors.Wrap(err, "failed to upsert corp assets")
	}

	assembledContainers, err := u.playerCorporationAssetsRepository.GetAssembledContainers(ctx, corp.ID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get corp assembled containers")
	}

	containerNames, err := u.esiClient.GetCorporationLocationNames(ctx, corp.ID, token, refresh, expire, assembledContainers)
	if err != nil {
		return errors.Wrap(err, "failed to get corp location names")
	}

	err = u.playerCorporationAssetsRepository.UpsertContainerNames(ctx, corp.ID, userID, containerNames)
	if err != nil {
		return errors.Wrap(err, "failed to upsert corp location names")
	}

	stationIDs, err := u.playerCorporationAssetsRepository.GetPlayerOwnedStationIDs(ctx, corp.ID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get corp player owned stations id")
	}

	stations, err := u.esiClient.GetPlayerOwnedStationInformation(ctx, token, refresh, expire, stationIDs)
	if err != nil {
		return errors.Wrap(err, "failed to get player owned station information")
	}

	err = u.stationRepository.Upsert(ctx, stations)
	if err != nil {
		return errors.Wrap(err, "failed to upsert player owned stations")
	}

	divisions, err := u.esiClient.GetCorporationDivisions(ctx, corp.ID, token, refresh, expire)
	if err != nil {
		return errors.Wrap(err, "failed to get corporation divisions from esi client")
	}

	err = u.playerCorporationsRepository.UpsertDivisions(ctx, corp.ID, userID, divisions)
	if err != nil {
		return errors.Wrap(err, "failed to upsert corporation divisions")
	}

	return nil
}
