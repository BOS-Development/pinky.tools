package updaters

import (
	"context"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

const CostIndicesUpdateInterval = 1 * time.Hour

type CostIndicesEsiClient interface {
	GetIndustryCostIndices(ctx context.Context) ([]*client.IndustryCostIndexSystem, error)
}

type CostIndicesRepository interface {
	UpsertIndices(ctx context.Context, indices []models.IndustryCostIndex) error
	GetLastUpdateTime(ctx context.Context) (*time.Time, error)
}

type IndustryCostIndicesUpdater struct {
	esiClient CostIndicesEsiClient
	repo      CostIndicesRepository
}

func NewIndustryCostIndices(esiClient CostIndicesEsiClient, repo CostIndicesRepository) *IndustryCostIndicesUpdater {
	return &IndustryCostIndicesUpdater{
		esiClient: esiClient,
		repo:      repo,
	}
}

func (u *IndustryCostIndicesUpdater) Update(ctx context.Context) error {
	lastUpdate, err := u.repo.GetLastUpdateTime(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last cost indices update time")
	}

	if lastUpdate != nil {
		timeSinceUpdate := time.Since(*lastUpdate)
		if timeSinceUpdate < CostIndicesUpdateInterval {
			log.Info("skipping cost indices update, last update was recent",
				"last_update", lastUpdate.Format(time.RFC3339),
				"next_update_in", (CostIndicesUpdateInterval - timeSinceUpdate).String())
			return nil
		}
	}

	log.Info("updating industry cost indices")

	systems, err := u.esiClient.GetIndustryCostIndices(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch industry cost indices")
	}

	// Flatten to rows
	var indices []models.IndustryCostIndex
	for _, sys := range systems {
		for _, ci := range sys.CostIndices {
			indices = append(indices, models.IndustryCostIndex{
				SystemID:  sys.SolarSystemID,
				Activity:  ci.Activity,
				CostIndex: ci.CostIndex,
			})
		}
	}

	if err := u.repo.UpsertIndices(ctx, indices); err != nil {
		return errors.Wrap(err, "failed to upsert cost indices")
	}

	log.Info("industry cost indices updated", "systems", len(systems), "indices", len(indices))
	return nil
}
