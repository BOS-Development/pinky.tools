package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type PiUpdater interface {
	UpdateAllUsers(ctx context.Context) error
}

type PiRunner struct {
	updater       PiUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewPiRunner(updater PiUpdater, interval time.Duration) *PiRunner {
	return &PiRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing
func (r *PiRunner) WithTickerFactory(factory TickerFactory) *PiRunner {
	r.tickerFactory = factory
	return r
}

func (r *PiRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Update immediately on startup
	log.Info("updating PI planets on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("failed to update PI planets on startup", "error", err)
	} else {
		log.Info("PI planets updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating PI planets (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("failed to update PI planets", "error", err)
			} else {
				log.Info("PI planets updated successfully")
			}
		}
	}
}
