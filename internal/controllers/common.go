package controllers

import (
	"github.com/annymsMthd/industry-tool/internal/web"
)

type Routerer interface {
	RegisterRestAPIRoute(path string, access web.AuthAccess, handler func(*web.HandlerArgs) (any, *web.HttpError), methods ...string)
}
