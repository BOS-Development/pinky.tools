package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type StockpileMarkers struct {
	db *sql.DB
}

func NewStockpileMarkers(db *sql.DB) *StockpileMarkers {
	return &StockpileMarkers{db: db}
}

func (r *StockpileMarkers) GetByUser(ctx context.Context, userID int64) ([]*models.StockpileMarker, error) {
	query := `
		SELECT user_id, type_id, owner_type, owner_id, location_id,
		       container_id, division_number, desired_quantity, notes,
		       price_source, price_percentage,
		       plan_id, auto_production_parallelism, auto_production_enabled
		FROM stockpile_markers
		WHERE user_id = $1
		ORDER BY type_id, location_id
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stockpile markers")
	}
	defer rows.Close()

	markers := []*models.StockpileMarker{}
	for rows.Next() {
		var marker models.StockpileMarker
		err = rows.Scan(
			&marker.UserID,
			&marker.TypeID,
			&marker.OwnerType,
			&marker.OwnerID,
			&marker.LocationID,
			&marker.ContainerID,
			&marker.DivisionNumber,
			&marker.DesiredQuantity,
			&marker.Notes,
			&marker.PriceSource,
			&marker.PricePercentage,
			&marker.PlanID,
			&marker.AutoProductionParallelism,
			&marker.AutoProductionEnabled,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan stockpile marker")
		}
		markers = append(markers, &marker)
	}

	return markers, nil
}

func (r *StockpileMarkers) Upsert(ctx context.Context, marker *models.StockpileMarker) error {
	query := `
		INSERT INTO stockpile_markers
		(user_id, type_id, owner_type, owner_id, location_id, container_id, division_number, desired_quantity, notes, price_source, price_percentage, plan_id, auto_production_parallelism, auto_production_enabled, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (user_id, type_id, owner_type, owner_id, location_id, COALESCE(container_id, 0::BIGINT), COALESCE(division_number, 0))
		DO UPDATE SET
			desired_quantity = EXCLUDED.desired_quantity,
			notes = EXCLUDED.notes,
			price_source = EXCLUDED.price_source,
			price_percentage = EXCLUDED.price_percentage,
			plan_id = EXCLUDED.plan_id,
			auto_production_parallelism = EXCLUDED.auto_production_parallelism,
			auto_production_enabled = EXCLUDED.auto_production_enabled,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		marker.UserID,
		marker.TypeID,
		marker.OwnerType,
		marker.OwnerID,
		marker.LocationID,
		marker.ContainerID,
		marker.DivisionNumber,
		marker.DesiredQuantity,
		marker.Notes,
		marker.PriceSource,
		marker.PricePercentage,
		marker.PlanID,
		marker.AutoProductionParallelism,
		marker.AutoProductionEnabled,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert stockpile marker")
	}

	return nil
}

// GetByContainerContext returns stockpile markers matching a specific container scope, keyed by TypeID.
func (r *StockpileMarkers) GetByContainerContext(
	ctx context.Context,
	userID int64,
	ownerType string,
	ownerID int64,
	locationID int64,
	containerID int64,
	divisionNumber *int,
) (map[int64]*models.StockpileMarker, error) {
	query := `
		SELECT user_id, type_id, owner_type, owner_id, location_id,
		       container_id, division_number, desired_quantity, notes,
		       price_source, price_percentage,
		       plan_id, auto_production_parallelism, auto_production_enabled
		FROM stockpile_markers
		WHERE user_id = $1
		  AND owner_type = $2
		  AND owner_id = $3
		  AND location_id = $4
		  AND COALESCE(container_id, 0::BIGINT) = COALESCE($5, 0::BIGINT)
		  AND COALESCE(division_number, 0) = COALESCE($6, 0)
	`

	rows, err := r.db.QueryContext(ctx, query,
		userID, ownerType, ownerID, locationID, containerID, divisionNumber,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stockpile markers for container context")
	}
	defer rows.Close()

	markers := make(map[int64]*models.StockpileMarker)
	for rows.Next() {
		var marker models.StockpileMarker
		err = rows.Scan(
			&marker.UserID,
			&marker.TypeID,
			&marker.OwnerType,
			&marker.OwnerID,
			&marker.LocationID,
			&marker.ContainerID,
			&marker.DivisionNumber,
			&marker.DesiredQuantity,
			&marker.Notes,
			&marker.PriceSource,
			&marker.PricePercentage,
			&marker.PlanID,
			&marker.AutoProductionParallelism,
			&marker.AutoProductionEnabled,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan stockpile marker")
		}
		markers[marker.TypeID] = &marker
	}

	return markers, nil
}

// GetAutoProductionMarkers returns all markers with auto-production enabled and a plan assigned.
func (r *StockpileMarkers) GetAutoProductionMarkers(ctx context.Context) ([]*models.StockpileMarker, error) {
	query := `
		SELECT user_id, type_id, owner_type, owner_id, location_id,
		       container_id, division_number, desired_quantity, notes,
		       price_source, price_percentage,
		       plan_id, auto_production_parallelism, auto_production_enabled
		FROM stockpile_markers
		WHERE auto_production_enabled = TRUE
		  AND plan_id IS NOT NULL
		ORDER BY user_id, plan_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query auto-production markers")
	}
	defer rows.Close()

	markers := []*models.StockpileMarker{}
	for rows.Next() {
		var marker models.StockpileMarker
		err = rows.Scan(
			&marker.UserID,
			&marker.TypeID,
			&marker.OwnerType,
			&marker.OwnerID,
			&marker.LocationID,
			&marker.ContainerID,
			&marker.DivisionNumber,
			&marker.DesiredQuantity,
			&marker.Notes,
			&marker.PriceSource,
			&marker.PricePercentage,
			&marker.PlanID,
			&marker.AutoProductionParallelism,
			&marker.AutoProductionEnabled,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan auto-production marker")
		}
		markers = append(markers, &marker)
	}

	return markers, nil
}

func (r *StockpileMarkers) Delete(ctx context.Context, marker *models.StockpileMarker) error {
	query := `
		DELETE FROM stockpile_markers
		WHERE user_id = $1
		  AND type_id = $2
		  AND owner_type = $3
		  AND owner_id = $4
		  AND location_id = $5
		  AND COALESCE(container_id, 0::BIGINT) = COALESCE($6, 0::BIGINT)
		  AND COALESCE(division_number, 0) = COALESCE($7, 0)
	`

	_, err := r.db.ExecContext(ctx, query,
		marker.UserID,
		marker.TypeID,
		marker.OwnerType,
		marker.OwnerID,
		marker.LocationID,
		marker.ContainerID,
		marker.DivisionNumber,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete stockpile marker")
	}

	return nil
}
