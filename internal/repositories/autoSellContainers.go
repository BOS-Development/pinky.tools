package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type AutoSellContainers struct {
	db *sql.DB
}

func NewAutoSellContainers(db *sql.DB) *AutoSellContainers {
	return &AutoSellContainers{db: db}
}

// GetByUser returns all active auto-sell configs for a user
func (r *AutoSellContainers) GetByUser(ctx context.Context, userID int64) ([]*models.AutoSellContainer, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, price_percentage, price_source, is_active, created_at, updated_at
		FROM auto_sell_containers
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query auto-sell containers")
	}
	defer rows.Close()

	items := []*models.AutoSellContainer{}
	for rows.Next() {
		var item models.AutoSellContainer
		err = rows.Scan(
			&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
			&item.LocationID, &item.ContainerID, &item.DivisionNumber,
			&item.PricePercentage, &item.PriceSource, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-sell container")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetAllActive returns all active auto-sell containers across all users
func (r *AutoSellContainers) GetAllActive(ctx context.Context) ([]*models.AutoSellContainer, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, price_percentage, price_source, is_active, created_at, updated_at
		FROM auto_sell_containers
		WHERE is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query all active auto-sell containers")
	}
	defer rows.Close()

	items := []*models.AutoSellContainer{}
	for rows.Next() {
		var item models.AutoSellContainer
		err = rows.Scan(
			&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
			&item.LocationID, &item.ContainerID, &item.DivisionNumber,
			&item.PricePercentage, &item.PriceSource, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-sell container")
		}
		items = append(items, &item)
	}

	return items, nil
}

// Upsert creates or updates an auto-sell container config
func (r *AutoSellContainers) Upsert(ctx context.Context, container *models.AutoSellContainer) error {
	query := `
		INSERT INTO auto_sell_containers
		(user_id, owner_type, owner_id, location_id, container_id, division_number, price_percentage, price_source, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, NOW())
		ON CONFLICT (user_id, owner_type, owner_id, location_id, coalesce(container_id, 0), coalesce(division_number, 0))
		WHERE is_active = true
		DO UPDATE SET
			price_percentage = EXCLUDED.price_percentage,
			price_source = EXCLUDED.price_source,
			updated_at = NOW()
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		container.UserID,
		container.OwnerType,
		container.OwnerID,
		container.LocationID,
		container.ContainerID,
		container.DivisionNumber,
		container.PricePercentage,
		container.PriceSource,
	).Scan(&container.ID, &container.IsActive, &container.CreatedAt, &container.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to upsert auto-sell container")
	}

	return nil
}

// Delete soft-deletes an auto-sell container config
func (r *AutoSellContainers) Delete(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE auto_sell_containers
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete auto-sell container")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("auto-sell container not found or user is not the owner")
	}

	return nil
}

// GetItemsInContainer returns items inside a container, grouped by type_id with summed quantities
func (r *AutoSellContainers) GetItemsInContainer(ctx context.Context, ownerType string, ownerID int64, containerID int64) ([]*models.ContainerItem, error) {
	var query string
	if ownerType == "character" {
		query = `
			SELECT type_id, SUM(quantity) as total_quantity
			FROM character_assets
			WHERE character_id = $1 AND location_id = $2 AND location_type = 'item'
			GROUP BY type_id
		`
	} else {
		query = `
			SELECT type_id, SUM(quantity) as total_quantity
			FROM corporation_assets
			WHERE corporation_id = $1 AND location_id = $2 AND location_type = 'item'
			GROUP BY type_id
		`
	}

	rows, err := r.db.QueryContext(ctx, query, ownerID, containerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query container items")
	}
	defer rows.Close()

	items := []*models.ContainerItem{}
	for rows.Next() {
		var item models.ContainerItem
		err = rows.Scan(&item.TypeID, &item.Quantity)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan container item")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetItemsInDivision returns items in a corp hangar division (no container), grouped by type_id with summed quantities.
// For corporation owners, joins through OfficeFolder to resolve station_id → division items.
// For character owners, queries items directly in the station hangar.
func (r *AutoSellContainers) GetItemsInDivision(ctx context.Context, ownerType string, ownerID int64, locationID int64, divisionNumber int) ([]*models.ContainerItem, error) {
	var query string
	if ownerType == "character" {
		query = `
			SELECT type_id, SUM(quantity) as total_quantity
			FROM character_assets
			WHERE character_id = $1 AND location_id = $2 AND location_type = 'other' AND location_flag = 'Hangar'
			GROUP BY type_id
		`
	} else {
		query = fmt.Sprintf(`
			SELECT ca.type_id, SUM(ca.quantity) as total_quantity
			FROM corporation_assets ca
			INNER JOIN corporation_assets office ON (
				office.item_id = ca.location_id
				AND office.corporation_id = ca.corporation_id
				AND office.user_id = ca.user_id
				AND office.location_flag = 'OfficeFolder'
				AND office.location_id = $2
			)
			WHERE ca.corporation_id = $1
				AND ca.location_type = 'item'
				AND ca.location_flag = 'CorpSAG%d'
			GROUP BY ca.type_id
		`, divisionNumber)
	}

	rows, err := r.db.QueryContext(ctx, query, ownerID, locationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query division items")
	}
	defer rows.Close()

	items := []*models.ContainerItem{}
	for rows.Next() {
		var item models.ContainerItem
		err = rows.Scan(&item.TypeID, &item.Quantity)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan division item")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetByID returns a single auto-sell container by ID
func (r *AutoSellContainers) GetByID(ctx context.Context, id int64) (*models.AutoSellContainer, error) {
	query := `
		SELECT id, user_id, owner_type, owner_id, location_id, container_id,
			division_number, price_percentage, price_source, is_active, created_at, updated_at
		FROM auto_sell_containers
		WHERE id = $1
	`

	var item models.AutoSellContainer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.UserID, &item.OwnerType, &item.OwnerID,
		&item.LocationID, &item.ContainerID, &item.DivisionNumber,
		&item.PricePercentage, &item.PriceSource, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get auto-sell container")
	}

	return &item, nil
}
