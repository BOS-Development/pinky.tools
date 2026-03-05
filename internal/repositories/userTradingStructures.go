package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type UserTradingStructures struct {
	db *sql.DB
}

func NewUserTradingStructures(db *sql.DB) *UserTradingStructures {
	return &UserTradingStructures{db: db}
}

// List returns all structures for a user.
func (r *UserTradingStructures) List(ctx context.Context, userID int64) ([]*models.UserTradingStructure, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, structure_id, name, system_id, region_id, character_id, access_ok, last_scanned_at, created_at
		FROM user_trading_structures
		WHERE user_id = $1
		ORDER BY name ASC`, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list user trading structures")
	}
	defer rows.Close()

	structures := []*models.UserTradingStructure{}
	for rows.Next() {
		var s models.UserTradingStructure
		var lastScannedAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StructureID, &s.Name, &s.SystemID, &s.RegionID,
			&s.CharacterID, &s.AccessOK, &lastScannedAt, &createdAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan user trading structure")
		}
		if lastScannedAt.Valid {
			ts := lastScannedAt.Time.Format(time.RFC3339)
			s.LastScannedAt = &ts
		}
		s.CreatedAt = createdAt.Format(time.RFC3339)
		structures = append(structures, &s)
	}
	return structures, nil
}

// Upsert creates or updates a structure entry.
func (r *UserTradingStructures) Upsert(ctx context.Context, s *models.UserTradingStructure) (*models.UserTradingStructure, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var id int64
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO user_trading_structures (user_id, structure_id, name, system_id, region_id, character_id, access_ok, last_scanned_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (user_id, structure_id) DO UPDATE SET
			name = EXCLUDED.name,
			system_id = EXCLUDED.system_id,
			region_id = EXCLUDED.region_id,
			character_id = EXCLUDED.character_id,
			access_ok = EXCLUDED.access_ok,
			last_scanned_at = EXCLUDED.last_scanned_at,
			updated_at = NOW()
		RETURNING id, created_at`,
		s.UserID, s.StructureID, s.Name, s.SystemID, s.RegionID, s.CharacterID, s.AccessOK, nil,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert user trading structure")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit upsert")
	}

	s.ID = id
	s.CreatedAt = createdAt.Format(time.RFC3339)
	return s, nil
}

// Delete removes a structure by id+userID.
func (r *UserTradingStructures) Delete(ctx context.Context, id int64, userID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM user_trading_structures WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete user trading structure")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("user trading structure not found")
	}
	return nil
}

// UpdateAccessStatus updates access_ok for a structure by user+structure.
func (r *UserTradingStructures) UpdateAccessStatus(ctx context.Context, userID int64, structureID int64, accessOK bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_trading_structures SET access_ok=$1, updated_at=NOW() WHERE user_id=$2 AND structure_id=$3`,
		accessOK, userID, structureID)
	if err != nil {
		return errors.Wrap(err, "failed to update access status")
	}
	return nil
}
