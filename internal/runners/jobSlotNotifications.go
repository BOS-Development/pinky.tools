package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

// JobSlotNotificationsUpdaterInterface is implemented by JobSlotNotificationsUpdater.
type JobSlotNotificationsUpdaterInterface interface {
	CheckAndNotifyCompletedJobs(ctx context.Context)
}

// JobSlotNotificationsRunner fires CheckAndNotifyCompletedJobs on a ticker interval.
type JobSlotNotificationsRunner struct {
	updater       JobSlotNotificationsUpdaterInterface
	interval      time.Duration
	tickerFactory TickerFactory
}

// NewJobSlotNotificationsRunner creates a new JobSlotNotificationsRunner.
func NewJobSlotNotificationsRunner(updater JobSlotNotificationsUpdaterInterface, interval time.Duration) *JobSlotNotificationsRunner {
	return &JobSlotNotificationsRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing.
func (r *JobSlotNotificationsRunner) WithTickerFactory(factory TickerFactory) *JobSlotNotificationsRunner {
	r.tickerFactory = factory
	return r
}

// Run starts the runner loop. It waits for the first tick before executing.
func (r *JobSlotNotificationsRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("job slot notifications: waiting for first scheduled tick")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("job slot notifications: checking completed jobs (scheduled)")
			r.updater.CheckAndNotifyCompletedJobs(ctx)
		}
	}
}
