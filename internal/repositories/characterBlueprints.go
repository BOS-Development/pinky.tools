package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type CharacterBlueprints struct {
	db *sql.DB
}

func NewCharacterBlueprints(db *sql.DB) *CharacterBlueprints {
	return &CharacterBlueprints{db: db}
}

// ReplaceBlueprints deletes all existing blueprints for the given owner and inserts
// the provided set in a single transaction.
func (r *CharacterBlueprints) ReplaceBlueprints(ctx context.Context, ownerID int64, ownerType string, userID int64, blueprints []*models.CharacterBlueprint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for blueprint replace")
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`DELETE FROM character_blueprints WHERE owner_id = $1 AND owner_type = $2`,
		ownerID, ownerType,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing blueprints")
	}

	if len(blueprints) > 0 {
		insertQuery := `
			INSERT INTO character_blueprints
				(item_id, user_id, owner_id, owner_type, type_id, location_id, location_flag,
				 quantity, material_efficiency, time_efficiency, runs, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
		`

		stmt, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			return errors.Wrap(err, "failed to prepare blueprint insert")
		}

		for _, bp := range blueprints {
			_, err = stmt.ExecContext(ctx,
				bp.ItemID,
				userID,
				ownerID,
				ownerType,
				bp.TypeID,
				bp.LocationID,
				bp.LocationFlag,
				bp.Quantity,
				bp.MaterialEfficiency,
				bp.TimeEfficiency,
				bp.Runs,
			)
			if err != nil {
				return errors.Wrap(err, "failed to execute blueprint insert")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit blueprint replace transaction")
	}

	return nil
}

// GetBlueprintLevels returns the best ME/TE level for each requested type_id,
// preferring BPCs (quantity = -2) over BPOs, then highest ME.
// Returns a map of typeID → *BlueprintLevel. Types with no matching blueprint are absent.
func (r *CharacterBlueprints) GetBlueprintLevels(ctx context.Context, userID int64, typeIDs []int64) (map[int64]*models.BlueprintLevel, error) {
	if len(typeIDs) == 0 {
		return map[int64]*models.BlueprintLevel{}, nil
	}

	query := `
		SELECT DISTINCT ON (cb.type_id)
			cb.type_id,
			cb.material_efficiency,
			cb.time_efficiency,
			cb.quantity,
			cb.runs,
			COALESCE(c.name, pc.name, '') AS owner_name
		FROM character_blueprints cb
		LEFT JOIN characters c ON cb.owner_type = 'character' AND c.id = cb.owner_id
		LEFT JOIN player_corporations pc ON cb.owner_type = 'corporation' AND pc.id = cb.owner_id
		WHERE cb.user_id = $1 AND cb.type_id = ANY($2)
		ORDER BY cb.type_id,
		         CASE WHEN cb.quantity = -2 THEN 0 ELSE 1 END,
		         cb.material_efficiency DESC,
		         cb.time_efficiency DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query blueprint levels")
	}
	defer rows.Close()

	result := map[int64]*models.BlueprintLevel{}
	for rows.Next() {
		var typeID int64
		var me, te, quantity, runs int
		var ownerName string

		err = rows.Scan(&typeID, &me, &te, &quantity, &runs, &ownerName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan blueprint level")
		}

		result[typeID] = &models.BlueprintLevel{
			MaterialEfficiency: me,
			TimeEfficiency:     te,
			IsCopy:             quantity == -2,
			OwnerName:          ownerName,
			Runs:               runs,
		}
	}

	return result, nil
}

// DeleteByOwner removes all blueprints belonging to the specified owner.
func (r *CharacterBlueprints) DeleteByOwner(ctx context.Context, ownerID int64, ownerType string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM character_blueprints WHERE owner_id = $1 AND owner_type = $2`,
		ownerID, ownerType,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete blueprints by owner")
	}
	return nil
}
