package controllers

import (
	"context"
	"encoding/json"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"

	"github.com/pkg/errors"
)

type CharacterModel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CharacterRepository interface {
	Get(ctx context.Context, id string) (*repositories.Character, error)
	Add(ctx context.Context, character *repositories.Character) error
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
}

type CharacterAssetUpdater interface {
	UpdateCharacterAssets(ctx context.Context, char *repositories.Character, userID int64) error
}

type Characters struct {
	repository CharacterRepository
	updater    CharacterAssetUpdater
}

func NewCharacters(router Routerer, repository CharacterRepository, updater CharacterAssetUpdater) *Characters {
	controller := &Characters{
		repository: repository,
		updater:    updater,
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
			ID:   char.ID,
			Name: char.Name,
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
		if err := c.updater.UpdateCharacterAssets(context.Background(), &character, character.UserID); err != nil {
			log.Error("failed to update assets after adding character", "characterID", character.ID, "error", err)
		}
	}()

	return nil, nil
}
