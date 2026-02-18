package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type AssetsUpdater interface {
	UpdateUserAssets(ctx context.Context, userID int64) error
}

type UserIDsProvider interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

type AssetsRunner struct {
	updater         AssetsUpdater
	userIDsProvider UserIDsProvider
	interval        time.Duration
	tickerFactory   TickerFactory
}

func NewAssetsRunner(updater AssetsUpdater, userIDsProvider UserIDsProvider, interval time.Duration) *AssetsRunner {
	return &AssetsRunner{
		updater:         updater,
		userIDsProvider: userIDsProvider,
		interval:        interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing
func (r *AssetsRunner) WithTickerFactory(factory TickerFactory) *AssetsRunner {
	r.tickerFactory = factory
	return r
}

func (r *AssetsRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Update immediately on startup
	log.Info("updating user assets on startup")
	r.updateAllUsers(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating user assets (scheduled)")
			r.updateAllUsers(ctx)
		}
	}
}

func (r *AssetsRunner) updateAllUsers(ctx context.Context) {
	userIDs, err := r.userIDsProvider.GetAllIDs(ctx)
	if err != nil {
		log.Error("failed to get user IDs for asset update", "error", err)
		return
	}

	for _, userID := range userIDs {
		if err := r.updater.UpdateUserAssets(ctx, userID); err != nil {
			log.Error("failed to update assets for user", "userID", userID, "error", err)
		} else {
			log.Info("assets updated successfully for user", "userID", userID)
		}
	}
}
