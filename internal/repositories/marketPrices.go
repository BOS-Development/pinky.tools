package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type MarketPrices struct {
	db *sql.DB
}

func NewMarketPrices(db *sql.DB) *MarketPrices {
	return &MarketPrices{db: db}
}

func (r *MarketPrices) UpsertPrices(ctx context.Context, prices []models.MarketPrice) error {
	if len(prices) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	market_prices
	(
		type_id,
		region_id,
		buy_price,
		sell_price,
		daily_volume,
		updated_at
	)
	values
		($1,$2,$3,$4,$5,NOW())
on conflict
	(type_id)
do update set
	region_id = EXCLUDED.region_id,
	buy_price = EXCLUDED.buy_price,
	sell_price = EXCLUDED.sell_price,
	daily_volume = EXCLUDED.daily_volume,
	updated_at = NOW()
`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for market prices upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for market prices upsert")
	}

	for _, price := range prices {
		_, err = smt.ExecContext(ctx,
			price.TypeID,
			price.RegionID,
			price.BuyPrice,
			price.SellPrice,
			price.DailyVolume,
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute market price upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit market prices transaction")
	}

	return nil
}

func (r *MarketPrices) DeleteAllForRegion(ctx context.Context, regionID int64) error {
	query := `DELETE FROM market_prices WHERE region_id = $1`

	_, err := r.db.ExecContext(ctx, query, regionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete market prices for region")
	}

	return nil
}

func (r *MarketPrices) GetPricesForTypes(ctx context.Context, typeIDs []int64, regionID int64) (map[int64]*models.MarketPrice, error) {
	if len(typeIDs) == 0 {
		return map[int64]*models.MarketPrice{}, nil
	}

	query := `
SELECT
	type_id,
	region_id,
	buy_price,
	sell_price,
	daily_volume,
	updated_at
FROM
	market_prices
WHERE
	region_id = $1
	AND type_id = ANY($2)
`

	rows, err := r.db.QueryContext(ctx, query, regionID, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query market prices")
	}
	defer rows.Close()

	prices := make(map[int64]*models.MarketPrice)
	for rows.Next() {
		var price models.MarketPrice
		var updatedAt time.Time

		err := rows.Scan(
			&price.TypeID,
			&price.RegionID,
			&price.BuyPrice,
			&price.SellPrice,
			&price.DailyVolume,
			&updatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan market price row")
		}

		price.UpdatedAt = updatedAt.Format(time.RFC3339)
		prices[price.TypeID] = &price
	}

	return prices, nil
}

func (r *MarketPrices) GetLastUpdateTime(ctx context.Context, regionID int64) (*time.Time, error) {
	query := `
SELECT
	MAX(updated_at) as last_update
FROM
	market_prices
WHERE
	region_id = $1
`

	var lastUpdate *time.Time
	err := r.db.QueryRowContext(ctx, query, regionID).Scan(&lastUpdate)
	if err != nil {
		if err == sql.ErrNoRows {
			// No prices yet, return nil
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query last market price update time")
	}

	return lastUpdate, nil
}
