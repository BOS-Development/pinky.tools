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
		       buy_price_isk, sell_price_isk, volume_m3, character_id, notes, created_at, updated_at
		FROM hauling_run_items WHERE run_id=$1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hauling run items")
	}
	defer rows.Close()
	items := []*models.HaulingRunItem{}
	for rows.Next() {
		var item models.HaulingRunItem
		var buyPrice, sellPrice, volumeM3 sql.NullFloat64
		var charID sql.NullInt64
		var notes sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.TypeID, &item.TypeName, &item.QuantityPlanned, &item.QuantityAcquired,
			&buyPrice, &sellPrice, &volumeM3, &charID, &notes, &createdAt, &updatedAt,
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
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		// Computed fields
		if item.QuantityPlanned > 0 {
			item.FillPercent = float64(item.QuantityAcquired) / float64(item.QuantityPlanned) * 100
		}
		if item.BuyPriceISK != nil && item.SellPriceISK != nil {
			net := (*item.SellPriceISK - *item.BuyPriceISK) * float64(item.QuantityPlanned)
			item.NetProfitISK = &net
		}
		items = append(items, &item)
	}
	return items, nil
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
