package cmd

import (
	"os"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/database"

	"github.com/pkg/errors"
)

type Settings struct {
	Port              int
	BackendKey        string
	DatabaseSettings  database.PostgresDatabaseSettings
	OAuthClientID     string
	OAuthClientSecret string
}

func GetSettings() (*Settings, error) {
	settings := &Settings{}

	var err error

	s := os.Getenv("PORT")
	settings.Port, err = strconv.Atoi(s)
	if err != nil {
		return nil, errors.Wrapf(err, "port '%s' is not a number", s)
	}

	settings.BackendKey = os.Getenv("BACKEND_KEY")

	settings.DatabaseSettings.Host = os.Getenv("DATABASE_HOST")
	settings.DatabaseSettings.User = os.Getenv("DATABASE_USER")
	settings.DatabaseSettings.Password = os.Getenv("DATABASE_PASSWORD")
	settings.DatabaseSettings.Name = os.Getenv("DATABASE_NAME")
	settings.OAuthClientID = os.Getenv("OAUTH_CLIENT_ID")
	settings.OAuthClientSecret = os.Getenv("OAUTH_CLIENT_SECRET")

	s = os.Getenv("DATABASE_PORT")
	settings.DatabaseSettings.Port, err = strconv.Atoi(s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert database port '%s' to number", s)
	}

	return settings, nil
}
