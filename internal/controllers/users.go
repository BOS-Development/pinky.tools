package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type UserRepository interface {
	Get(ctx context.Context, id int64) (*repositories.User, error)
	Add(ctx context.Context, user *repositories.User) error
}

type Updater interface {
	UpdateUserAssets(ctx context.Context, userID int64) error
}

type Users struct {
	repository UserRepository
	updater    Updater
}

func NewUsers(router Routerer, repository UserRepository, updater Updater) *Users {
	controller := &Users{
		repository: repository,
		updater:    updater,
	}

	router.RegisterRestAPIRoute("/v1/users/refreshAssets", web.AuthAccessUser, controller.RefreshAssets, "GET")
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

func (c *Users) RefreshAssets(args *web.HandlerArgs) (any, *web.HttpError) {
	err := c.updater.UpdateUserAssets(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to update user assets"),
		}
	}
	return nil, nil
}
