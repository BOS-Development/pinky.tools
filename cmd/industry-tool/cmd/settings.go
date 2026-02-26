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
	PiUpdateIntervalSec              int
	SkillsUpdateIntervalSec          int
	IndustryJobsUpdateIntervalSec    int
	BlueprintsUpdateIntervalSec      int
	AutoProductionIntervalSec        int
	FrontendURL                      string
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

	if s := os.Getenv("SKILLS_UPDATE_INTERVAL_SEC"); s != "" {
		settings.SkillsUpdateIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "SKILLS_UPDATE_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.SkillsUpdateIntervalSec = 21600 // 6 hours
	}

	if s := os.Getenv("INDUSTRY_JOBS_UPDATE_INTERVAL_SEC"); s != "" {
		settings.IndustryJobsUpdateIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "INDUSTRY_JOBS_UPDATE_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.IndustryJobsUpdateIntervalSec = 600 // 10 minutes
	}

	if s := os.Getenv("BLUEPRINTS_UPDATE_INTERVAL_SEC"); s != "" {
		settings.BlueprintsUpdateIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "BLUEPRINTS_UPDATE_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.BlueprintsUpdateIntervalSec = 3600 // 1 hour
	}

	if s := os.Getenv("AUTO_PRODUCTION_INTERVAL_SEC"); s != "" {
		settings.AutoProductionIntervalSec, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.Wrapf(err, "AUTO_PRODUCTION_INTERVAL_SEC '%s' is not a number", s)
		}
	} else {
		settings.AutoProductionIntervalSec = 1800 // 30 minutes
	}

	settings.FrontendURL = os.Getenv("FRONTEND_URL")

	return settings, nil
}
