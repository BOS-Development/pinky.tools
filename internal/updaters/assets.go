package updaters

import (
	"context"
	"time"

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
}

type PlayerCorporationRepository interface {
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
	UpsertDivisions(ctx context.Context, corp, user int64, divisions *models.CorporationDivisions) error
}

type EsiClient interface {
	GetCharacterAssets(ctx context.Context, characterID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error)
	GetCharacterLocationNames(ctx context.Context, characterID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error)
	GetPlayerOwnedStationInformation(ctx context.Context, token, refresh string, expire time.Time, ids []int64) ([]models.Station, error)
	GetCorporationAssets(ctx context.Context, corpID int64, token, refresh string, expire time.Time) ([]*models.EveAsset, error)
	GetCorporationLocationNames(ctx context.Context, corpID int64, token, refresh string, expire time.Time, ids []int64) (map[int64]string, error)
	GetCorporationDivisions(ctx context.Context, corpID int64, token, refresh string, expire time.Time) (*models.CorporationDivisions, error)
}

type Assets struct {
	characterRepository               CharacterRepository
	characterAssetsRepository         CharacterAssetsRepository
	stationRepository                 StationRepository
	playerCorporationsRepository      PlayerCorporationRepository
	playerCorporationAssetsRepository PlayerCorporationAssetsRepository
	esiClient                         EsiClient
}

func NewAssets(
	characterAssetsRepository CharacterAssetsRepository,
	characterRepository CharacterRepository,
	stationRepository StationRepository,
	playerCorporationsRepository PlayerCorporationRepository,
	playerCorporationAssetsRepository PlayerCorporationAssetsRepository,
	esiClient EsiClient) *Assets {
	return &Assets{
		characterAssetsRepository:         characterAssetsRepository,
		characterRepository:               characterRepository,
		stationRepository:                 stationRepository,
		playerCorporationsRepository:      playerCorporationsRepository,
		playerCorporationAssetsRepository: playerCorporationAssetsRepository,
		esiClient:                         esiClient,
	}
}

func (u *Assets) UpdateUserAssets(ctx context.Context, userID int64) error {
	characters, err := u.characterRepository.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user chars from repository")
	}

	for _, char := range characters {
		assets, err := u.esiClient.GetCharacterAssets(ctx, char.ID, char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn)
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

		containerNames, err := u.esiClient.GetCharacterLocationNames(ctx, char.ID, char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn, containers)
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

		stations, err := u.esiClient.GetPlayerOwnedStationInformation(ctx, char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn, playerOwnedStationIDs)
		if err != nil {
			return errors.Wrap(err, "failed to get player owned station information from esi")
		}

		err = u.stationRepository.Upsert(ctx, stations)
		if err != nil {
			return errors.Wrap(err, "failed to upsert player owned stations")
		}
	}

	corporations, err := u.playerCorporationsRepository.Get(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user player corporations")
	}

	for _, corp := range corporations {
		assets, err := u.esiClient.GetCorporationAssets(ctx, corp.ID, corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn)
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

		containerNames, err := u.esiClient.GetCorporationLocationNames(ctx, corp.ID, corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn, assembledContainers)
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

		stations, err := u.esiClient.GetPlayerOwnedStationInformation(ctx, corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn, stationIDs)
		if err != nil {
			return errors.Wrap(err, "failed to get player owned station information")
		}

		err = u.stationRepository.Upsert(ctx, stations)
		if err != nil {
			return errors.Wrap(err, "failed to upsert player owned stations")
		}

		divisions, err := u.esiClient.GetCorporationDivisions(ctx, corp.ID, corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn)
		if err != nil {
			return errors.Wrap(err, "failed to get corporation divisions from esi client")
		}

		err = u.playerCorporationsRepository.UpsertDivisions(ctx, corp.ID, userID, divisions)
		if err != nil {
			return errors.Wrap(err, "failed to upsert corporation divisions")
		}
	}

	return nil
}
