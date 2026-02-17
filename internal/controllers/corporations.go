package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type PlayerCorporationRepository interface {
	Upsert(ctx context.Context, corp repositories.PlayerCorporation) error
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
}

type EsiClient interface {
	GetCharacterCorporation(ctx context.Context, characterID int64, token, refresh string, expire time.Time) (*models.Corporation, error)
}

type Corporations struct {
	router     Routerer
	esiClient  EsiClient
	repository PlayerCorporationRepository
}

func NewCorporations(router Routerer, esiClient EsiClient, repository PlayerCorporationRepository) *Corporations {
	controller := &Corporations{
		router:     router,
		esiClient:  esiClient,
		repository: repository,
	}

	router.RegisterRestAPIRoute("/v1/corporations", web.AuthAccessUser, controller.Add, "POST")
	router.RegisterRestAPIRoute("/v1/corporations", web.AuthAccessUser, controller.Get, "GET")

	return controller
}

func (c *Corporations) Add(args *web.HandlerArgs) (interface{}, *web.HttpError) {
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

	corp, err := c.esiClient.GetCharacterCorporation(args.Request.Context(), character.ID, character.EsiToken, character.EsiRefreshToken, character.EsiTokenExpiresOn)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get corp information from esi"),
		}
	}

	err = c.repository.Upsert(args.Request.Context(), repositories.PlayerCorporation{
		ID:              corp.ID,
		UserID:          *args.User,
		Name:            corp.Name,
		EsiToken:        character.EsiToken,
		EsiRefreshToken: character.EsiRefreshToken,
		EsiExpiresOn:    character.EsiTokenExpiresOn,
		EsiScopes:       character.EsiScopes,
	})
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to upsert corp to repository"),
		}
	}

	return nil, nil
}

type PlayerCorporation struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *Corporations) Get(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	corps, err := c.repository.Get(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get user corps from repository"),
		}
	}

	webCorps := []PlayerCorporation{}

	for _, corp := range corps {
		webCorps = append(webCorps, PlayerCorporation{
			ID:   corp.ID,
			Name: corp.Name,
		})
	}

	return webCorps, nil
}
