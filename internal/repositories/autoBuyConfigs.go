package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type AutoBuyConfigs struct {
	db *sql.DB
}

func NewAutoBuyConfigs(db *sql.DB) *AutoBuyConfigs {
	return &AutoBuyConfigs{db: db}
}

// GetByUser returns all active auto-buy configs for a user
func (r *AutoBuyConfigs) GetByUser(ctx context.Context, userID int64) ([]*models.AutoBuyConfig, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, min_price_percentage, max_price_percentage, price_source,
			is_active, created_at, updated_at
		FROM auto_buy_configs
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query auto-buy configs")
	}
	defer rows.Close()

	items := []*models.AutoBuyConfig{}
	for rows.Next() {
		var item models.AutoBuyConfig
		err = rows.Scan(
			&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
			&item.LocationID, &item.ContainerID, &item.DivisionNumber,
			&item.MinPricePercentage, &item.MaxPricePercentage, &item.PriceSource,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-buy config")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetAllActive returns all active auto-buy configs across all users
func (r *AutoBuyConfigs) GetAllActive(ctx context.Context) ([]*models.AutoBuyConfig, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, min_price_percentage, max_price_percentage, price_source,
			is_active, created_at, updated_at
		FROM auto_buy_configs
		WHERE is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query all active auto-buy configs")
	}
	defer rows.Close()

	items := []*models.AutoBuyConfig{}
	for rows.Next() {
		var item models.AutoBuyConfig
		err = rows.Scan(
			&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
			&item.LocationID, &item.ContainerID, &item.DivisionNumber,
			&item.MinPricePercentage, &item.MaxPricePercentage, &item.PriceSource,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-buy config")
		}
		items = append(items, &item)
	}

	return items, nil
}

// Upsert creates or updates an auto-buy config
func (r *AutoBuyConfigs) Upsert(ctx context.Context, config *models.AutoBuyConfig) error {
	query := `
		INSERT INTO auto_buy_configs
		(user_id, owner_type, owner_id, location_id, container_id, division_number,
		 min_price_percentage, max_price_percentage, price_source, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW())
		ON CONFLICT (user_id, owner_type, owner_id, location_id, coalesce(container_id, 0), coalesce(division_number, 0))
		WHERE is_active = true
		DO UPDATE SET
			min_price_percentage = EXCLUDED.min_price_percentage,
			max_price_percentage = EXCLUDED.max_price_percentage,
			price_source = EXCLUDED.price_source,
			updated_at = NOW()
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		config.UserID,
		config.OwnerType,
		config.OwnerID,
		config.LocationID,
		config.ContainerID,
		config.DivisionNumber,
		config.MinPricePercentage,
		config.MaxPricePercentage,
		config.PriceSource,
	).Scan(&config.ID, &config.IsActive, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to upsert auto-buy config")
	}

	return nil
}

// Delete soft-deletes an auto-buy config
func (r *AutoBuyConfigs) Delete(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE auto_buy_configs
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete auto-buy config")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("auto-buy config not found or user is not the owner")
	}

	return nil
}

// GetByID returns a single auto-buy config by ID
func (r *AutoBuyConfigs) GetByID(ctx context.Context, id int64) (*models.AutoBuyConfig, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, min_price_percentage, max_price_percentage, price_source,
			is_active, created_at, updated_at
		FROM auto_buy_configs
		WHERE id = $1
	`

	var item models.AutoBuyConfig
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
		&item.LocationID, &item.ContainerID, &item.DivisionNumber,
		&item.MinPricePercentage, &item.MaxPricePercentage, &item.PriceSource,
		&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auto-buy config")
	}

	return &item, nil
}

// GetStockpileDeficitsForConfig returns stockpile deficits for items matching the config's container context.
// It joins stockpile_markers with asset tables to compute current quantities and deficits.
func (r *AutoBuyConfigs) GetStockpileDeficitsForConfig(ctx context.Context, config *models.AutoBuyConfig) ([]*models.StockpileDeficitItem, error) {
	var query string

	if config.ContainerID != nil {
		// Items in a specific container
		if config.OwnerType == "character" {
			query = `
				SELECT
					sm.type_id,
					sm.desired_quantity,
					COALESCE(SUM(ca.quantity), 0) AS current_quantity,
					sm.desired_quantity - COALESCE(SUM(ca.quantity), 0) AS deficit,
					sm.price_source,
					sm.price_percentage
				FROM stockpile_markers sm
				LEFT JOIN character_assets ca
					ON ca.type_id = sm.type_id
					AND ca.character_id = sm.owner_id
					AND ca.location_id = $5
					AND ca.location_type = 'item'
				WHERE sm.user_id = $1
					AND sm.owner_type = $2
					AND sm.owner_id = $3
					AND sm.location_id = $4
					AND COALESCE(sm.container_id, 0) = COALESCE($5, 0)
					AND COALESCE(sm.division_number, 0) = COALESCE($6, 0)
				GROUP BY sm.type_id, sm.desired_quantity, sm.price_source, sm.price_percentage
			`
		} else {
			query = `
				SELECT
					sm.type_id,
					sm.desired_quantity,
					COALESCE(SUM(ca.quantity), 0) AS current_quantity,
					sm.desired_quantity - COALESCE(SUM(ca.quantity), 0) AS deficit,
					sm.price_source,
					sm.price_percentage
				FROM stockpile_markers sm
				LEFT JOIN corporation_assets ca
					ON ca.type_id = sm.type_id
					AND ca.corporation_id = sm.owner_id
					AND ca.location_id = $5
					AND ca.location_type = 'item'
				WHERE sm.user_id = $1
					AND sm.owner_type = $2
					AND sm.owner_id = $3
					AND sm.location_id = $4
					AND COALESCE(sm.container_id, 0) = COALESCE($5, 0)
					AND COALESCE(sm.division_number, 0) = COALESCE($6, 0)
				GROUP BY sm.type_id, sm.desired_quantity, sm.price_source, sm.price_percentage
			`
		}
	} else {
		// Items in hangar (no container)
		if config.OwnerType == "character" {
			query = `
				SELECT
					sm.type_id,
					sm.desired_quantity,
					COALESCE(SUM(ca.quantity), 0) AS current_quantity,
					sm.desired_quantity - COALESCE(SUM(ca.quantity), 0) AS deficit,
					sm.price_source,
					sm.price_percentage
				FROM stockpile_markers sm
				LEFT JOIN character_assets ca
					ON ca.type_id = sm.type_id
					AND ca.character_id = sm.owner_id
					AND ca.location_id = $4
					AND ca.location_type = 'other'
					AND ca.location_flag = 'Hangar'
				WHERE sm.user_id = $1
					AND sm.owner_type = $2
					AND sm.owner_id = $3
					AND sm.location_id = $4
					AND sm.container_id IS NULL
					AND COALESCE(sm.division_number, 0) = COALESCE($5, 0)
				GROUP BY sm.type_id, sm.desired_quantity, sm.price_source, sm.price_percentage
			`
		} else {
			query = `
				SELECT
					sm.type_id,
					sm.desired_quantity,
					COALESCE(SUM(items.quantity), 0) AS current_quantity,
					sm.desired_quantity - COALESCE(SUM(items.quantity), 0) AS deficit,
					sm.price_source,
					sm.price_percentage
				FROM stockpile_markers sm
				LEFT JOIN (
					SELECT ca.type_id, ca.corporation_id, SUM(ca.quantity) as quantity
					FROM corporation_assets ca
					INNER JOIN corporation_assets office ON (
						office.item_id = ca.location_id
						AND office.corporation_id = ca.corporation_id
						AND office.user_id = ca.user_id
						AND office.location_flag = 'OfficeFolder'
						AND office.location_id = $4
					)
					WHERE ca.corporation_id = $3
						AND ca.location_type = 'item'
						AND ca.location_flag = 'CorpSAG' || $5::text
					GROUP BY ca.type_id, ca.corporation_id
				) items ON items.type_id = sm.type_id
				WHERE sm.user_id = $1
					AND sm.owner_type = $2
					AND sm.owner_id = $3
					AND sm.location_id = $4
					AND sm.container_id IS NULL
					AND COALESCE(sm.division_number, 0) = COALESCE($5, 0)
				GROUP BY sm.type_id, sm.desired_quantity, sm.price_source, sm.price_percentage
			`
		}
	}

	var rows *sql.Rows
	var err error

	if config.ContainerID != nil {
		rows, err = r.db.QueryContext(ctx, query,
			config.UserID, config.OwnerType, config.OwnerID,
			config.LocationID, config.ContainerID, config.DivisionNumber,
		)
	} else {
		rows, err = r.db.QueryContext(ctx, query,
			config.UserID, config.OwnerType, config.OwnerID,
			config.LocationID, config.DivisionNumber,
		)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to query stockpile deficits for config")
	}
	defer rows.Close()

	items := []*models.StockpileDeficitItem{}
	for rows.Next() {
		var item models.StockpileDeficitItem
		err = rows.Scan(
			&item.TypeID,
			&item.DesiredQuantity,
			&item.CurrentQuantity,
			&item.Deficit,
			&item.PriceSource,
			&item.PricePercentage,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan stockpile deficit item")
		}
		items = append(items, &item)
	}

	return items, nil
}
