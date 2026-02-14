package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type ItemTypeRepository struct {
	db *sql.DB
}

func NewItemTypeRepository(db *sql.DB) *ItemTypeRepository {
	return &ItemTypeRepository{
		db: db,
	}
}

func (r *ItemTypeRepository) UpsertItemTypes(ctx context.Context, itemTypes []models.EveInventoryType) error {
	if len(itemTypes) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	asset_item_types
	(
		type_id,
		type_name,
		volume,
		icon_id
	)
	values
		($1,$2,$3,$4)
on conflict
	(type_id)
do update set
	type_name = EXCLUDED.type_name,
	volume = EXCLUDED.volume,
	icon_id = EXCLUDED.icon_id
`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for item type upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for item type upsert")
	}

	for _, itemType := range itemTypes {
		_, err = smt.ExecContext(ctx,
			itemType.TypeID,
			itemType.TypeName,
			itemType.Volume,
			itemType.IconID,
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute item type upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit item type transaction")
	}

	return nil
}
