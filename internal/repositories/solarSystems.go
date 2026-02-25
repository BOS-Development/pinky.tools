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

func (r *SolarSystems) GetByID(ctx context.Context, id int64) (*models.SolarSystem, error) {
	query := `
		select solar_system_id, name, constellation_id, security, x, y, z
		from solar_systems
		where solar_system_id = $1
	`

	var system models.SolarSystem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&system.ID,
		&system.Name,
		&system.ConstellationID,
		&system.Security,
		&system.X,
		&system.Y,
		&system.Z,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get solar system by ID")
	}

	return &system, nil
}

func (r *SolarSystems) GetByIDs(ctx context.Context, ids []int64) ([]*models.SolarSystem, error) {
	if len(ids) == 0 {
		return []*models.SolarSystem{}, nil
	}

	query := `
		select solar_system_id, name, constellation_id, security, x, y, z
		from solar_systems
		where solar_system_id = any($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query solar systems by IDs")
	}
	defer rows.Close()

	systems := []*models.SolarSystem{}
	for rows.Next() {
		var system models.SolarSystem
		if err := rows.Scan(
			&system.ID,
			&system.Name,
			&system.ConstellationID,
			&system.Security,
			&system.X,
			&system.Y,
			&system.Z,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan solar system")
		}
		systems = append(systems, &system)
	}

	return systems, nil
}

func (r *SolarSystems) Search(ctx context.Context, query string, limit int) ([]*models.SolarSystem, error) {
	if limit <= 0 {
		limit = 20
	}

	searchQuery := `
		select solar_system_id, name, constellation_id, security, x, y, z
		from solar_systems
		where lower(name) like lower($1)
		order by
			case
				when lower(name) = lower($3) then 1
				when lower(name) like lower($3) || '%' then 2
				else 3
			end,
			name
		limit $2
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search solar systems")
	}
	defer rows.Close()

	systems := []*models.SolarSystem{}
	for rows.Next() {
		var system models.SolarSystem
		if err := rows.Scan(
			&system.ID,
			&system.Name,
			&system.ConstellationID,
			&system.Security,
			&system.X,
			&system.Y,
			&system.Z,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan solar system")
		}
		systems = append(systems, &system)
	}

	return systems, nil
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
		security,
		x,
		y,
		z
	)
	values
		($1,$2,$3,$4,$5,$6,$7)
on conflict
	(solar_system_id)
do update set
	name = EXCLUDED.name,
	constellation_id = EXCLUDED.constellation_id,
	security = EXCLUDED.security,
	x = EXCLUDED.x,
	y = EXCLUDED.y,
	z = EXCLUDED.z;
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
		_, err := smt.ExecContext(ctx, system.ID, system.Name, system.ConstellationID, system.Security, system.X, system.Y, system.Z)
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
