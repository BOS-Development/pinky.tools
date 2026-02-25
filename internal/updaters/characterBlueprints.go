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

type BlueprintsCharacterBlueprintsRepository interface {
	ReplaceBlueprints(ctx context.Context, ownerID int64, ownerType string, userID int64, blueprints []*models.CharacterBlueprint) error
}

type BlueprintsEsiClient interface {
	GetCharacterBlueprints(ctx context.Context, characterID int64, token string) ([]*client.EsiBlueprint, error)
	GetCorporationBlueprints(ctx context.Context, corporationID int64, token string) ([]*client.EsiBlueprint, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type BlueprintsUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

type BlueprintsCharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type BlueprintsCorporationRepository interface {
	Get(ctx context.Context, userID int64) ([]repositories.PlayerCorporation, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type CharacterBlueprintsUpdater struct {
	userRepo      BlueprintsUserRepository
	characterRepo BlueprintsCharacterRepository
	corpRepo      BlueprintsCorporationRepository
	blueprintsRepo BlueprintsCharacterBlueprintsRepository
	esiClient     BlueprintsEsiClient
}

func NewCharacterBlueprintsUpdater(
	userRepo BlueprintsUserRepository,
	characterRepo BlueprintsCharacterRepository,
	corpRepo BlueprintsCorporationRepository,
	blueprintsRepo BlueprintsCharacterBlueprintsRepository,
	esiClient BlueprintsEsiClient,
) *CharacterBlueprintsUpdater {
	return &CharacterBlueprintsUpdater{
		userRepo:       userRepo,
		characterRepo:  characterRepo,
		corpRepo:       corpRepo,
		blueprintsRepo: blueprintsRepo,
		esiClient:      esiClient,
	}
}

func (u *CharacterBlueprintsUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for blueprints update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUser(ctx, userID); err != nil {
			log.Error("failed to update blueprints for user", "userID", userID, "error", err)
		}
	}

	return nil
}

func (u *CharacterBlueprintsUpdater) UpdateUser(ctx context.Context, userID int64) error {
	// Update character blueprints
	characters, err := u.characterRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user characters for blueprints update")
	}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-characters.read_blueprints.v1") {
			continue
		}

		if err := u.updateCharacterBlueprints(ctx, char, userID); err != nil {
			log.Error("failed to update blueprints for character", "characterID", char.ID, "error", err)
		}
	}

	// Update corporation blueprints
	corporations, err := u.corpRepo.Get(ctx, userID)
	if err != nil {
		log.Error("failed to get user corporations for blueprints update", "userID", userID, "error", err)
		return nil
	}

	for _, corp := range corporations {
		if !strings.Contains(corp.EsiScopes, "esi-corporations.read_blueprints.v1") {
			continue
		}

		if err := u.updateCorporationBlueprints(ctx, corp, userID); err != nil {
			log.Error("failed to update blueprints for corporation", "corporationID", corp.ID, "error", err)
		}
	}

	return nil
}

func (u *CharacterBlueprintsUpdater) updateCharacterBlueprints(ctx context.Context, char *repositories.Character, userID int64) error {
	token := char.EsiToken

	if time.Now().After(char.EsiTokenExpiresOn) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, char.EsiRefreshToken)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken

		err = u.characterRepo.UpdateTokens(ctx, char.ID, char.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for character %d", char.ID)
		}
		log.Info("refreshed ESI token for character (blueprints)", "characterID", char.ID)
	}

	esiBlueprints, err := u.esiClient.GetCharacterBlueprints(ctx, char.ID, token)
	if err != nil {
		return errors.Wrap(err, "failed to get character blueprints from ESI")
	}

	blueprints := convertEsiBlueprints(esiBlueprints)

	err = u.blueprintsRepo.ReplaceBlueprints(ctx, char.ID, "character", userID, blueprints)
	if err != nil {
		return errors.Wrap(err, "failed to replace character blueprints")
	}

	log.Info("updated character blueprints", "characterID", char.ID, "count", len(blueprints))
	return nil
}

func (u *CharacterBlueprintsUpdater) updateCorporationBlueprints(ctx context.Context, corp repositories.PlayerCorporation, userID int64) error {
	token := corp.EsiToken

	if time.Now().After(corp.EsiExpiresOn) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, corp.EsiRefreshToken)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for corporation %d", corp.ID)
		}
		token = refreshed.AccessToken

		err = u.corpRepo.UpdateTokens(ctx, corp.ID, corp.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for corporation %d", corp.ID)
		}
		log.Info("refreshed ESI token for corporation (blueprints)", "corporationID", corp.ID)
	}

	esiBlueprints, err := u.esiClient.GetCorporationBlueprints(ctx, corp.ID, token)
	if err != nil {
		return errors.Wrap(err, "failed to get corporation blueprints from ESI")
	}

	blueprints := convertEsiBlueprints(esiBlueprints)

	err = u.blueprintsRepo.ReplaceBlueprints(ctx, corp.ID, "corporation", userID, blueprints)
	if err != nil {
		return errors.Wrap(err, "failed to replace corporation blueprints")
	}

	log.Info("updated corporation blueprints", "corporationID", corp.ID, "count", len(blueprints))
	return nil
}

// convertEsiBlueprints converts ESI blueprint responses into domain models.
func convertEsiBlueprints(esiBlueprints []*client.EsiBlueprint) []*models.CharacterBlueprint {
	blueprints := []*models.CharacterBlueprint{}
	for _, eb := range esiBlueprints {
		blueprints = append(blueprints, &models.CharacterBlueprint{
			ItemID:             eb.ItemID,
			TypeID:             eb.TypeID,
			LocationID:         eb.LocationID,
			LocationFlag:       eb.LocationFlag,
			Quantity:           eb.Quantity,
			MaterialEfficiency: eb.MaterialEfficiency,
			TimeEfficiency:     eb.TimeEfficiency,
			Runs:               eb.Runs,
		})
	}
	return blueprints
}
