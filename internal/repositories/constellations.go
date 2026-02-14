package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type Constellations struct {
	db *sql.DB
}

func NewConstellations(db *sql.DB) *Constellations {
	return &Constellations{
		db: db,
	}
}

func (r *Constellations) Upsert(ctx context.Context, constellations []models.Constellation) error {
	if len(constellations) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	constellations
	(
		constellation_id,
		name,
		region_id
	)
	values
		($1,$2,$3)
on conflict
	(constellation_id)
do update set
	name = EXCLUDED.name,
	region_id = EXCLUDED.region_id;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for constellation upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for constellation upsert")
	}

	for _, constellation := range constellations {
		_, err := smt.ExecContext(ctx, constellation.ID, constellation.Name, constellation.RegionID)
		if err != nil {
			return errors.Wrap(err, "failed to execute constellation upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit constellation transaction")
	}

	return nil
}
