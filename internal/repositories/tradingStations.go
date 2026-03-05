package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type TradingStations struct {
	db *sql.DB
}

func NewTradingStations(db *sql.DB) *TradingStations {
	return &TradingStations{db: db}
}

// ListStations returns all trading stations ordered by preset first, then name.
func (r *TradingStations) ListStations(ctx context.Context) ([]*models.TradingStation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, station_id, name, system_id, region_id, is_preset
		FROM trading_stations
		ORDER BY is_preset DESC, name ASC`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list trading stations")
	}
	defer rows.Close()

	stations := []*models.TradingStation{}
	for rows.Next() {
		var s models.TradingStation
		if err := rows.Scan(&s.ID, &s.StationID, &s.Name, &s.SystemID, &s.RegionID, &s.IsPreset); err != nil {
			return nil, errors.Wrap(err, "failed to scan trading station")
		}
		stations = append(stations, &s)
	}
	return stations, nil
}
