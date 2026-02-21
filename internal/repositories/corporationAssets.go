package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type CorporationAssets struct {
	db *sql.DB
}

func NewCorporationAssets(db *sql.DB) *CorporationAssets {
	return &CorporationAssets{
		db: db,
	}
}

func (r *CorporationAssets) Upsert(ctx context.Context, corp, user int64, assets []*models.EveAsset) error {
	if len(assets) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	corporation_assets
	(
		corporation_id,
		user_id,
		item_id,
		is_blueprint_copy,
		is_singleton,
		location_id,
		location_type,
		quantity,
		type_id,
		location_flag,
		update_key
	)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
on conflict
	(corporation_id, user_id, item_id)
do update set
	is_blueprint_copy = EXCLUDED.is_blueprint_copy,
	is_singleton = EXCLUDED.is_singleton,
	location_id = EXCLUDED.location_id,
	location_type = EXCLUDED.location_type,
	quantity = EXCLUDED.quantity,
	type_id = EXCLUDED.type_id,
	location_flag = EXCLUDED.location_flag,
	update_key = EXCLUDED.update_key;`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for corporation asset insert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for corporation asset insert")
	}

	updateKey := time.Now()

	for _, asset := range assets {
		_, err = smt.ExecContext(ctx,
			corp,
			user,
			asset.ItemID,
			asset.IsBlueprintCopy,
			asset.IsSingleton,
			asset.LocationID,
			asset.LocationType,
			asset.Quantity,
			asset.TypeID,
			asset.LocationFlag,
			updateKey)
		if err != nil {
			return errors.Wrap(err, "failed to execute corporation asset upsert")
		}
	}

	_, err = tx.ExecContext(ctx, `
delete from
	corporation_assets
where
	user_id=$2 and 
	corporation_id=$1 and 
	update_key!=$3;`, corp, user, updateKey)
	if err != nil {
		return errors.Wrap(err, "failed to delete from corporation assets")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit corporation asset update transaction")
	}

	return nil
}

func (r *CorporationAssets) GetAssembledContainers(ctx context.Context, corp, user int64) ([]int64, error) {
	query := `
SELECT
    corpAssets.item_id
FROM
    corporation_assets corpAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=corpAssets.type_id

WHERE
    assetTypes.type_name like '%Container' AND
    corporation_id=$1 AND
    user_id=$2 AND
    is_singleton=true;
	`

	rows, err := r.db.QueryContext(ctx, query, corp, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query containers")
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan container row")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *CorporationAssets) UpsertContainerNames(ctx context.Context, corp, user int64, locationNames map[int64]string) error {
	if len(locationNames) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	corporation_asset_location_names
	(
		corporation_id,
		user_id,
		item_id,
		name
	)
	values
	($1, $2, $3, $4)
on conflict
	(corporation_id, user_id, item_id)
do update set
	name = EXCLUDED.name;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for container name upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for container name upsert")
	}

	for id, name := range locationNames {
		_, err := smt.ExecContext(ctx, corp, user, id, name)
		if err != nil {
			return errors.Wrap(err, "failed to execute container name upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit container name transaction")
	}

	return nil
}

func (r *CorporationAssets) GetPlayerOwnedStationIDs(ctx context.Context, corp, user int64) ([]int64, error) {
	query := `
SELECT DISTINCT
    corporation_assets.location_id
FROM
    corporation_assets corporation_assets
WHERE
	corporation_assets.corporation_id=$1 AND
	corporation_assets.user_id=$2 AND
    location_flag='OfficeFolder';
	`

	rows, err := r.db.QueryContext(ctx, query, corp, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query containers")
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan container row")
		}
		ids = append(ids, id)
	}
	return ids, nil
}
