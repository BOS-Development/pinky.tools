package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type UserStations struct {
	db *sql.DB
}

func NewUserStations(db *sql.DB) *UserStations {
	return &UserStations{db: db}
}

func (r *UserStations) GetByUser(ctx context.Context, userID int64) ([]*models.UserStation, error) {
	stationQuery := `
		SELECT us.id, us.user_id, us.station_id, us.structure, us.facility_tax,
		       us.created_at, us.updated_at,
		       COALESCE(st.name, '') as station_name,
		       ss.solar_system_id,
		       COALESCE(ss.name, '') as solar_system_name,
		       COALESCE(ss.security, 0) as security_status,
		       CASE
		           WHEN ss.security >= 0.45 THEN 'high'
		           WHEN ss.security > 0.0 THEN 'low'
		           ELSE 'null'
		       END AS security
		FROM user_stations us
		JOIN stations st ON st.station_id = us.station_id
		JOIN solar_systems ss ON ss.solar_system_id = st.solar_system_id
		WHERE us.user_id = $1
		ORDER BY us.id
	`

	rows, err := r.db.QueryContext(ctx, stationQuery, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user stations")
	}
	defer rows.Close()

	stations := []*models.UserStation{}
	stationIDs := []int64{}
	stationMap := map[int64]*models.UserStation{}

	for rows.Next() {
		var s models.UserStation
		err := rows.Scan(
			&s.ID, &s.UserID, &s.StationID, &s.Structure, &s.FacilityTax,
			&s.CreatedAt, &s.UpdatedAt,
			&s.StationName, &s.SolarSystemID, &s.SolarSystemName, &s.SecurityStatus, &s.Security,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user station")
		}
		s.Rigs = []*models.UserStationRig{}
		s.Services = []*models.UserStationService{}
		s.Activities = []string{}
		stations = append(stations, &s)
		stationIDs = append(stationIDs, s.ID)
		stationMap[s.ID] = &s
	}

	if len(stationIDs) == 0 {
		return stations, nil
	}

	if err := r.loadRigs(ctx, stationMap); err != nil {
		return nil, err
	}

	if err := r.loadServices(ctx, stationMap); err != nil {
		return nil, err
	}

	return stations, nil
}

func (r *UserStations) GetByID(ctx context.Context, id, userID int64) (*models.UserStation, error) {
	stationQuery := `
		SELECT us.id, us.user_id, us.station_id, us.structure, us.facility_tax,
		       us.created_at, us.updated_at,
		       COALESCE(st.name, '') as station_name,
		       ss.solar_system_id,
		       COALESCE(ss.name, '') as solar_system_name,
		       COALESCE(ss.security, 0) as security_status,
		       CASE
		           WHEN ss.security >= 0.45 THEN 'high'
		           WHEN ss.security > 0.0 THEN 'low'
		           ELSE 'null'
		       END AS security
		FROM user_stations us
		JOIN stations st ON st.station_id = us.station_id
		JOIN solar_systems ss ON ss.solar_system_id = st.solar_system_id
		WHERE us.id = $1 AND us.user_id = $2
	`

	var s models.UserStation
	err := r.db.QueryRowContext(ctx, stationQuery, id, userID).Scan(
		&s.ID, &s.UserID, &s.StationID, &s.Structure, &s.FacilityTax,
		&s.CreatedAt, &s.UpdatedAt,
		&s.StationName, &s.SolarSystemID, &s.SolarSystemName, &s.SecurityStatus, &s.Security,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user station")
	}

	s.Rigs = []*models.UserStationRig{}
	s.Services = []*models.UserStationService{}
	s.Activities = []string{}

	stationMap := map[int64]*models.UserStation{s.ID: &s}

	if err := r.loadRigs(ctx, stationMap); err != nil {
		return nil, err
	}

	if err := r.loadServices(ctx, stationMap); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *UserStations) Create(ctx context.Context, station *models.UserStation) (*models.UserStation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		INSERT INTO user_stations (user_id, station_id, structure, facility_tax)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		station.UserID, station.StationID, station.Structure, station.FacilityTax,
	).Scan(&station.ID, &station.CreatedAt, &station.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert user station")
	}

	if err := r.insertRigs(ctx, tx, station.ID, station.Rigs); err != nil {
		return nil, err
	}

	if err := r.insertServices(ctx, tx, station.ID, station.Services); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return station, nil
}

func (r *UserStations) Update(ctx context.Context, station *models.UserStation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		UPDATE user_stations
		SET structure = $3, facility_tax = $4, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := tx.ExecContext(ctx, query,
		station.ID, station.UserID, station.Structure, station.FacilityTax,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update user station")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("user station not found")
	}

	// Delete existing rigs and services, then re-insert
	_, err = tx.ExecContext(ctx, `DELETE FROM user_station_rigs WHERE user_station_id = $1`, station.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing rigs")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM user_station_services WHERE user_station_id = $1`, station.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing services")
	}

	if err := r.insertRigs(ctx, tx, station.ID, station.Rigs); err != nil {
		return err
	}

	if err := r.insertServices(ctx, tx, station.ID, station.Services); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserStations) Delete(ctx context.Context, id, userID int64) error {
	query := `DELETE FROM user_stations WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete user station")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("user station not found")
	}

	return nil
}

func (r *UserStations) loadRigs(ctx context.Context, stationMap map[int64]*models.UserStation) error {
	ids := make([]int64, 0, len(stationMap))
	for id := range stationMap {
		ids = append(ids, id)
	}

	query := `
		SELECT id, user_station_id, rig_name, category, tier
		FROM user_station_rigs
		WHERE user_station_id = ANY($1)
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return errors.Wrap(err, "failed to query user station rigs")
	}
	defer rows.Close()

	for rows.Next() {
		var rig models.UserStationRig
		err := rows.Scan(&rig.ID, &rig.UserStationID, &rig.RigName, &rig.Category, &rig.Tier)
		if err != nil {
			return errors.Wrap(err, "failed to scan user station rig")
		}
		if station, ok := stationMap[rig.UserStationID]; ok {
			station.Rigs = append(station.Rigs, &rig)
		}
	}

	return nil
}

func (r *UserStations) loadServices(ctx context.Context, stationMap map[int64]*models.UserStation) error {
	ids := make([]int64, 0, len(stationMap))
	for id := range stationMap {
		ids = append(ids, id)
	}

	query := `
		SELECT id, user_station_id, service_name, activity
		FROM user_station_services
		WHERE user_station_id = ANY($1)
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return errors.Wrap(err, "failed to query user station services")
	}
	defer rows.Close()

	activitySeen := map[int64]map[string]bool{}
	for rows.Next() {
		var svc models.UserStationService
		err := rows.Scan(&svc.ID, &svc.UserStationID, &svc.ServiceName, &svc.Activity)
		if err != nil {
			return errors.Wrap(err, "failed to scan user station service")
		}
		if station, ok := stationMap[svc.UserStationID]; ok {
			station.Services = append(station.Services, &svc)
			if activitySeen[svc.UserStationID] == nil {
				activitySeen[svc.UserStationID] = map[string]bool{}
			}
			if !activitySeen[svc.UserStationID][svc.Activity] {
				station.Activities = append(station.Activities, svc.Activity)
				activitySeen[svc.UserStationID][svc.Activity] = true
			}
		}
	}

	return nil
}

func (r *UserStations) insertRigs(ctx context.Context, tx *sql.Tx, stationID int64, rigs []*models.UserStationRig) error {
	if len(rigs) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_station_rigs (user_station_id, rig_name, category, tier)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	for _, rig := range rigs {
		err := tx.QueryRowContext(ctx, query,
			stationID, rig.RigName, rig.Category, rig.Tier,
		).Scan(&rig.ID)
		if err != nil {
			return errors.Wrap(err, "failed to insert user station rig")
		}
		rig.UserStationID = stationID
	}

	return nil
}

func (r *UserStations) insertServices(ctx context.Context, tx *sql.Tx, stationID int64, services []*models.UserStationService) error {
	if len(services) == 0 {
		return nil
	}

	query := `
		INSERT INTO user_station_services (user_station_id, service_name, activity)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	for _, svc := range services {
		err := tx.QueryRowContext(ctx, query,
			stationID, svc.ServiceName, svc.Activity,
		).Scan(&svc.ID)
		if err != nil {
			return errors.Wrap(err, "failed to insert user station service")
		}
		svc.UserStationID = stationID
	}

	return nil
}
