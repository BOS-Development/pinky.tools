package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

// HaulingCorpOrdersUpdaterInterface is the interface for the hauling corp orders updater.
type HaulingCorpOrdersUpdaterInterface interface {
	UpdateAllUsers(ctx context.Context) error
}

// HaulingCorpOrdersRunner runs the corp orders updater on a schedule.
type HaulingCorpOrdersRunner struct {
	updater       HaulingCorpOrdersUpdaterInterface
	interval      time.Duration
	tickerFactory TickerFactory
}

// NewHaulingCorpOrdersRunner creates a new HaulingCorpOrdersRunner.
func NewHaulingCorpOrdersRunner(updater HaulingCorpOrdersUpdaterInterface, interval time.Duration) *HaulingCorpOrdersRunner {
	return &HaulingCorpOrdersRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing.
func (r *HaulingCorpOrdersRunner) WithTickerFactory(factory TickerFactory) *HaulingCorpOrdersRunner {
	r.tickerFactory = factory
	return r
}

// Run starts the corp orders runner loop.
func (r *HaulingCorpOrdersRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("hauling corp orders: running on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("hauling corp orders: failed on startup", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("hauling corp orders: running (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("hauling corp orders: failed", "error", err)
			}
		}
	}
}
