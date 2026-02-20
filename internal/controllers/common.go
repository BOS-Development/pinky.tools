package controllers

import (
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/web"
)

type Routerer interface {
	RegisterRestAPIRoute(path string, access web.AuthAccess, handler func(*web.HandlerArgs) (any, *web.HttpError), methods ...string)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
