package controllers

import (
	"context"

	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type AssetsRepository interface {
	GetUserAssets(ctx context.Context, user int64) (*repositories.AssetsResponse, error)
	GetUserAssetsSummary(ctx context.Context, user int64) (*repositories.AssetsSummary, error)
}

type Assets struct {
	repository AssetsRepository
}

func NewAssets(router Routerer, repository AssetsRepository) *Assets {
	controller := &Assets{
		repository: repository,
	}

	router.RegisterRestAPIRoute("/v1/assets/", web.AuthAccessUser, controller.GetUserAssets, "GET")
	router.RegisterRestAPIRoute("/v1/assets/summary", web.AuthAccessUser, controller.GetUserAssetsSummary, "GET")

	return controller
}

func (c *Assets) GetUserAssets(args *web.HandlerArgs) (any, *web.HttpError) {
	assets, err := c.repository.GetUserAssets(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get user assets"),
		}
	}

	return assets, nil
}

func (c *Assets) GetUserAssetsSummary(args *web.HandlerArgs) (any, *web.HttpError) {
	summary, err := c.repository.GetUserAssetsSummary(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to get assets summary"),
		}
	}

	return summary, nil
}
