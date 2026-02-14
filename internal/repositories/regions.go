package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type Regions struct {
	db *sql.DB
}

func NewRegions(db *sql.DB) *Regions {
	return &Regions{
		db: db,
	}
}

func (r *Regions) Upsert(ctx context.Context, regions []models.Region) error {
	if len(regions) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	regions
	(
		region_id,
		name
	)
	values
		($1,$2)
on conflict
	(region_id)
do update set
	name = EXCLUDED.name;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for region upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for region upsert")
	}

	for _, region := range regions {
		_, err := smt.ExecContext(ctx, region.ID, region.Name)
		if err != nil {
			return errors.Wrap(err, "failed to execute region upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit region transaction")
	}
	return nil
}
