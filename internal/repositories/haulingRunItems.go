package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingRunItems struct {
	db *sql.DB
}

func NewHaulingRunItems(db *sql.DB) *HaulingRunItems {
	return &HaulingRunItems{db: db}
}

// AddItem adds an item to a hauling run.
func (r *HaulingRunItems) AddItem(ctx context.Context, item *models.HaulingRunItem) (*models.HaulingRunItem, error) {
	query := `
		INSERT INTO hauling_run_items (run_id, type_id, type_name, quantity_planned, quantity_acquired,
			buy_price_isk, sell_price_isk, volume_m3, character_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`
	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		item.RunID, item.TypeID, item.TypeName, item.QuantityPlanned, item.QuantityAcquired,
		item.BuyPriceISK, item.SellPriceISK, item.VolumeM3, item.CharacterID, item.Notes,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add hauling run item")
	}
	item.ID = id
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return item, nil
}

// GetItemsByRunID returns all items for a hauling run, with computed fields.
func (r *HaulingRunItems) GetItemsByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunItem, error) {
	query := `
		SELECT id, run_id, type_id, type_name, quantity_planned, quantity_acquired,
		       buy_price_isk, sell_price_isk, volume_m3, character_id, notes,
		       sell_order_id, qty_sold, actual_sell_price_isk,
		       created_at, updated_at
		FROM hauling_run_items WHERE run_id=$1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hauling run items")
	}
	defer rows.Close()
	items := []*models.HaulingRunItem{}
	for rows.Next() {
		var item models.HaulingRunItem
		var buyPrice, sellPrice, volumeM3, actualSellPrice sql.NullFloat64
		var charID, sellOrderID sql.NullInt64
		var notes sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.TypeID, &item.TypeName, &item.QuantityPlanned, &item.QuantityAcquired,
			&buyPrice, &sellPrice, &volumeM3, &charID, &notes,
			&sellOrderID, &item.QtySold, &actualSellPrice,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan hauling run item")
		}
		if buyPrice.Valid {
			item.BuyPriceISK = &buyPrice.Float64
		}
		if sellPrice.Valid {
			item.SellPriceISK = &sellPrice.Float64
		}
		if volumeM3.Valid {
			item.VolumeM3 = &volumeM3.Float64
		}
		if charID.Valid {
			item.CharacterID = &charID.Int64
		}
		if notes.Valid {
			item.Notes = &notes.String
		}
		if sellOrderID.Valid {
			item.SellOrderID = &sellOrderID.Int64
		}
		if actualSellPrice.Valid {
			item.ActualSellPriceISK = &actualSellPrice.Float64
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		// Computed fields
		if item.QuantityPlanned > 0 {
			item.FillPercent = float64(item.QuantityAcquired) / float64(item.QuantityPlanned) * 100
			item.SellFillPercent = float64(item.QtySold) / float64(item.QuantityPlanned) * 100
		}
		if item.BuyPriceISK != nil && item.SellPriceISK != nil {
			net := (*item.SellPriceISK - *item.BuyPriceISK) * float64(item.QuantityPlanned)
			item.NetProfitISK = &net
		}
		if item.ActualSellPriceISK != nil && item.QtySold > 0 {
			rev := *item.ActualSellPriceISK * float64(item.QtySold)
			item.ActualRevenueISK = &rev
		}
		items = append(items, &item)
	}
	return items, nil
}

// UpdateItemSold updates qty_sold and optionally the sell_order_id for an item.
func (r *HaulingRunItems) UpdateItemSold(ctx context.Context, itemID int64, runID int64, qtySold int64, sellOrderID *int64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE hauling_run_items SET qty_sold=$1, sell_order_id=$2, updated_at=NOW() WHERE id=$3 AND run_id=$4`,
		qtySold, sellOrderID, itemID, runID)
	if err != nil {
		return errors.Wrap(err, "failed to update item sold quantity")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run item not found")
	}
	return nil
}

// UpdateItemActualSellPrice updates the actual_sell_price_isk for an item.
func (r *HaulingRunItems) UpdateItemActualSellPrice(ctx context.Context, itemID int64, runID int64, actualSellPriceISK float64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE hauling_run_items SET actual_sell_price_isk=$1, updated_at=NOW() WHERE id=$2 AND run_id=$3`,
		actualSellPriceISK, itemID, runID)
	if err != nil {
		return errors.Wrap(err, "failed to update item actual sell price")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run item not found")
	}
	return nil
}

// UpdateItemAcquired updates the quantity_acquired for an item.
func (r *HaulingRunItems) UpdateItemAcquired(ctx context.Context, itemID int64, runID int64, quantityAcquired int64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE hauling_run_items SET quantity_acquired=$1, updated_at=NOW() WHERE id=$2 AND run_id=$3`,
		quantityAcquired, itemID, runID)
	if err != nil {
		return errors.Wrap(err, "failed to update item acquired quantity")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run item not found")
	}
	return nil
}

// RemoveItem removes an item from a hauling run.
func (r *HaulingRunItems) RemoveItem(ctx context.Context, itemID int64, runID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM hauling_run_items WHERE id=$1 AND run_id=$2`, itemID, runID)
	if err != nil {
		return errors.Wrap(err, "failed to remove hauling run item")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run item not found")
	}
	return nil
}
