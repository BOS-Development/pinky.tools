package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingRuns struct {
	db *sql.DB
}

func NewHaulingRuns(db *sql.DB) *HaulingRuns {
	return &HaulingRuns{db: db}
}

// CreateRun inserts a new hauling run and returns the created run.
func (r *HaulingRuns) CreateRun(ctx context.Context, run *models.HaulingRun) (*models.HaulingRun, error) {
	query := `
		INSERT INTO hauling_runs (user_id, name, status, from_region_id, from_system_id, to_region_id,
			max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at`
	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		run.UserID, run.Name, run.Status, run.FromRegionID, run.FromSystemID, run.ToRegionID,
		run.MaxVolumeM3, run.HaulThresholdISK, run.NotifyTier2, run.NotifyTier3, run.DailyDigest, run.Notes,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create hauling run")
	}
	run.ID = id
	run.CreatedAt = createdAt.Format(time.RFC3339)
	run.UpdatedAt = updatedAt.Format(time.RFC3339)
	return run, nil
}

// GetRunByID returns a hauling run by ID (without items).
func (r *HaulingRuns) GetRunByID(ctx context.Context, id int64, userID int64) (*models.HaulingRun, error) {
	query := `
		SELECT id, user_id, name, status, from_region_id, from_system_id, to_region_id,
		       max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest, notes,
		       created_at, updated_at
		FROM hauling_runs WHERE id=$1 AND user_id=$2`
	var run models.HaulingRun
	var fromSystemID sql.NullInt64
	var maxVolume, haulThreshold sql.NullFloat64
	var notes sql.NullString
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&run.ID, &run.UserID, &run.Name, &run.Status, &run.FromRegionID, &fromSystemID, &run.ToRegionID,
		&maxVolume, &haulThreshold, &run.NotifyTier2, &run.NotifyTier3, &run.DailyDigest, &notes,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hauling run")
	}
	if fromSystemID.Valid {
		run.FromSystemID = &fromSystemID.Int64
	}
	if maxVolume.Valid {
		run.MaxVolumeM3 = &maxVolume.Float64
	}
	if haulThreshold.Valid {
		run.HaulThresholdISK = &haulThreshold.Float64
	}
	if notes.Valid {
		run.Notes = &notes.String
	}
	run.CreatedAt = createdAt.Format(time.RFC3339)
	run.UpdatedAt = updatedAt.Format(time.RFC3339)
	run.Items = []*models.HaulingRunItem{}
	return &run, nil
}

// ListRunsByUser returns all hauling runs for a user.
func (r *HaulingRuns) ListRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	query := `
		SELECT id, user_id, name, status, from_region_id, from_system_id, to_region_id,
		       max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest, notes,
		       created_at, updated_at
		FROM hauling_runs WHERE user_id=$1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list hauling runs")
	}
	defer rows.Close()
	runs := []*models.HaulingRun{}
	for rows.Next() {
		var run models.HaulingRun
		var fromSystemID sql.NullInt64
		var maxVolume, haulThreshold sql.NullFloat64
		var notes sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&run.ID, &run.UserID, &run.Name, &run.Status, &run.FromRegionID, &fromSystemID, &run.ToRegionID,
			&maxVolume, &haulThreshold, &run.NotifyTier2, &run.NotifyTier3, &run.DailyDigest, &notes,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan hauling run")
		}
		if fromSystemID.Valid {
			run.FromSystemID = &fromSystemID.Int64
		}
		if maxVolume.Valid {
			run.MaxVolumeM3 = &maxVolume.Float64
		}
		if haulThreshold.Valid {
			run.HaulThresholdISK = &haulThreshold.Float64
		}
		if notes.Valid {
			run.Notes = &notes.String
		}
		run.CreatedAt = createdAt.Format(time.RFC3339)
		run.UpdatedAt = updatedAt.Format(time.RFC3339)
		run.Items = []*models.HaulingRunItem{}
		runs = append(runs, &run)
	}
	return runs, nil
}

// UpdateRunStatus updates a hauling run's status.
func (r *HaulingRuns) UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE hauling_runs SET status=$1, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
		status, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to update hauling run status")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run not found")
	}
	return nil
}

// UpdateRun updates mutable fields of a hauling run.
func (r *HaulingRuns) UpdateRun(ctx context.Context, run *models.HaulingRun) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE hauling_runs SET
			name=$1, from_region_id=$2, from_system_id=$3, to_region_id=$4,
			max_volume_m3=$5, haul_threshold_isk=$6, notify_tier2=$7, notify_tier3=$8,
			daily_digest=$9, notes=$10, updated_at=NOW()
		WHERE id=$11 AND user_id=$12`,
		run.Name, run.FromRegionID, run.FromSystemID, run.ToRegionID,
		run.MaxVolumeM3, run.HaulThresholdISK, run.NotifyTier2, run.NotifyTier3,
		run.DailyDigest, run.Notes, run.ID, run.UserID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update hauling run")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run not found")
	}
	return nil
}

// DeleteRun deletes a hauling run (cascades to items).
func (r *HaulingRuns) DeleteRun(ctx context.Context, id int64, userID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM hauling_runs WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete hauling run")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("hauling run not found")
	}
	return nil
}
