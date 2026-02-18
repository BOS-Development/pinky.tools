package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type ForSaleItems struct {
	db *sql.DB
}

func NewForSaleItems(db *sql.DB) *ForSaleItems {
	return &ForSaleItems{db: db}
}

// GetUserIDByCharacterID converts a character ID to a user ID
func (r *ForSaleItems) GetUserIDByCharacterID(ctx context.Context, characterID int64) (int64, error) {
	query := `SELECT user_id FROM characters WHERE id = $1`
	var userID int64
	err := r.db.QueryRowContext(ctx, query, characterID).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, errors.New("character not found")
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to get user ID for character")
	}
	return userID, nil
}

// GetByUser returns all active for-sale items owned by userID
func (r *ForSaleItems) GetByUser(ctx context.Context, userID int64) ([]*models.ForSaleItem, error) {
	query := `
		SELECT
			f.id,
			f.user_id,
			f.type_id,
			t.type_name,
			f.owner_type,
			f.owner_id,
			CASE
				WHEN f.owner_type = 'character' THEN c.name
				WHEN f.owner_type = 'corporation' THEN corp.name
				ELSE 'Unknown'
			END AS owner_name,
			f.location_id,
			COALESCE(s.name, st.name, 'Unknown Location') AS location_name,
			f.container_id,
			f.division_number,
			f.quantity_available,
			f.price_per_unit,
			f.notes,
			f.auto_sell_container_id,
			f.is_active,
			f.created_at,
			f.updated_at
		FROM for_sale_items f
		JOIN asset_item_types t ON f.type_id = t.type_id
		LEFT JOIN characters c ON f.owner_type = 'character' AND f.owner_id = c.id
		LEFT JOIN player_corporations corp ON f.owner_type = 'corporation' AND f.owner_id = corp.id
		LEFT JOIN solar_systems s ON f.location_id = s.solar_system_id
		LEFT JOIN stations st ON f.location_id = st.station_id
		WHERE f.user_id = $1 AND f.is_active = true
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for-sale items")
	}
	defer rows.Close()

	var items []*models.ForSaleItem
	for rows.Next() {
		var item models.ForSaleItem
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TypeID,
			&item.TypeName,
			&item.OwnerType,
			&item.OwnerID,
			&item.OwnerName,
			&item.LocationID,
			&item.LocationName,
			&item.ContainerID,
			&item.DivisionNumber,
			&item.QuantityAvailable,
			&item.PricePerUnit,
			&item.Notes,
			&item.AutoSellContainerID,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan for-sale item")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetBrowsableItems returns items from sellerUserIDs that buyerUserID can browse
func (r *ForSaleItems) GetBrowsableItems(ctx context.Context, buyerUserID int64, sellerUserIDs []int64) ([]*models.ForSaleItem, error) {
	if len(sellerUserIDs) == 0 {
		return []*models.ForSaleItem{}, nil
	}

	query := `
		SELECT
			f.id,
			f.user_id,
			f.type_id,
			t.type_name,
			f.owner_type,
			f.owner_id,
			CASE
				WHEN f.owner_type = 'character' THEN c.name
				WHEN f.owner_type = 'corporation' THEN corp.name
				ELSE 'Unknown'
			END AS owner_name,
			f.location_id,
			COALESCE(s.name, st.name, 'Unknown Location') AS location_name,
			f.container_id,
			f.division_number,
			f.quantity_available,
			f.price_per_unit,
			f.notes,
			f.auto_sell_container_id,
			f.is_active,
			f.created_at,
			f.updated_at
		FROM for_sale_items f
		JOIN asset_item_types t ON f.type_id = t.type_id
		LEFT JOIN characters c ON f.owner_type = 'character' AND f.owner_id = c.id
		LEFT JOIN player_corporations corp ON f.owner_type = 'corporation' AND f.owner_id = corp.id
		LEFT JOIN solar_systems s ON f.location_id = s.solar_system_id
		LEFT JOIN stations st ON f.location_id = st.station_id
		WHERE f.user_id = ANY($1) AND f.is_active = true
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(sellerUserIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query browsable items")
	}
	defer rows.Close()

	var items []*models.ForSaleItem
	for rows.Next() {
		var item models.ForSaleItem
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TypeID,
			&item.TypeName,
			&item.OwnerType,
			&item.OwnerID,
			&item.OwnerName,
			&item.LocationID,
			&item.LocationName,
			&item.ContainerID,
			&item.DivisionNumber,
			&item.QuantityAvailable,
			&item.PricePerUnit,
			&item.Notes,
			&item.AutoSellContainerID,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan browsable item")
		}
		items = append(items, &item)
	}

	return items, nil
}

// Upsert creates or updates a for-sale listing
func (r *ForSaleItems) Upsert(ctx context.Context, item *models.ForSaleItem) error {
	query := `
		INSERT INTO for_sale_items
		(user_id, type_id, owner_type, owner_id, location_id, container_id, division_number,
		 quantity_available, price_per_unit, notes, auto_sell_container_id, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id, type_id, owner_type, owner_id, location_id, COALESCE(container_id, 0), COALESCE(division_number, 0))
		WHERE is_active = true
		DO UPDATE SET
			quantity_available = EXCLUDED.quantity_available,
			price_per_unit = EXCLUDED.price_per_unit,
			notes = EXCLUDED.notes,
			auto_sell_container_id = EXCLUDED.auto_sell_container_id,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.UserID,
		item.TypeID,
		item.OwnerType,
		item.OwnerID,
		item.LocationID,
		item.ContainerID,
		item.DivisionNumber,
		item.QuantityAvailable,
		item.PricePerUnit,
		item.Notes,
		item.AutoSellContainerID,
		item.IsActive,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to upsert for-sale item")
	}

	return nil
}

// Delete soft-deletes (sets is_active = false)
func (r *ForSaleItems) Delete(ctx context.Context, itemID int64, userID int64) error {
	query := `
		UPDATE for_sale_items
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, itemID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete for-sale item")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("for-sale item not found or user is not the owner")
	}

	return nil
}

// UpdateQuantity decreases quantity after purchase (within transaction)
func (r *ForSaleItems) UpdateQuantity(ctx context.Context, tx *sql.Tx, itemID int64, newQuantity int64) error {
	// When newQuantity <= 0, only mark as inactive without updating quantity
	// to avoid violating the for_sale_positive_quantity constraint
	query := `
		UPDATE for_sale_items
		SET quantity_available = CASE
		                           WHEN $2 > 0 THEN $2
		                           ELSE quantity_available
		                         END,
		    is_active = ($2::bigint > 0),
		    updated_at = NOW()
		WHERE id = $1 AND is_active = true
	`

	result, err := tx.ExecContext(ctx, query, itemID, newQuantity)
	if err != nil {
		return errors.Wrap(err, "failed to update for-sale item quantity")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("for-sale item not found or no longer active")
	}

	return nil
}

// GetByID returns a specific for-sale item
func (r *ForSaleItems) GetByID(ctx context.Context, itemID int64) (*models.ForSaleItem, error) {
	query := `
		SELECT
			f.id,
			f.user_id,
			f.type_id,
			t.type_name,
			f.owner_type,
			f.owner_id,
			CASE
				WHEN f.owner_type = 'character' THEN c.name
				WHEN f.owner_type = 'corporation' THEN corp.name
				ELSE 'Unknown'
			END AS owner_name,
			f.location_id,
			COALESCE(s.name, st.name, 'Unknown Location') AS location_name,
			f.container_id,
			f.division_number,
			f.quantity_available,
			f.price_per_unit,
			f.notes,
			f.auto_sell_container_id,
			f.is_active,
			f.created_at,
			f.updated_at
		FROM for_sale_items f
		JOIN asset_item_types t ON f.type_id = t.type_id
		LEFT JOIN characters c ON f.owner_type = 'character' AND f.owner_id = c.id
		LEFT JOIN player_corporations corp ON f.owner_type = 'corporation' AND f.owner_id = corp.id
		LEFT JOIN solar_systems s ON f.location_id = s.solar_system_id
		LEFT JOIN stations st ON f.location_id = st.station_id
		WHERE f.id = $1
	`

	var item models.ForSaleItem
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.UserID,
		&item.TypeID,
		&item.TypeName,
		&item.OwnerType,
		&item.OwnerID,
		&item.OwnerName,
		&item.LocationID,
		&item.LocationName,
		&item.ContainerID,
		&item.DivisionNumber,
		&item.QuantityAvailable,
		&item.PricePerUnit,
		&item.Notes,
		&item.AutoSellContainerID,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("for-sale item not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get for-sale item")
	}

	return &item, nil
}

// GetActiveAutoSellListings returns all active listings tied to a specific auto-sell container
func (r *ForSaleItems) GetActiveAutoSellListings(ctx context.Context, autoSellContainerID int64) ([]*models.ForSaleItem, error) {
	query := `
		SELECT
			f.id,
			f.user_id,
			f.type_id,
			f.owner_type,
			f.owner_id,
			f.location_id,
			f.container_id,
			f.division_number,
			f.quantity_available,
			f.price_per_unit,
			f.notes,
			f.auto_sell_container_id,
			f.is_active
		FROM for_sale_items f
		WHERE f.auto_sell_container_id = $1 AND f.is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, autoSellContainerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query auto-sell listings")
	}
	defer rows.Close()

	items := []*models.ForSaleItem{}
	for rows.Next() {
		var item models.ForSaleItem
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TypeID,
			&item.OwnerType,
			&item.OwnerID,
			&item.LocationID,
			&item.ContainerID,
			&item.DivisionNumber,
			&item.QuantityAvailable,
			&item.PricePerUnit,
			&item.Notes,
			&item.AutoSellContainerID,
			&item.IsActive,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-sell listing")
		}
		items = append(items, &item)
	}

	return items, nil
}

// DeactivateAutoSellListings soft-deletes all active listings for an auto-sell container
func (r *ForSaleItems) DeactivateAutoSellListings(ctx context.Context, autoSellContainerID int64) error {
	query := `
		UPDATE for_sale_items
		SET is_active = false, updated_at = NOW()
		WHERE auto_sell_container_id = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, autoSellContainerID)
	if err != nil {
		return errors.Wrap(err, "failed to deactivate auto-sell listings")
	}

	return nil
}
