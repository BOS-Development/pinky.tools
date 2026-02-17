package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type IndustryCostIndices struct {
	db *sql.DB
}

func NewIndustryCostIndices(db *sql.DB) *IndustryCostIndices {
	return &IndustryCostIndices{db: db}
}

func (r *IndustryCostIndices) UpsertIndices(ctx context.Context, indices []models.IndustryCostIndex) error {
	if len(indices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for cost indices upsert")
	}
	defer tx.Rollback()

	upsertQuery := `
INSERT INTO industry_cost_indices (system_id, activity, cost_index, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (system_id, activity) DO UPDATE SET cost_index = EXCLUDED.cost_index, updated_at = NOW()
`

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare cost indices upsert")
	}

	for _, idx := range indices {
		_, err = smt.ExecContext(ctx, idx.SystemID, idx.Activity, idx.CostIndex)
		if err != nil {
			return errors.Wrap(err, "failed to upsert cost index")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit cost indices transaction")
	}

	return nil
}

func (r *IndustryCostIndices) GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error) {
	query := `
SELECT system_id, activity, cost_index, updated_at
FROM industry_cost_indices
WHERE system_id = $1 AND activity = $2
`

	var idx models.IndustryCostIndex
	err := r.db.QueryRowContext(ctx, query, systemID, activity).Scan(
		&idx.SystemID,
		&idx.Activity,
		&idx.CostIndex,
		&idx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cost index")
	}

	return &idx, nil
}

func (r *IndustryCostIndices) GetLastUpdateTime(ctx context.Context) (*time.Time, error) {
	query := `SELECT MAX(updated_at) FROM industry_cost_indices`

	var lastUpdate *time.Time
	err := r.db.QueryRowContext(ctx, query).Scan(&lastUpdate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query last cost indices update time")
	}

	return lastUpdate, nil
}
