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
		icon_id,
		group_id,
		packaged_volume,
		mass,
		capacity,
		portion_size,
		published,
		market_group_id,
		graphic_id,
		race_id,
		description
	)
	values
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
on conflict
	(type_id)
do update set
	type_name = EXCLUDED.type_name,
	volume = EXCLUDED.volume,
	icon_id = EXCLUDED.icon_id,
	group_id = EXCLUDED.group_id,
	packaged_volume = EXCLUDED.packaged_volume,
	mass = EXCLUDED.mass,
	capacity = EXCLUDED.capacity,
	portion_size = EXCLUDED.portion_size,
	published = EXCLUDED.published,
	market_group_id = EXCLUDED.market_group_id,
	graphic_id = EXCLUDED.graphic_id,
	race_id = EXCLUDED.race_id,
	description = EXCLUDED.description
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
			itemType.GroupID,
			itemType.PackagedVolume,
			itemType.Mass,
			itemType.Capacity,
			itemType.PortionSize,
			itemType.Published,
			itemType.MarketGroupID,
			itemType.GraphicID,
			itemType.RaceID,
			itemType.Description,
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

// SearchItemTypes searches for item types by name (case-insensitive, partial match)
func (r *ItemTypeRepository) SearchItemTypes(ctx context.Context, query string, limit int) ([]models.EveInventoryType, error) {
	if limit <= 0 {
		limit = 20
	}

	searchQuery := `
		SELECT type_id, type_name, volume, icon_id
		FROM asset_item_types
		WHERE LOWER(type_name) LIKE LOWER($1)
		ORDER BY
			CASE
				WHEN LOWER(type_name) = LOWER($3) THEN 1
				WHEN LOWER(type_name) LIKE LOWER($3) || '%' THEN 2
				ELSE 3
			END,
			type_name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search item types")
	}
	defer rows.Close()

	var items []models.EveInventoryType
	for rows.Next() {
		var item models.EveInventoryType
		err := rows.Scan(&item.TypeID, &item.TypeName, &item.Volume, &item.IconID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item type")
		}
		items = append(items, item)
	}

	return items, nil
}

// GetItemTypeByName gets an exact item type by name
func (r *ItemTypeRepository) GetItemTypeByName(ctx context.Context, typeName string) (*models.EveInventoryType, error) {
	query := `
		SELECT type_id, type_name, volume, icon_id
		FROM asset_item_types
		WHERE type_name = $1
	`

	var item models.EveInventoryType
	err := r.db.QueryRowContext(ctx, query, typeName).Scan(
		&item.TypeID,
		&item.TypeName,
		&item.Volume,
		&item.IconID,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("item type not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item type")
	}

	return &item, nil
}
