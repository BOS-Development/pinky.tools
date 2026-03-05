package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingStructures struct {
	db *sql.DB
}

func NewHaulingStructures(db *sql.DB) *HaulingStructures {
	return &HaulingStructures{db: db}
}

// UpsertStructureSnapshots batch upserts structure market snapshots into hauling_structure_snapshots.
func (r *HaulingStructures) UpsertStructureSnapshots(ctx context.Context, structureID int64, snapshots []*models.HaulingMarketSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO hauling_structure_snapshots (type_id, structure_id, buy_price, sell_price, volume_available, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (type_id, structure_id) DO UPDATE SET
			buy_price = EXCLUDED.buy_price,
			sell_price = EXCLUDED.sell_price,
			volume_available = EXCLUDED.volume_available,
			updated_at = NOW()`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare structure snapshot upsert")
	}
	defer stmt.Close()

	for _, s := range snapshots {
		if _, err := stmt.ExecContext(ctx, s.TypeID, structureID, s.BuyPrice, s.SellPrice, s.VolumeAvailable); err != nil {
			return errors.Wrap(err, "failed to upsert structure snapshot")
		}
	}
	return errors.Wrap(tx.Commit(), "failed to commit structure snapshot upsert")
}

// GetStructureSnapshotAge returns MIN(updated_at) for a structure's snapshots.
func (r *HaulingStructures) GetStructureSnapshotAge(ctx context.Context, structureID int64) (*time.Time, error) {
	var t *time.Time
	err := r.db.QueryRowContext(ctx,
		`SELECT MIN(updated_at) FROM hauling_structure_snapshots WHERE structure_id=$1`,
		structureID,
	).Scan(&t)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to get structure snapshot age")
	}
	return t, nil
}

// GetStructureScannerResults returns arbitrage opportunities when SOURCE is a structure.
// Joins hauling_structure_snapshots (src) with hauling_market_snapshots (dst).
func (r *HaulingStructures) GetStructureScannerResults(ctx context.Context, sourceStructureID int64, destRegionID int64, destSystemID int64) ([]*models.HaulingArbitrageRow, error) {
	query := `
		SELECT
			src.type_id,
			COALESCE(it.type_name, src.type_id::text) as type_name,
			it.volume as volume_m3,
			src.sell_price as buy_price,
			dst.buy_price as sell_price,
			src.volume_available,
			src.updated_at
		FROM hauling_structure_snapshots src
		JOIN hauling_market_snapshots dst ON dst.type_id = src.type_id AND dst.region_id = $2 AND dst.system_id = $3
		LEFT JOIN asset_item_types it ON it.type_id = src.type_id
		WHERE src.structure_id = $1
		  AND src.sell_price IS NOT NULL
		  AND dst.buy_price IS NOT NULL
		  AND dst.buy_price > src.sell_price
		ORDER BY (dst.buy_price - src.sell_price) * COALESCE(src.volume_available, 0) DESC`

	rows, err := r.db.QueryContext(ctx, query, sourceStructureID, destRegionID, destSystemID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get structure scanner results")
	}
	defer rows.Close()
	return scanArbitrageRows(rows)
}

// GetRegionToStructureResults returns arbitrage opportunities when DESTINATION is a structure.
// Joins hauling_market_snapshots (src) with hauling_structure_snapshots (dst).
func (r *HaulingStructures) GetRegionToStructureResults(ctx context.Context, srcRegionID int64, srcSystemID int64, destStructureID int64) ([]*models.HaulingArbitrageRow, error) {
	query := `
		SELECT
			src.type_id,
			COALESCE(it.type_name, src.type_id::text) as type_name,
			it.volume as volume_m3,
			src.sell_price as buy_price,
			dst.buy_price as sell_price,
			src.volume_available,
			src.updated_at
		FROM hauling_market_snapshots src
		JOIN hauling_structure_snapshots dst ON dst.type_id = src.type_id AND dst.structure_id = $3
		LEFT JOIN asset_item_types it ON it.type_id = src.type_id
		WHERE src.region_id = $1 AND src.system_id = $2
		  AND src.sell_price IS NOT NULL
		  AND dst.buy_price IS NOT NULL
		  AND dst.buy_price > src.sell_price
		ORDER BY (dst.buy_price - src.sell_price) * COALESCE(src.volume_available, 0) DESC`

	rows, err := r.db.QueryContext(ctx, query, srcRegionID, srcSystemID, destStructureID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get region-to-structure results")
	}
	defer rows.Close()
	return scanArbitrageRows(rows)
}

// scanArbitrageRows scans a result set of arbitrage rows (shared by both structure scanner methods).
func scanArbitrageRows(rows *sql.Rows) ([]*models.HaulingArbitrageRow, error) {
	results := []*models.HaulingArbitrageRow{}
	for rows.Next() {
		var row models.HaulingArbitrageRow
		var volumeM3, buyPrice, sellPrice sql.NullFloat64
		var volAvail sql.NullInt64
		var updatedAt time.Time
		if err := rows.Scan(
			&row.TypeID, &row.TypeName, &volumeM3, &buyPrice, &sellPrice,
			&volAvail, &updatedAt,
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
		row.UpdatedAt = updatedAt.Format(time.RFC3339)
		// Compute net profit and spread
		if row.BuyPrice != nil && row.SellPrice != nil {
			net := *row.SellPrice - *row.BuyPrice
			row.NetProfitISK = &net
			spread := (*row.SellPrice / *row.BuyPrice) - 1
			row.Spread = &spread
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
