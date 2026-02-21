package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
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

func (r *ItemTypeRepository) GetNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	names := map[int64]string{}
	rows, err := r.db.QueryContext(ctx, `select type_id, type_name from asset_item_types where type_id = any($1)`, pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query item type names")
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, errors.Wrap(err, "failed to scan item type name")
		}
		names[id] = name
	}

	return names, nil
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

// SearchStations searches for stations by name (case-insensitive, partial match)
func (r *ItemTypeRepository) SearchStations(ctx context.Context, query string, limit int) ([]models.StationSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	searchQuery := `
		SELECT s.station_id, s.name, ss.name AS solar_system_name
		FROM stations s
		LEFT JOIN solar_systems ss ON s.solar_system_id = ss.solar_system_id
		WHERE LOWER(s.name) LIKE LOWER($1)
		ORDER BY
			CASE
				WHEN LOWER(s.name) = LOWER($3) THEN 1
				WHEN LOWER(s.name) LIKE LOWER($3) || '%' THEN 2
				ELSE 3
			END,
			s.name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search stations")
	}
	defer rows.Close()

	results := []models.StationSearchResult{}
	for rows.Next() {
		var result models.StationSearchResult
		err := rows.Scan(&result.StationID, &result.Name, &result.SolarSystemName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan station")
		}
		results = append(results, result)
	}

	return results, nil
}
