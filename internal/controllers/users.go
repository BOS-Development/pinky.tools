package controllers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type UserRepository interface {
	Get(ctx context.Context, id int64) (*repositories.User, error)
	Add(ctx context.Context, user *repositories.User) error
}

type UserAssetsStatusRepository interface {
	GetAssetsLastUpdated(ctx context.Context, userID int64) (*time.Time, error)
}

type Users struct {
	repository       UserRepository
	assetsStatusRepo UserAssetsStatusRepository
}

func NewUsers(router Routerer, repository UserRepository, assetsStatusRepo UserAssetsStatusRepository) *Users {
	controller := &Users{
		repository:       repository,
		assetsStatusRepo: assetsStatusRepo,
	}

	router.RegisterRestAPIRoute("/v1/users/asset-status", web.AuthAccessUser, controller.GetAssetStatus, "GET")
	router.RegisterRestAPIRoute("/v1/users/{id}", web.AuthAccessBackend, controller.GetUser, "GET")
	router.RegisterRestAPIRoute("/v1/users/", web.AuthAccessBackend, controller.AddUser, "POST")

	return controller
}

func (c *Users) GetUser(args *web.HandlerArgs) (any, *web.HttpError) {
	idS, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.New("Must provide user id"),
		}
	}

	id, err := strconv.Atoi(idS)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 400,
			Error:      errors.New("id must be a number"),
		}
	}

	user, err := c.repository.Get(args.Request.Context(), int64(id))
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get user from repository"),
		}
	}

	return user, nil
}

func (c *Users) AddUser(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var user repositories.User
	err := d.Decode(&user)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to decode user json"),
		}
	}

	err = c.repository.Add(args.Request.Context(), &user)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to insert user into database"),
		}
	}

	return nil, nil
}

type AssetStatusResponse struct {
	LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
	NextUpdateAt  *time.Time `json:"nextUpdateAt"`
}

func (c *Users) GetAssetStatus(args *web.HandlerArgs) (any, *web.HttpError) {
	lastUpdated, err := c.assetsStatusRepo.GetAssetsLastUpdated(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get asset status"),
		}
	}

	response := &AssetStatusResponse{
		LastUpdatedAt: lastUpdated,
	}

	if lastUpdated != nil {
		nextUpdate := lastUpdated.Add(1 * time.Hour)
		response.NextUpdateAt = &nextUpdate
	}

	return response, nil
}
