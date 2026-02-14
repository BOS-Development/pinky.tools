package controllers

import (
	"context"

	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type StaticUpdater interface {
	Update(ctx context.Context) error
}

type Static struct {
	itemTypeUpdater StaticUpdater
}

func NewStatic(router Routerer, itemTypeUpdater StaticUpdater) *Static {
	controller := &Static{
		itemTypeUpdater: itemTypeUpdater,
	}

	router.RegisterRestAPIRoute("/v1/static/update", web.AuthAccessBackend, controller.UpdateStatics, "GET")

	return controller
}

func (c *Static) UpdateStatics(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	err := c.itemTypeUpdater.Update(args.Request.Context())
	if err != nil {
		return nil, &web.HttpError{
			StatusCode: 500,
			Error:      errors.Wrap(err, "failed to update item types"),
		}
	}
	return nil, nil
}
