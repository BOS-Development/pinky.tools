package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type BuyOrders struct {
	db *sql.DB
}

func NewBuyOrders(db *sql.DB) *BuyOrders {
	return &BuyOrders{db: db}
}

// Create creates a new buy order
func (r *BuyOrders) Create(ctx context.Context, order *models.BuyOrder) error {
	query := `
		INSERT INTO buy_orders (
			buyer_user_id,
			type_id,
			location_id,
			quantity_desired,
			max_price_per_unit,
			notes,
			is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		order.BuyerUserID,
		order.TypeID,
		order.LocationID,
		order.QuantityDesired,
		order.MaxPricePerUnit,
		order.Notes,
		order.IsActive,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create buy order")
	}

	return nil
}

// UpsertAutoBuy creates or updates an auto-buy order using the unique index on
// (buyer_user_id, type_id, location_id, auto_buy_config_id)
func (r *BuyOrders) UpsertAutoBuy(ctx context.Context, order *models.BuyOrder) error {
	query := `
		INSERT INTO buy_orders (
			buyer_user_id, type_id, location_id,
			quantity_desired, max_price_per_unit, notes,
			auto_buy_config_id, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		ON CONFLICT (buyer_user_id, type_id, location_id, auto_buy_config_id)
			WHERE auto_buy_config_id IS NOT NULL AND is_active = true
		DO UPDATE SET
			quantity_desired = EXCLUDED.quantity_desired,
			max_price_per_unit = EXCLUDED.max_price_per_unit,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		order.BuyerUserID,
		order.TypeID,
		order.LocationID,
		order.QuantityDesired,
		order.MaxPricePerUnit,
		order.Notes,
		order.AutoBuyConfigID,
	).Scan(&order.ID, &order.IsActive, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to upsert auto-buy order")
	}

	return nil
}

// GetActiveAutoBuyOrders returns active buy orders linked to a specific auto-buy config
func (r *BuyOrders) GetActiveAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) ([]*models.BuyOrder, error) {
	query := `
		SELECT
			id, buyer_user_id, type_id, location_id,
			quantity_desired, max_price_per_unit, notes,
			auto_buy_config_id, is_active, created_at, updated_at
		FROM buy_orders
		WHERE auto_buy_config_id = $1 AND is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, autoBuyConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query active auto-buy orders")
	}
	defer rows.Close()

	orders := []*models.BuyOrder{}
	for rows.Next() {
		order := &models.BuyOrder{}
		err := rows.Scan(
			&order.ID, &order.BuyerUserID, &order.TypeID, &order.LocationID,
			&order.QuantityDesired, &order.MaxPricePerUnit, &order.Notes,
			&order.AutoBuyConfigID, &order.IsActive, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-buy order")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// DeactivateAutoBuyOrders deactivates all active buy orders for a given auto-buy config
func (r *BuyOrders) DeactivateAutoBuyOrders(ctx context.Context, autoBuyConfigID int64) error {
	query := `
		UPDATE buy_orders
		SET is_active = false, updated_at = NOW()
		WHERE auto_buy_config_id = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, autoBuyConfigID)
	if err != nil {
		return errors.Wrap(err, "failed to deactivate auto-buy orders")
	}

	return nil
}

// DeactivateAutoBuyOrder deactivates a single auto-buy order by ID
func (r *BuyOrders) DeactivateAutoBuyOrder(ctx context.Context, orderID int64) error {
	query := `
		UPDATE buy_orders
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return errors.Wrap(err, "failed to deactivate auto-buy order")
	}

	return nil
}

// GetByID retrieves a buy order by ID with type name populated
func (r *BuyOrders) GetByID(ctx context.Context, id int64) (*models.BuyOrder, error) {
	query := `
		SELECT
			bo.id,
			bo.buyer_user_id,
			bo.type_id,
			it.type_name,
			bo.location_id,
			COALESCE(st.name, ss.name, '') AS location_name,
			bo.quantity_desired,
			bo.max_price_per_unit,
			bo.notes,
			bo.auto_buy_config_id,
			bo.is_active,
			bo.created_at,
			bo.updated_at
		FROM buy_orders bo
		LEFT JOIN asset_item_types it ON bo.type_id = it.type_id
		LEFT JOIN stations st ON bo.location_id = st.station_id
		LEFT JOIN solar_systems ss ON bo.location_id = ss.solar_system_id
		WHERE bo.id = $1
	`

	order := &models.BuyOrder{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.BuyerUserID,
		&order.TypeID,
		&order.TypeName,
		&order.LocationID,
		&order.LocationName,
		&order.QuantityDesired,
		&order.MaxPricePerUnit,
		&order.Notes,
		&order.AutoBuyConfigID,
		&order.IsActive,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("buy order not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get buy order")
	}

	return order, nil
}

// GetByUser returns all buy orders for a user, ordered by created_at DESC
func (r *BuyOrders) GetByUser(ctx context.Context, userID int64) ([]*models.BuyOrder, error) {
	query := `
		SELECT
			bo.id,
			bo.buyer_user_id,
			bo.type_id,
			it.type_name,
			bo.location_id,
			COALESCE(st.name, ss.name, '') AS location_name,
			bo.quantity_desired,
			bo.max_price_per_unit,
			bo.notes,
			bo.auto_buy_config_id,
			bo.is_active,
			bo.created_at,
			bo.updated_at
		FROM buy_orders bo
		LEFT JOIN asset_item_types it ON bo.type_id = it.type_id
		LEFT JOIN stations st ON bo.location_id = st.station_id
		LEFT JOIN solar_systems ss ON bo.location_id = ss.solar_system_id
		WHERE bo.buyer_user_id = $1
		ORDER BY bo.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query buy orders")
	}
	defer rows.Close()

	orders := []*models.BuyOrder{}
	for rows.Next() {
		order := &models.BuyOrder{}
		err := rows.Scan(
			&order.ID,
			&order.BuyerUserID,
			&order.TypeID,
			&order.TypeName,
			&order.LocationID,
			&order.LocationName,
			&order.QuantityDesired,
			&order.MaxPricePerUnit,
			&order.Notes,
			&order.AutoBuyConfigID,
			&order.IsActive,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan buy order")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetDemandForSeller returns active buy orders from users who have granted seller the for_sale_browse permission
func (r *BuyOrders) GetDemandForSeller(ctx context.Context, sellerUserID int64) ([]*models.BuyOrder, error) {
	query := `
		SELECT DISTINCT
			bo.id,
			bo.buyer_user_id,
			bo.type_id,
			it.type_name,
			bo.location_id,
			COALESCE(st.name, ss.name, '') AS location_name,
			bo.quantity_desired,
			bo.max_price_per_unit,
			bo.notes,
			bo.auto_buy_config_id,
			bo.is_active,
			bo.created_at,
			bo.updated_at
		FROM buy_orders bo
		LEFT JOIN asset_item_types it ON bo.type_id = it.type_id
		LEFT JOIN stations st ON bo.location_id = st.station_id
		LEFT JOIN solar_systems ss ON bo.location_id = ss.solar_system_id
		INNER JOIN contact_permissions cp ON cp.granting_user_id = bo.buyer_user_id
			AND cp.receiving_user_id = $1
			AND cp.service_type = 'for_sale_browse'
			AND cp.can_access = true
		WHERE bo.is_active = true
		ORDER BY bo.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sellerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query demand")
	}
	defer rows.Close()

	orders := []*models.BuyOrder{}
	for rows.Next() {
		order := &models.BuyOrder{}
		err := rows.Scan(
			&order.ID,
			&order.BuyerUserID,
			&order.TypeID,
			&order.TypeName,
			&order.LocationID,
			&order.LocationName,
			&order.QuantityDesired,
			&order.MaxPricePerUnit,
			&order.Notes,
			&order.AutoBuyConfigID,
			&order.IsActive,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan buy order")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Update updates a buy order
func (r *BuyOrders) Update(ctx context.Context, order *models.BuyOrder) error {
	query := `
		UPDATE buy_orders
		SET
			quantity_desired = $2,
			max_price_per_unit = $3,
			notes = $4,
			is_active = $5,
			location_id = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		order.ID,
		order.QuantityDesired,
		order.MaxPricePerUnit,
		order.Notes,
		order.IsActive,
		order.LocationID,
	).Scan(&order.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.New("buy order not found")
	}
	if err != nil {
		return errors.Wrap(err, "failed to update buy order")
	}

	return nil
}

// Delete soft-deletes a buy order by setting is_active = false
func (r *BuyOrders) Delete(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE buy_orders
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND buyer_user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete buy order")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("buy order not found or not owned by user")
	}

	return nil
}
