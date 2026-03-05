package repositories

import (
	"context"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
)

// HaulingMarketCombined combines HaulingMarket and HaulingStructures into a single
// adapter that satisfies the updaters.HaulingMarketRepository interface.
type HaulingMarketCombined struct {
	market     *HaulingMarket
	structures *HaulingStructures
}

func NewHaulingMarketCombined(market *HaulingMarket, structures *HaulingStructures) *HaulingMarketCombined {
	return &HaulingMarketCombined{market: market, structures: structures}
}

func (c *HaulingMarketCombined) UpsertSnapshots(ctx context.Context, snapshots []*models.HaulingMarketSnapshot) error {
	return c.market.UpsertSnapshots(ctx, snapshots)
}

func (c *HaulingMarketCombined) GetSnapshotAge(ctx context.Context, regionID int64, systemID int64) (*time.Time, error) {
	return c.market.GetSnapshotAge(ctx, regionID, systemID)
}

func (c *HaulingMarketCombined) UpsertStructureSnapshots(ctx context.Context, structureID int64, snapshots []*models.HaulingMarketSnapshot) error {
	return c.structures.UpsertStructureSnapshots(ctx, structureID, snapshots)
}

func (c *HaulingMarketCombined) GetStructureSnapshotAge(ctx context.Context, structureID int64) (*time.Time, error) {
	return c.structures.GetStructureSnapshotAge(ctx, structureID)
}
