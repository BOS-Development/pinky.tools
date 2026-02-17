package runners

import (
	"context"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"
)

type CcpPricesUpdater interface {
	Update(ctx context.Context) error
}

type CcpPricesRunner struct {
	updater       CcpPricesUpdater
	interval      time.Duration
	tickerFactory TickerFactory
}

func NewCcpPricesRunner(updater CcpPricesUpdater, interval time.Duration) *CcpPricesRunner {
	return &CcpPricesRunner{
		updater:  updater,
		interval: interval,
		tickerFactory: func(d time.Duration) Ticker {
			return &realTicker{time.NewTicker(d)}
		},
	}
}

func (r *CcpPricesRunner) WithTickerFactory(factory TickerFactory) *CcpPricesRunner {
	r.tickerFactory = factory
	return r
}

func (r *CcpPricesRunner) Run(ctx context.Context) error {
	ticker := r.tickerFactory(r.interval)
	defer ticker.Stop()

	log.Info("updating CCP adjusted prices on startup")
	if err := r.updater.Update(ctx); err != nil {
		log.Error("failed to update CCP adjusted prices on startup", "error", err)
	} else {
		log.Info("CCP adjusted prices updated successfully")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C():
			log.Info("updating CCP adjusted prices (scheduled)")
			if err := r.updater.Update(ctx); err != nil {
				log.Error("failed to update CCP adjusted prices", "error", err)
			} else {
				log.Info("CCP adjusted prices updated successfully")
			}
		}
	}
}
