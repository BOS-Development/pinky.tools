package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type MarketPricesUpdater interface {
	UpdateJitaMarket(ctx context.Context) error
}

// Ticker interface allows mocking time.Ticker for testing
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// realTicker wraps time.Ticker to implement the Ticker interface
type realTicker struct {
	*time.Ticker
}

func (t *realTicker) C() <-chan time.Time {
	return t.Ticker.C
}

// TickerFactory creates a new Ticker
type TickerFactory func(d time.Duration) Ticker

type MarketPricesRunner struct {
	updater       MarketPricesUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewMarketPricesRunner(updater MarketPricesUpdater, interval time.Duration) *MarketPricesRunner {
	return &MarketPricesRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

// WithTickerFactory allows injecting a custom ticker factory for testing
func (r *MarketPricesRunner) WithTickerFactory(factory TickerFactory) *MarketPricesRunner {
	r.tickerFactory = factory
	return r
}

func (r *MarketPricesRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	// Update immediately on startup
	log.Info("updating market prices on startup")
	if err := r.updater.UpdateJitaMarket(ctx); err != nil {
		log.Error("failed to update market prices on startup", "error", err)
	} else {
		log.Info("market prices updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating market prices (scheduled)")
			if err := r.updater.UpdateJitaMarket(ctx); err != nil {
				log.Error("failed to update market prices", "error", err)
			} else {
				log.Info("market prices updated successfully")
			}
		}
	}
}
