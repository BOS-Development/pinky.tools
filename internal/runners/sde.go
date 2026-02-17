package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type SdeUpdater interface {
	Update(ctx context.Context) error
}

type SdeRunner struct {
	updater       SdeUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewSdeRunner(updater SdeUpdater, interval time.Duration) *SdeRunner {
	return &SdeRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *SdeRunner) WithTickerFactory(factory TickerFactory) *SdeRunner {
	r.tickerFactory = factory
	return r
}

func (r *SdeRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating SDE on startup")
	if err := r.updater.Update(ctx); err != nil {
		log.Error("failed to update SDE on startup", "error", err)
	} else {
		log.Info("SDE updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating SDE (scheduled)")
			if err := r.updater.Update(ctx); err != nil {
				log.Error("failed to update SDE", "error", err)
			} else {
				log.Info("SDE updated successfully")
			}
		}
	}
}
