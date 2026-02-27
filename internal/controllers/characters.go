package controllers

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"

	"github.com/pkg/errors"
)

type CharacterModel struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	EsiScopes string `json:"esiScopes"`
}

type CharacterRepository interface {
	Get(ctx context.Context, id string) (*repositories.Character, error)
	Add(ctx context.Context, character *repositories.Character) error
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateCorporationID(ctx context.Context, characterID, userID, corporationID int64) error
}

type CharacterAssetUpdater interface {
	UpdateCharacterAssets(ctx context.Context, char *repositories.Character, userID int64) error
}

type CharacterEsiClient interface {
	GetCharacterCorporation(ctx context.Context, characterID int64, token, refresh string, expire time.Time) (*models.Corporation, error)
}

type CharacterContactRuleApplier interface {
	ApplyRulesForNewCorporation(ctx context.Context, userID int64, corpID int64, allianceID int64) error
}

type Characters struct {
	repository          CharacterRepository
	updater             CharacterAssetUpdater
	esiClient           CharacterEsiClient
	contactRulesApplier CharacterContactRuleApplier
}

func NewCharacters(router Routerer, repository CharacterRepository, updater CharacterAssetUpdater, esiClient CharacterEsiClient, contactRulesApplier CharacterContactRuleApplier) *Characters {
	controller := &Characters{
		repository:          repository,
		updater:             updater,
		esiClient:           esiClient,
		contactRulesApplier: contactRulesApplier,
	}

	router.RegisterRestAPIRoute("/v1/characters/{id}", web.AuthAccessUser, controller.GetCharacter, "GET")
	router.RegisterRestAPIRoute("/v1/characters/", web.AuthAccessUser, controller.AddCharacter, "POST")
	router.RegisterRestAPIRoute("/v1/characters/", web.AuthAccessUser, controller.GetAllCharacters, "GET")

	return controller
}

func (c *Characters) GetAllCharacters(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	characters, err := c.repository.GetAll(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get user characters from repository"),
		}
	}

	models := []CharacterModel{}
	for _, char := range characters {
		models = append(models, CharacterModel{
			ID:        char.ID,
			Name:      char.Name,
			EsiScopes: char.EsiScopes,
		})
	}
	return models, nil
}

func (c *Characters) GetCharacter(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.New("Must provide character id"),
		}
	}

	char, err := c.repository.Get(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "Failed to get character from repository"),
		}
	}

	if char == nil {
		return nil, &web.HttpError{
			StatusCode: 404,
			Error:      errors.New("character does not exist"),
		}
	}

	return char, nil
}

func (c *Characters) AddCharacter(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var character repositories.Character
	err := d.Decode(&character)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to decode json"),
		}
	}
	character.UserID = *args.User

	err = c.repository.Add(args.Request.Context(), &character)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to insert character"),
		}
	}

	go func() {
		ctx := context.Background()
		if err := c.updater.UpdateCharacterAssets(ctx, &character, character.UserID); err != nil {
			log.Error("failed to update assets after adding character", "characterID", character.ID, "error", err)
		}
		// Look up the character's corporation, store it, and apply any matching contact rules
		if c.esiClient != nil {
			corp, err := c.esiClient.GetCharacterCorporation(ctx, character.ID, character.EsiToken, character.EsiRefreshToken, character.EsiTokenExpiresOn)
			if err != nil {
				log.Error("failed to get character corporation", "characterID", character.ID, "error", err)
			} else {
				if err := c.repository.UpdateCorporationID(ctx, character.ID, character.UserID, corp.ID); err != nil {
					log.Error("failed to update character corporation_id", "characterID", character.ID, "corporationID", corp.ID, "error", err)
				}
				if c.contactRulesApplier != nil {
					allianceID := int64(0)
					if corp.AllianceID > 0 {
						allianceID = corp.AllianceID
					}
					if err := c.contactRulesApplier.ApplyRulesForNewCorporation(ctx, character.UserID, corp.ID, allianceID); err != nil {
						log.Error("failed to apply contact rules for new character", "characterID", character.ID, "corporationID", corp.ID, "error", err)
					}
				}
			}
		}
	}()

	return nil, nil
}
