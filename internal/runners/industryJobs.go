package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type IndustryJobsUpdater interface {
	UpdateAllUsers(ctx context.Context) error
}

type IndustryJobsRunner struct {
	updater       IndustryJobsUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewIndustryJobsRunner(updater IndustryJobsUpdater, interval time.Duration) *IndustryJobsRunner {
	return &IndustryJobsRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *IndustryJobsRunner) WithTickerFactory(factory TickerFactory) *IndustryJobsRunner {
	r.tickerFactory = factory
	return r
}

func (r *IndustryJobsRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating industry jobs on startup")
	if err := r.updater.UpdateAllUsers(ctx); err != nil {
		log.Error("failed to update industry jobs on startup", "error", err)
	} else {
		log.Info("industry jobs updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating industry jobs (scheduled)")
			if err := r.updater.UpdateAllUsers(ctx); err != nil {
				log.Error("failed to update industry jobs", "error", err)
			} else {
				log.Info("industry jobs updated successfully")
			}
		}
	}
}
