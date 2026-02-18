package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type Stations struct {
	db *sql.DB
}

func NewStations(db *sql.DB) *Stations {
	return &Stations{
		db: db,
	}
}

func (r *Stations) Upsert(ctx context.Context, stations []models.Station) error {
	if len(stations) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	stations
	(
		station_id,
		name,
		solar_system_id,
		corporation_id,
		is_npc_station
	)
	values
		($1,$2,$3,$4,$5)
on conflict
	(station_id)
do update set
	name = CASE WHEN EXCLUDED.name = '' THEN stations.name ELSE EXCLUDED.name END,
	solar_system_id = EXCLUDED.solar_system_id,
	corporation_id = EXCLUDED.corporation_id,
	is_npc_station = EXCLUDED.is_npc_station;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for station upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for station upsert")
	}

	for _, station := range stations {
		_, err := smt.ExecContext(ctx, station.ID, station.Name, station.SolarSystemID, station.CorporationID, station.IsNPC)
		if err != nil {
			return errors.Wrap(err, "failed to execute station upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit station transaction")
	}
	return nil
}

func (r *Stations) GetStationsWithEmptyNames(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
select station_id from stations where name = '' and is_npc_station = true`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stations with empty names")
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "failed to scan station id")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Stations) UpdateNames(ctx context.Context, names map[int64]string) error {
	if len(names) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for station name update")
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `update stations set name = $1 where station_id = $2`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare station name update")
	}

	for id, name := range names {
		if _, err := stmt.ExecContext(ctx, name, id); err != nil {
			return errors.Wrap(err, "failed to update station name")
		}
	}

	return tx.Commit()
}
