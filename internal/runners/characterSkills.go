package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type CharacterSkillsUpdater interface {
	UpdateAllUsers(ctx context.Context) error
}

type CharacterSkillsRunner struct {
	updater       CharacterSkillsUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewCharacterSkillsRunner(updater CharacterSkillsUpdater, interval time.Duration) *CharacterSkillsRunner {
	return &CharacterSkillsRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *CharacterSkillsRunner) WithTickerFactory(factory TickerFactory) *CharacterSkillsRunner {
	r.tickerFactory = factory
	return r
}

func (r *CharacterSkillsRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating character skills on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("failed to update character skills on startup", "error", err)
	} else {
		log.Info("character skills updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating character skills (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("failed to update character skills", "error", err)
			} else {
				log.Info("character skills updated successfully")
			}
		}
	}
}
