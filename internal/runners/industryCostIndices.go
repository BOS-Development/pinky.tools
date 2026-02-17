package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type IndustryCostIndicesUpdater interface {
	Update(ctx context.Context) error
}

type IndustryCostIndicesRunner struct {
	updater       IndustryCostIndicesUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewIndustryCostIndicesRunner(updater IndustryCostIndicesUpdater, interval time.Duration) *IndustryCostIndicesRunner {
	return &IndustryCostIndicesRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *IndustryCostIndicesRunner) WithTickerFactory(factory TickerFactory) *IndustryCostIndicesRunner {
	r.tickerFactory = factory
	return r
}

func (r *IndustryCostIndicesRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating industry cost indices on startup")
	if err := r.updater.Update(ctx); err != nil {
		log.Error("failed to update industry cost indices on startup", "error", err)
	} else {
		log.Info("industry cost indices updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating industry cost indices (scheduled)")
			if err := r.updater.Update(ctx); err != nil {
				log.Error("failed to update industry cost indices", "error", err)
			} else {
				log.Info("industry cost indices updated successfully")
			}
		}
	}
}
