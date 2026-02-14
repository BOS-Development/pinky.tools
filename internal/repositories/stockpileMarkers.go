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
		       container_id, division_number, desired_quantity, notes
		FROM stockpile_markers
		WHERE user_id = $1
		ORDER BY type_id, location_id
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stockpile markers")
	}
	defer rows.Close()

	var markers []*models.StockpileMarker
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
		(user_id, type_id, owner_type, owner_id, location_id, container_id, division_number, desired_quantity, notes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (user_id, type_id, owner_type, owner_id, location_id, COALESCE(container_id, 0::BIGINT), COALESCE(division_number, 0))
		DO UPDATE SET
			desired_quantity = EXCLUDED.desired_quantity,
			notes = EXCLUDED.notes,
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
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert stockpile marker")
	}

	return nil
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
