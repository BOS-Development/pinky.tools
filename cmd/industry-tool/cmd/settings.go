package cmd

import (
	"os"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/database"

	"github.com/pkg/errors"
)

type Settings struct {
	Port                   int
	BackendKey             string
	DatabaseSettings       database.PostgresDatabaseSettings
	OAuthClientID          string
	OAuthClientSecret      string
	EsiBaseURL             string
	AssetUpdateConcurrency int
	AssetUpdateIntervalSec int
	DiscordBotToken        string
	DiscordClientID        string
	DiscordClientSecret    string
	PiUpdateIntervalSec    int
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
	settings.EsiBaseURL = os.Getenv("ESI_BASE_URL")

	s = os.Getenv("DATABASE_PORT")
	settings.DatabaseSettings.Port, err = strconv.Atoi(s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert database port '%s' to number", s)
	}

	if s := os.Getenv("ASSET_UPDATE_CONCURRENCY"); s != "" {
		settings.AssetUpdateConcurrency, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "ASSET_UPDATE_CONCURRENCY '%s' is not a number", s)
		}
	} else {
		settings.AssetUpdateConcurrency = 5
	}

	if s := os.Getenv("ASSET_UPDATE_INTERVAL_SEC"); s != "" {
		settings.AssetUpdateIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "ASSET_UPDATE_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.AssetUpdateIntervalSec = 3600
	}

	settings.DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	settings.DiscordClientID = os.Getenv("DISCORD_CLIENT_ID")
	settings.DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")

	if s := os.Getenv("PI_UPDATE_INTERVAL_SEC"); s != "" {
		settings.PiUpdateIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "PI_UPDATE_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.PiUpdateIntervalSec = 3600
	}

	return settings, nil
}
