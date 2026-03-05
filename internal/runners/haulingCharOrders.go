package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

// HaulingCharOrdersUpdaterInterface is the interface for the hauling char orders updater.
type HaulingCharOrdersUpdaterInterface interface {
	UpdateAllUsers(ctx context.Context) error
}

// HaulingCharOrdersRunner runs the char orders updater on a schedule.
type HaulingCharOrdersRunner struct {
	updater       HaulingCharOrdersUpdaterInterface
	interval      time.Duration
	tickerFactory TickerFactory
}

// NewHaulingCharOrdersRunner creates a new HaulingCharOrdersRunner.
func NewHaulingCharOrdersRunner(updater HaulingCharOrdersUpdaterInterface, interval time.Duration) *HaulingCharOrdersRunner {
	return &HaulingCharOrdersRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing.
func (r *HaulingCharOrdersRunner) WithTickerFactory(factory TickerFactory) *HaulingCharOrdersRunner {
	r.tickerFactory = factory
	return r
}

// Run starts the char orders runner loop.
func (r *HaulingCharOrdersRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("hauling char orders: running on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("hauling char orders: failed on startup", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("hauling char orders: running (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("hauling char orders: failed", "error", err)
			}
		}
	}
}
