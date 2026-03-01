package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingMarket struct {
	db *sql.DB
}

func NewHaulingMarket(db *sql.DB) *HaulingMarket {
	return &HaulingMarket{db: db}
}

// UpsertSnapshots batch upserts market snapshots.
func (r *HaulingMarket) UpsertSnapshots(ctx context.Context, snapshots []*models.HaulingMarketSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO hauling_market_snapshots (type_id, region_id, system_id, buy_price, sell_price,
			volume_available, avg_daily_volume, days_to_sell, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		ON CONFLICT (type_id, region_id, system_id) DO UPDATE SET
			buy_price = EXCLUDED.buy_price,
			sell_price = EXCLUDED.sell_price,
			volume_available = EXCLUDED.volume_available,
			avg_daily_volume = EXCLUDED.avg_daily_volume,
			days_to_sell = EXCLUDED.days_to_sell,
			updated_at = NOW()`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare snapshot upsert")
	}
	defer stmt.Close()
	for _, s := range snapshots {
		if _, err := stmt.ExecContext(ctx, s.TypeID, s.RegionID, s.SystemID, s.BuyPrice, s.SellPrice,
			s.VolumeAvailable, s.AvgDailyVolume, s.DaysToSell); err != nil {
			return errors.Wrap(err, "failed to upsert snapshot")
		}
	}
	return errors.Wrap(tx.Commit(), "failed to commit snapshot upsert")
}

// GetScannerResults returns arbitrage opportunities between source and destination regions.
// sourceSystemID=0 means region-wide for source.
func (r *HaulingMarket) GetScannerResults(ctx context.Context, sourceRegionID int64, sourceSystemID int64, destRegionID int64) ([]*models.HaulingArbitrageRow, error) {
	query := `
		SELECT
			src.type_id,
			COALESCE(it.type_name, src.type_id::text) as type_name,
			it.volume as volume_m3,
			src.sell_price as buy_price,
			dst.buy_price as sell_price,
			src.volume_available,
			src.days_to_sell,
			src.updated_at
		FROM hauling_market_snapshots src
		JOIN hauling_market_snapshots dst ON dst.type_id = src.type_id AND dst.region_id = $3 AND dst.system_id = 0
		LEFT JOIN asset_item_types it ON it.type_id = src.type_id
		WHERE src.region_id = $1 AND src.system_id = $2
		  AND src.sell_price IS NOT NULL
		  AND dst.buy_price IS NOT NULL
		  AND dst.buy_price > src.sell_price
		ORDER BY (dst.buy_price - src.sell_price) * COALESCE(src.volume_available, 0) DESC`
	rows, err := r.db.QueryContext(ctx, query, sourceRegionID, sourceSystemID, destRegionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get scanner results")
	}
	defer rows.Close()
	results := []*models.HaulingArbitrageRow{}
	for rows.Next() {
		var row models.HaulingArbitrageRow
		var volumeM3, buyPrice, sellPrice sql.NullFloat64
		var volAvail sql.NullInt64
		var daysToSell sql.NullFloat64
		var updatedAt time.Time
		if err := rows.Scan(
			&row.TypeID, &row.TypeName, &volumeM3, &buyPrice, &sellPrice,
			&volAvail, &daysToSell, &updatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan arbitrage row")
		}
		if volumeM3.Valid {
			row.VolumeM3 = &volumeM3.Float64
		}
		if buyPrice.Valid {
			row.BuyPrice = &buyPrice.Float64
		}
		if sellPrice.Valid {
			row.SellPrice = &sellPrice.Float64
		}
		if volAvail.Valid {
			row.VolumeAvailable = &volAvail.Int64
		}
		if daysToSell.Valid {
			row.DaysToSell = &daysToSell.Float64
		}
		row.UpdatedAt = updatedAt.Format(time.RFC3339)
		// Compute net profit and spread
		if row.BuyPrice != nil && row.SellPrice != nil {
			net := *row.SellPrice - *row.BuyPrice // per unit
			row.NetProfitISK = &net
			spread := (*row.SellPrice / *row.BuyPrice) - 1
			row.Spread = &spread
			// Indicator
			if spread > 0.15 {
				row.Indicator = "gap"
			} else if spread > 0.05 {
				row.Indicator = "markup"
			} else {
				row.Indicator = "thin"
			}
		}
		results = append(results, &row)
	}
	return results, nil
}

// GetSnapshotAge returns the oldest updated_at for a region/system combo.
func (r *HaulingMarket) GetSnapshotAge(ctx context.Context, regionID int64, systemID int64) (*time.Time, error) {
	var t *time.Time
	err := r.db.QueryRowContext(ctx,
		`SELECT MIN(updated_at) FROM hauling_market_snapshots WHERE region_id=$1 AND system_id=$2`,
		regionID, systemID,
	).Scan(&t)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to get snapshot age")
	}
	return t, nil
}
