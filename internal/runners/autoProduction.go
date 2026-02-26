package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type AutoProductionUpdaterInterface interface {
	RunAll(ctx context.Context) error
}

type AutoProductionRunner struct {
	updater       AutoProductionUpdaterInterface
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewAutoProductionRunner(updater AutoProductionUpdaterInterface, interval time.Duration) *AutoProductionRunner {
	return &AutoProductionRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing.
func (r *AutoProductionRunner) WithTickerFactory(factory TickerFactory) *AutoProductionRunner {
	r.tickerFactory = factory
	return r
}

// Run waits for the first scheduled tick before running. This avoids triggering
// auto-production before asset data has been populated after startup.
func (r *AutoProductionRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Do NOT run on startup — wait for asset data to be populated first
	log.Info("auto-production: waiting for first scheduled tick")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("auto-production: running (scheduled)")
			if err := r.updater.RunAll(ctx); err != nil {
				log.Error("auto-production: failed", "error", err)
			}
		}
	}
}
