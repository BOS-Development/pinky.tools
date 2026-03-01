package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
)

// HaulingDigestUserRepository provides user IDs for digest processing.
type HaulingDigestUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

// HaulingDigestRunsRepository provides runs with daily_digest=true.
type HaulingDigestRunsRepository interface {
	ListDigestRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error)
}

// HaulingDigestNotifier sends daily digest notifications.
type HaulingDigestNotifier interface {
	SendHaulingDailyDigest(ctx context.Context, userID int64, runs []*models.HaulingRun)
}

// HaulingDigestRunner sends daily digest notifications to users with active digest-enabled runs.
type HaulingDigestRunner struct {
	userRepo      HaulingDigestUserRepository
	runsRepo      HaulingDigestRunsRepository
	notifier      HaulingDigestNotifier
	interval      time.Duration
	tickerFactory TickerFactory
}

// NewHaulingDigestRunner creates a new HaulingDigestRunner.
func NewHaulingDigestRunner(
	notifier HaulingDigestNotifier,
	runsRepo HaulingDigestRunsRepository,
	userRepo HaulingDigestUserRepository,
	interval time.Duration,
) *HaulingDigestRunner {
	return &HaulingDigestRunner{
		userRepo:      userRepo,
		runsRepo:      runsRepo,
		notifier:      notifier,
		interval:      interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing.
func (r *HaulingDigestRunner) WithTickerFactory(factory TickerFactory) *HaulingDigestRunner {
	r.tickerFactory = factory
	return r
}

// Run starts the digest runner loop.
func (r *HaulingDigestRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("hauling digest: running on startup")
	r.sendDigests(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("hauling digest: running (scheduled)")
			r.sendDigests(ctx)
		}
	}
}

func (r *HaulingDigestRunner) sendDigests(ctx context.Context) {
	userIDs, err := r.userRepo.GetAllIDs(ctx)
	if err != nil {
		log.Error("hauling digest: failed to get user IDs", "error", err)
		return
	}

	for _, userID := range userIDs {
		runs, err := r.runsRepo.ListDigestRunsByUser(ctx, userID)
		if err != nil {
			log.Error("hauling digest: failed to list digest runs", "userID", userID, "error", err)
			continue
		}
		if len(runs) == 0 {
			continue
		}
		r.notifier.SendHaulingDailyDigest(ctx, userID, runs)
	}
}
