package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type CharacterBlueprintsUpdater interface {
	UpdateAllUsers(ctx context.Context) error
}

type CharacterBlueprintsRunner struct {
	updater       CharacterBlueprintsUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewCharacterBlueprintsRunner(updater CharacterBlueprintsUpdater, interval time.Duration) *CharacterBlueprintsRunner {
	return &CharacterBlueprintsRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *CharacterBlueprintsRunner) WithTickerFactory(factory TickerFactory) *CharacterBlueprintsRunner {
	r.tickerFactory = factory
	return r
}

func (r *CharacterBlueprintsRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating character blueprints on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("failed to update character blueprints on startup", "error", err)
	} else {
		log.Info("character blueprints updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating character blueprints (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("failed to update character blueprints", "error", err)
			} else {
				log.Info("character blueprints updated successfully")
			}
		}
	}
}
