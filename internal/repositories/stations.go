package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
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
		is_npc_station,
		last_updated_at
	)
	values
		($1,$2,$3,$4,$5,now())
on conflict
	(station_id)
do update set
	name = CASE
		WHEN EXCLUDED.name IN ('', 'Unknown Structure')
			AND stations.name NOT IN ('', 'Unknown Structure')
			THEN stations.name
		ELSE EXCLUDED.name
	END,
	solar_system_id = CASE
		WHEN EXCLUDED.solar_system_id = 0 THEN stations.solar_system_id
		ELSE EXCLUDED.solar_system_id
	END,
	corporation_id = CASE
		WHEN EXCLUDED.corporation_id = 0 THEN stations.corporation_id
		ELSE EXCLUDED.corporation_id
	END,
	is_npc_station = EXCLUDED.is_npc_station,
	last_updated_at = now();
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

// FilterStaleStationIDs returns only station IDs that need refreshing from ESI.
// Known stations (with a real name) use knownMaxAge; unknown stations use unknownMaxAge.
// Stations not in the table at all are always returned.
func (r *Stations) FilterStaleStationIDs(ctx context.Context, ids []int64, knownMaxAge, unknownMaxAge time.Duration) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}

	knownCutoff := time.Now().Add(-knownMaxAge)
	unknownCutoff := time.Now().Add(-unknownMaxAge)

	query := `
		SELECT unnested_id FROM unnest($1::bigint[]) AS unnested_id
		WHERE unnested_id NOT IN (
			SELECT station_id FROM stations
			WHERE station_id = ANY($1)
			  AND last_updated_at IS NOT NULL
			  AND last_updated_at > CASE
				WHEN name NOT IN ('', 'Unknown Structure') THEN $2::timestamp
				ELSE $3::timestamp
			  END
		)
	`

	rows, err := r.db.QueryContext(ctx, query,
		pq.Array(ids),
		knownCutoff,
		unknownCutoff,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter stale station IDs")
	}
	defer rows.Close()

	staleIDs := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "failed to scan stale station ID")
		}
		staleIDs = append(staleIDs, id)
	}
	return staleIDs, nil
}
