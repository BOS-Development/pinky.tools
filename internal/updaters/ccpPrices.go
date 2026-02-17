package updaters

import (
	"context"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/pkg/errors"
)

const CcpPricesUpdateInterval = 1 * time.Hour

type CcpPricesEsiClient interface {
	GetCcpMarketPrices(ctx context.Context) ([]*client.CcpMarketPrice, error)
}

type CcpPricesMarketRepo interface {
	UpsertAdjustedPrices(ctx context.Context, prices map[int64]float64) error
	GetAdjustedPriceLastUpdateTime(ctx context.Context) (*time.Time, error)
}

type CcpPrices struct {
	esiClient  CcpPricesEsiClient
	marketRepo CcpPricesMarketRepo
}

func NewCcpPrices(esiClient CcpPricesEsiClient, marketRepo CcpPricesMarketRepo) *CcpPrices {
	return &CcpPrices{
		esiClient:  esiClient,
		marketRepo: marketRepo,
	}
}

func (u *CcpPrices) Update(ctx context.Context) error {
	lastUpdate, err := u.marketRepo.GetAdjustedPriceLastUpdateTime(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last adjusted price update time")
	}

	if lastUpdate != nil {
		timeSinceUpdate := time.Since(*lastUpdate)
		if timeSinceUpdate < CcpPricesUpdateInterval {
			log.Info("skipping CCP adjusted prices update, last update was recent",
				"last_update", lastUpdate.Format(time.RFC3339),
				"next_update_in", (CcpPricesUpdateInterval - timeSinceUpdate).String())
			return nil
		}
	}

	log.Info("updating CCP adjusted prices")

	prices, err := u.esiClient.GetCcpMarketPrices(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch CCP market prices")
	}

	adjustedPrices := make(map[int64]float64)
	for _, p := range prices {
		if p.AdjustedPrice != nil {
			adjustedPrices[p.TypeID] = *p.AdjustedPrice
		}
	}

	if err := u.marketRepo.UpsertAdjustedPrices(ctx, adjustedPrices); err != nil {
		return errors.Wrap(err, "failed to upsert adjusted prices")
	}

	log.Info("CCP adjusted prices updated", "count", len(adjustedPrices))
	return nil
}
