package repositories

import (
	"context"
	"database/sql"
	"math/rand"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type CharacterAsset struct {
	CharacterID     int64
	UserID          int64
	ItemID          int64
	IsBlueprintCopy bool
	IsSingleton     bool
	LocationID      int64
	LocationType    string
	Quantity        int64
	TypeID          int64
	LocationFlag    string
}

type CharacterAssets struct {
	db *sql.DB
}

func NewCharacterAssets(db *sql.DB) *CharacterAssets {
	return &CharacterAssets{
		db: db,
	}
}

func (r *CharacterAssets) Get(ctx context.Context, userID, characterID int64) ([]*CharacterAsset, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	character_id,
	user_id,
	item_id,
	is_blueprint_copy,
	is_singleton,
	location_id,
	location_type,
	quantity,
	type_id,
	location_flag
from
	character_assets
where
	character_id = $1 and
	user_id = $2
	`, characterID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query character assets from database")
	}
	defer rows.Close()

	assets := []*CharacterAsset{}
	for rows.Next() {
		asset := CharacterAsset{}
		err = rows.Scan(
			&asset.CharacterID,
			&asset.UserID,
			&asset.ItemID,
			&asset.IsBlueprintCopy,
			&asset.IsSingleton,
			&asset.LocationID,
			&asset.LocationType,
			&asset.Quantity,
			&asset.TypeID,
			&asset.LocationFlag)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan character asset")
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (r *CharacterAssets) UpdateAssets(ctx context.Context, characterID, userID int64, assets []*models.EveAsset) error {
	if len(assets) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	character_assets
	(
		character_id,
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
	values
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
on conflict
	(character_id, user_id, item_id)
do update set
	is_blueprint_copy = EXCLUDED.is_blueprint_copy,
	is_singleton = EXCLUDED.is_singleton,
	location_id = EXCLUDED.location_id,
	location_type = EXCLUDED.location_type,
	quantity = EXCLUDED.quantity,
	type_id = EXCLUDED.type_id,
	location_flag = EXCLUDED.location_flag,
	update_key = EXCLUDED.update_key;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for character asset insert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for character asset insert")
	}

	updateKey := strconv.Itoa(rand.Int())

	for _, asset := range assets {
		_, err = smt.ExecContext(ctx,
			characterID,
			userID,
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
			return errors.Wrap(err, "failed to execute character asset upsert")
		}
	}

	_, err = tx.ExecContext(ctx, `
delete from
	character_assets
where
	user_id=$1 and 
	character_id=$2 and 
	update_key!=$3;`, userID, characterID, updateKey)
	if err != nil {
		return errors.Wrap(err, "failed to delete from character assets")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit character asset update transaction")
	}

	return nil
}

func (r *CharacterAssets) GetAssembledContainers(ctx context.Context, character, user int64) ([]int64, error) {
	query := `
SELECT
    characterAssets.item_id
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id

WHERE
    assetTypes.type_name like '%Container' AND
    character_id=$1 AND
    user_id=$2 AND
    is_singleton=true;
	`

	rows, err := r.db.QueryContext(ctx, query, character, user)
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

func (r *CharacterAssets) UpsertContainerNames(ctx context.Context, characterID, userID int64, locationNames map[int64]string) error {
	if len(locationNames) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	character_asset_location_names
	(
		character_id,
		user_id,
		item_id,
		name
	)
	values
	($1, $2, $3, $4)
on conflict
	(character_id, user_id, item_id)
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
		_, err := smt.ExecContext(ctx, characterID, userID, id, name)
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

func (r *CharacterAssets) GetPlayerOwnedStationIDs(ctx context.Context, character, user int64) ([]int64, error) {
	query := `
SELECT DISTINCT
    characterAssets.location_id
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id
WHERE
	characterAssets.character_id=$1 AND
	characterAssets.user_id=$2 AND
    location_flag='Hangar' AND
    location_type='item';
	`

	rows, err := r.db.QueryContext(ctx, query, character, user)
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
