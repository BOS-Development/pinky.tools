package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type SolarSystems struct {
	db *sql.DB
}

func NewSolarSystems(db *sql.DB) *SolarSystems {
	return &SolarSystems{
		db: db,
	}
}

func (r *SolarSystems) GetNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	names := map[int64]string{}
	query := `select solar_system_id, name from solar_systems where solar_system_id = any($1)`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query solar system names")
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, errors.Wrap(err, "failed to scan solar system name")
		}
		names[id] = name
	}

	return names, nil
}

func (r *SolarSystems) Upsert(ctx context.Context, systems []models.SolarSystem) error {
	if len(systems) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	solar_systems
	(
		solar_system_id,
		name,
		constellation_id,
		security
	)
	values
		($1,$2,$3,$4)
on conflict
	(solar_system_id)
do update set
	name = EXCLUDED.name,
	constellation_id = EXCLUDED.constellation_id,
	security = EXCLUDED.security;
`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for solar system upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for solar system upsert")
	}

	for _, system := range systems {
		_, err := smt.ExecContext(ctx, system.ID, system.Name, system.ConstellationID, system.Security)
		if err != nil {
			return errors.Wrap(err, "failed to execute for solar system upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit solar system")
	}
	return nil
}
