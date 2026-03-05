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
			from_station_id, to_station_id,
			max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, updated_at`
	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		run.UserID, run.Name, run.Status, run.FromRegionID, run.FromSystemID, run.ToRegionID,
		run.FromStationID, run.ToStationID,
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
	row := r.db.QueryRowContext(ctx, listRunsColumns+` WHERE id=$1 AND user_id=$2`, id, userID)
	run, err := scanHaulingRun(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hauling run")
	}
	return run, nil
}

// scanHaulingRun scans a row into a HaulingRun, including optional station IDs.
func scanHaulingRun(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.HaulingRun, error) {
	var run models.HaulingRun
	var fromSystemID, fromStationID, toStationID sql.NullInt64
	var maxVolume, haulThreshold sql.NullFloat64
	var notes sql.NullString
	var completedAt sql.NullTime
	var createdAt, updatedAt time.Time
	if err := scanner.Scan(
		&run.ID, &run.UserID, &run.Name, &run.Status, &run.FromRegionID, &fromSystemID, &run.ToRegionID,
		&fromStationID, &toStationID,
		&maxVolume, &haulThreshold, &run.NotifyTier2, &run.NotifyTier3, &run.DailyDigest, &notes,
		&completedAt, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	if fromSystemID.Valid {
		run.FromSystemID = &fromSystemID.Int64
	}
	if fromStationID.Valid {
		run.FromStationID = &fromStationID.Int64
	}
	if toStationID.Valid {
		run.ToStationID = &toStationID.Int64
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
	if completedAt.Valid {
		s := completedAt.Time.Format(time.RFC3339)
		run.CompletedAt = &s
	}
	run.CreatedAt = createdAt.Format(time.RFC3339)
	run.UpdatedAt = updatedAt.Format(time.RFC3339)
	run.Items = []*models.HaulingRunItem{}
	return &run, nil
}

// listRunsQuery is the shared SELECT column list for hauling runs.
const listRunsColumns = `
	SELECT id, user_id, name, status, from_region_id, from_system_id, to_region_id,
	       from_station_id, to_station_id,
	       max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest, notes,
	       completed_at, created_at, updated_at
	FROM hauling_runs`

// ListRunsByUser returns all hauling runs for a user.
func (r *HaulingRuns) ListRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	rows, err := r.db.QueryContext(ctx, listRunsColumns+` WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list hauling runs")
	}
	defer rows.Close()
	runs := []*models.HaulingRun{}
	for rows.Next() {
		run, err := scanHaulingRun(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan hauling run")
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// UpdateRunStatus updates a hauling run's status.
// When status is set to 'COMPLETE', completed_at is also set to NOW().
func (r *HaulingRuns) UpdateRunStatus(ctx context.Context, id int64, userID int64, status string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE hauling_runs SET status=$1, completed_at = CASE WHEN $1 = 'COMPLETE' THEN NOW() ELSE completed_at END, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
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
			from_station_id=$5, to_station_id=$6,
			max_volume_m3=$7, haul_threshold_isk=$8, notify_tier2=$9, notify_tier3=$10,
			daily_digest=$11, notes=$12, updated_at=NOW()
		WHERE id=$13 AND user_id=$14`,
		run.Name, run.FromRegionID, run.FromSystemID, run.ToRegionID,
		run.FromStationID, run.ToStationID,
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

// ListAccumulatingByUser returns all ACCUMULATING runs for a user.
func (r *HaulingRuns) ListAccumulatingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	rows, err := r.db.QueryContext(ctx, listRunsColumns+` WHERE user_id=$1 AND status='ACCUMULATING' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list accumulating runs")
	}
	defer rows.Close()
	runs := []*models.HaulingRun{}
	for rows.Next() {
		run, err := scanHaulingRun(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan accumulating run")
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// ListDigestRunsByUser returns active runs with daily_digest=true (excludes COMPLETE/CANCELLED).
func (r *HaulingRuns) ListDigestRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	rows, err := r.db.QueryContext(ctx,
		listRunsColumns+` WHERE user_id=$1 AND daily_digest=true AND status NOT IN ('COMPLETE', 'CANCELLED') ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list digest runs")
	}
	defer rows.Close()
	runs := []*models.HaulingRun{}
	for rows.Next() {
		run, err := scanHaulingRun(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan digest run")
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// ListSellingByUser returns all SELLING-status runs for a user.
func (r *HaulingRuns) ListSellingByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	rows, err := r.db.QueryContext(ctx, listRunsColumns+` WHERE user_id=$1 AND status='SELLING' ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list selling runs")
	}
	defer rows.Close()
	runs := []*models.HaulingRun{}
	for rows.Next() {
		run, err := scanHaulingRun(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan selling run")
		}
		runs = append(runs, run)
	}
	return runs, nil
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
