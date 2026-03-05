package updaters

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingMarketEsiClient interface {
	GetMarketOrdersFiltered(ctx context.Context, regionID int64, systemID int64) ([]*client.MarketOrder, error)
	GetMarketHistory(ctx context.Context, regionID int64, typeID int64) ([]*client.MarketHistoryEntry, error)
	GetStructureMarketOrders(ctx context.Context, structureID int64, token string) ([]*client.MarketOrder, error)
}

type HaulingMarketRepository interface {
	UpsertSnapshots(ctx context.Context, snapshots []*models.HaulingMarketSnapshot) error
	GetSnapshotAge(ctx context.Context, regionID int64, systemID int64) (*time.Time, error)
	UpsertStructureSnapshots(ctx context.Context, structureID int64, snapshots []*models.HaulingMarketSnapshot) error
	GetStructureSnapshotAge(ctx context.Context, structureID int64) (*time.Time, error)
}

type HaulingMarket struct {
	repo     HaulingMarketRepository
	esi      HaulingMarketEsiClient
	maxAge   time.Duration
	regionMu sync.Map
}

func NewHaulingMarket(repo HaulingMarketRepository, esi HaulingMarketEsiClient) *HaulingMarket {
	return &HaulingMarket{repo: repo, esi: esi, maxAge: 30 * time.Minute}
}

// ScanRegion fetches and caches market data for a region (or system within region).
// systemID=0 means region-wide.
func (u *HaulingMarket) ScanRegion(ctx context.Context, regionID int64, systemID int64) error {
	// Serialize concurrent calls for the same region+system to make the
	// freshness check atomic with the upsert and avoid redundant ESI calls.
	lockKey := fmt.Sprintf("%d:%d", regionID, systemID)
	mu, _ := u.regionMu.LoadOrStore(lockKey, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()

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

	// Sort by TypeID so concurrent transactions acquire row locks in a
	// deterministic order, eliminating deadlocks.
	slices.SortFunc(snapshots, func(a, b *models.HaulingMarketSnapshot) int {
		return int(a.TypeID - b.TypeID)
	})

	if err := u.repo.UpsertSnapshots(ctx, snapshots); err != nil {
		return errors.Wrap(err, "failed to upsert snapshots")
	}

	slog.Info("hauling market scan complete", "region_id", regionID, "system_id", systemID, "types", len(snapshots))
	return nil
}

// ScanStructure fetches and caches market data for a player-owned structure.
// token is the character's ESI access token.
// Returns (accessOK bool, err error) — accessOK=false means 403 from ESI.
func (u *HaulingMarket) ScanStructure(ctx context.Context, structureID int64, token string) (bool, error) {
	// Check cache freshness (30 min TTL)
	age, err := u.repo.GetStructureSnapshotAge(ctx, structureID)
	if err != nil {
		return false, errors.Wrap(err, "failed to check structure snapshot age")
	}
	if age != nil && time.Since(*age) < u.maxAge {
		slog.Info("structure market snapshot is fresh, skipping", "structure_id", structureID)
		return true, nil
	}

	slog.Info("scanning structure market", "structure_id", structureID)

	orders, err := u.esi.GetStructureMarketOrders(ctx, structureID, token)
	if err != nil {
		return false, errors.Wrap(err, "failed to fetch structure market orders")
	}
	// nil orders means 403 — no access
	if orders == nil {
		return false, nil
	}

	// Group by type_id: best sell price (min), max buy price, volume available
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
			TypeID: typeID,
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

	// Sort by TypeID for deterministic lock ordering (deadlock prevention)
	slices.SortFunc(snapshots, func(a, b *models.HaulingMarketSnapshot) int {
		return int(a.TypeID - b.TypeID)
	})

	if err := u.repo.UpsertStructureSnapshots(ctx, structureID, snapshots); err != nil {
		return false, errors.Wrap(err, "failed to upsert structure snapshots")
	}

	slog.Info("structure market scan complete", "structure_id", structureID, "types", len(snapshots))
	return true, nil
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
