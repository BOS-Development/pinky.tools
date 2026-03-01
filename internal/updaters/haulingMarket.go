package updaters

import (
	"context"
	"log/slog"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingMarketEsiClient interface {
	GetMarketOrdersFiltered(ctx context.Context, regionID int64, systemID int64) ([]*client.MarketOrder, error)
	GetMarketHistory(ctx context.Context, regionID int64, typeID int64) ([]*client.MarketHistoryEntry, error)
}

type HaulingMarketRepository interface {
	UpsertSnapshots(ctx context.Context, snapshots []*models.HaulingMarketSnapshot) error
	GetSnapshotAge(ctx context.Context, regionID int64, systemID int64) (*time.Time, error)
}

type HaulingMarket struct {
	repo   HaulingMarketRepository
	esi    HaulingMarketEsiClient
	maxAge time.Duration
}

func NewHaulingMarket(repo HaulingMarketRepository, esi HaulingMarketEsiClient) *HaulingMarket {
	return &HaulingMarket{repo: repo, esi: esi, maxAge: 30 * time.Minute}
}

// ScanRegion fetches and caches market data for a region (or system within region).
// systemID=0 means region-wide.
func (u *HaulingMarket) ScanRegion(ctx context.Context, regionID int64, systemID int64) error {
	// Check cache freshness
	age, err := u.repo.GetSnapshotAge(ctx, regionID, systemID)
	if err != nil {
		return errors.Wrap(err, "failed to check snapshot age")
	}
	if age != nil && time.Since(*age) < u.maxAge {
		slog.Info("hauling market snapshot is fresh, skipping", "region_id", regionID, "system_id", systemID)
		return nil
	}

	slog.Info("scanning hauling market", "region_id", regionID, "system_id", systemID)

	// Fetch sell orders (we want to buy from source)
	orders, err := u.esi.GetMarketOrdersFiltered(ctx, regionID, systemID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch market orders")
	}

	// Group by type_id: best sell price (min), volume available
	type typeStats struct {
		minSellPrice float64
		maxBuyPrice  float64
		volume       int64
		hasSell      bool
		hasBuy       bool
	}
	stats := map[int64]*typeStats{}
	for _, o := range orders {
		s, ok := stats[o.TypeID]
		if !ok {
			s = &typeStats{}
			stats[o.TypeID] = s
		}
		if o.IsBuyOrder {
			if !s.hasBuy || o.Price > s.maxBuyPrice {
				s.maxBuyPrice = o.Price
				s.hasBuy = true
			}
		} else {
			s.volume += o.VolumeRemain
			if !s.hasSell || o.Price < s.minSellPrice {
				s.minSellPrice = o.Price
				s.hasSell = true
			}
		}
	}

	snapshots := []*models.HaulingMarketSnapshot{}
	for typeID, s := range stats {
		snap := &models.HaulingMarketSnapshot{
			TypeID:   typeID,
			RegionID: regionID,
			SystemID: systemID,
		}
		if s.hasSell {
			snap.SellPrice = &s.minSellPrice
			snap.VolumeAvailable = &s.volume
		}
		if s.hasBuy {
			snap.BuyPrice = &s.maxBuyPrice
		}
		snapshots = append(snapshots, snap)
	}

	if err := u.repo.UpsertSnapshots(ctx, snapshots); err != nil {
		return errors.Wrap(err, "failed to upsert snapshots")
	}

	slog.Info("hauling market scan complete", "region_id", regionID, "system_id", systemID, "types", len(snapshots))
	return nil
}

// ScanForHistory fetches market history to compute avg_daily_volume and days_to_sell for a set of type IDs.
func (u *HaulingMarket) ScanForHistory(ctx context.Context, regionID int64, typeIDs []int64) error {
	for _, typeID := range typeIDs {
		history, err := u.esi.GetMarketHistory(ctx, regionID, typeID)
		if err != nil {
			slog.Error("failed to fetch market history", "type_id", typeID, "error", err)
			continue
		}
		if len(history) == 0 {
			continue
		}
		// Compute avg daily volume over last 30 days
		days := history
		if len(days) > 30 {
			days = history[len(history)-30:]
		}
		var totalVol int64
		for _, d := range days {
			totalVol += d.Volume
		}
		avgDaily := float64(totalVol) / float64(len(days))
		// Update via upsert with avg_daily_volume
		snap := &models.HaulingMarketSnapshot{
			TypeID:         typeID,
			RegionID:       regionID,
			SystemID:       0,
			AvgDailyVolume: &avgDaily,
		}
		if err := u.repo.UpsertSnapshots(ctx, []*models.HaulingMarketSnapshot{snap}); err != nil {
			slog.Error("failed to update history snapshot", "type_id", typeID, "error", err)
		}
	}
	return nil
}
