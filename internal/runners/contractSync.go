package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type ContractSyncUpdater interface {
	SyncAll(ctx context.Context) error
}

type ContractSyncRunner struct {
	updater       ContractSyncUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewContractSyncRunner(updater ContractSyncUpdater, interval time.Duration) *ContractSyncRunner {
	return &ContractSyncRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing
func (r *ContractSyncRunner) WithTickerFactory(factory TickerFactory) *ContractSyncRunner {
	r.tickerFactory = factory
	return r
}

func (r *ContractSyncRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("contract sync: running on startup")
	if err := r.updater.SyncAll(ctx); err != nil {
		log.Error("contract sync: failed on startup", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("contract sync: running (scheduled)")
			if err := r.updater.SyncAll(ctx); err != nil {
				log.Error("contract sync: failed", "error", err)
			}
		}
	}
}
