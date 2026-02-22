package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type IndustryJobs struct {
	db *sql.DB
}

func NewIndustryJobs(db *sql.DB) *IndustryJobs {
	return &IndustryJobs{db: db}
}

func (r *IndustryJobs) UpsertJobs(ctx context.Context, userID int64, jobs []*models.IndustryJob) error {
	if len(jobs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for industry jobs upsert")
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO esi_industry_jobs
			(job_id, installer_id, user_id, facility_id, station_id, activity_id,
			 blueprint_id, blueprint_type_id, blueprint_location_id, output_location_id,
			 runs, cost, licensed_runs, probability, product_type_id, status,
			 duration, start_date, end_date, pause_date, completed_date,
			 completed_character_id, successful_runs, solar_system_id, source, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, now())
		ON CONFLICT (job_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			pause_date = EXCLUDED.pause_date,
			completed_date = EXCLUDED.completed_date,
			completed_character_id = EXCLUDED.completed_character_id,
			successful_runs = EXCLUDED.successful_runs,
			source = EXCLUDED.source,
			updated_at = now()
	`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare industry job upsert")
	}

	for _, job := range jobs {
		_, err = stmt.ExecContext(ctx,
			job.JobID,
			job.InstallerID,
			userID,
			job.FacilityID,
			job.StationID,
			job.ActivityID,
			job.BlueprintID,
			job.BlueprintTypeID,
			job.BlueprintLocationID,
			job.OutputLocationID,
			job.Runs,
			job.Cost,
			job.LicensedRuns,
			job.Probability,
			job.ProductTypeID,
			job.Status,
			job.Duration,
			job.StartDate,
			job.EndDate,
			job.PauseDate,
			job.CompletedDate,
			job.CompletedCharacterID,
			job.SuccessfulRuns,
			job.SolarSystemID,
			job.Source,
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute industry job upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit industry jobs transaction")
	}
	return nil
}

// GetActiveJobs returns active (non-delivered, non-cancelled) ESI industry jobs for a user,
// enriched with blueprint/product/system names.
func (r *IndustryJobs) GetActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	query := `
		SELECT j.job_id, j.installer_id, j.user_id, j.facility_id, j.station_id,
		       j.activity_id, j.blueprint_id, j.blueprint_type_id, j.blueprint_location_id,
		       j.output_location_id, j.runs, j.cost, j.licensed_runs, j.probability,
		       j.product_type_id, j.status, j.duration, j.start_date, j.end_date,
		       j.pause_date, j.completed_date, j.completed_character_id, j.successful_runs,
		       j.solar_system_id, j.source, j.updated_at,
		       COALESCE(bp.type_name, ''),
		       COALESCE(prod.type_name, ''),
		       COALESCE(c.name, ''),
		       COALESCE(ss.name, ss2.name, '')
		FROM esi_industry_jobs j
		LEFT JOIN asset_item_types bp ON bp.type_id = j.blueprint_type_id
		LEFT JOIN asset_item_types prod ON prod.type_id = j.product_type_id
		LEFT JOIN characters c ON c.id = j.installer_id
		LEFT JOIN solar_systems ss ON ss.solar_system_id = j.solar_system_id
		LEFT JOIN stations st ON st.station_id = j.station_id
		LEFT JOIN solar_systems ss2 ON ss2.solar_system_id = st.solar_system_id
		WHERE j.user_id = $1
		  AND j.status IN ('active', 'paused', 'ready')
		ORDER BY j.end_date ASC
	`

	return r.queryJobs(ctx, query, userID)
}

// GetAllJobs returns all ESI industry jobs for a user (including completed/cancelled).
func (r *IndustryJobs) GetAllJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	query := `
		SELECT j.job_id, j.installer_id, j.user_id, j.facility_id, j.station_id,
		       j.activity_id, j.blueprint_id, j.blueprint_type_id, j.blueprint_location_id,
		       j.output_location_id, j.runs, j.cost, j.licensed_runs, j.probability,
		       j.product_type_id, j.status, j.duration, j.start_date, j.end_date,
		       j.pause_date, j.completed_date, j.completed_character_id, j.successful_runs,
		       j.solar_system_id, j.source, j.updated_at,
		       COALESCE(bp.type_name, ''),
		       COALESCE(prod.type_name, ''),
		       COALESCE(c.name, ''),
		       COALESCE(ss.name, ss2.name, '')
		FROM esi_industry_jobs j
		LEFT JOIN asset_item_types bp ON bp.type_id = j.blueprint_type_id
		LEFT JOIN asset_item_types prod ON prod.type_id = j.product_type_id
		LEFT JOIN characters c ON c.id = j.installer_id
		LEFT JOIN solar_systems ss ON ss.solar_system_id = j.solar_system_id
		LEFT JOIN stations st ON st.station_id = j.station_id
		LEFT JOIN solar_systems ss2 ON ss2.solar_system_id = st.solar_system_id
		WHERE j.user_id = $1
		ORDER BY j.start_date DESC
	`

	return r.queryJobs(ctx, query, userID)
}

// GetActiveJobsForMatching returns active ESI jobs for queue matching (no enriched names needed).
func (r *IndustryJobs) GetActiveJobsForMatching(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	query := `
		SELECT job_id, installer_id, user_id, facility_id, station_id,
		       activity_id, blueprint_id, blueprint_type_id, blueprint_location_id,
		       output_location_id, runs, cost, licensed_runs, probability,
		       product_type_id, status, duration, start_date, end_date,
		       pause_date, completed_date, completed_character_id, successful_runs,
		       solar_system_id, source, updated_at,
		       '', '', '', ''
		FROM esi_industry_jobs
		WHERE user_id = $1
		  AND status = 'active'
		ORDER BY start_date DESC
	`

	return r.queryJobs(ctx, query, userID)
}

func (r *IndustryJobs) queryJobs(ctx context.Context, query string, args ...interface{}) ([]*models.IndustryJob, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query industry jobs")
	}
	defer rows.Close()

	jobs := []*models.IndustryJob{}
	for rows.Next() {
		var job models.IndustryJob
		err = rows.Scan(
			&job.JobID,
			&job.InstallerID,
			&job.UserID,
			&job.FacilityID,
			&job.StationID,
			&job.ActivityID,
			&job.BlueprintID,
			&job.BlueprintTypeID,
			&job.BlueprintLocationID,
			&job.OutputLocationID,
			&job.Runs,
			&job.Cost,
			&job.LicensedRuns,
			&job.Probability,
			&job.ProductTypeID,
			&job.Status,
			&job.Duration,
			&job.StartDate,
			&job.EndDate,
			&job.PauseDate,
			&job.CompletedDate,
			&job.CompletedCharacterID,
			&job.SuccessfulRuns,
			&job.SolarSystemID,
			&job.Source,
			&job.UpdatedAt,
			&job.BlueprintName,
			&job.ProductName,
			&job.InstallerName,
			&job.SystemName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan industry job")
		}

		job.ActivityName = activityName(job.ActivityID)

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetJobByID returns a single ESI industry job by its job_id.
func (r *IndustryJobs) GetJobByID(ctx context.Context, jobID int64) (*models.IndustryJob, error) {
	query := `
		SELECT job_id, installer_id, user_id, facility_id, station_id,
		       activity_id, blueprint_id, blueprint_type_id, blueprint_location_id,
		       output_location_id, runs, cost, licensed_runs, probability,
		       product_type_id, status, duration, start_date, end_date,
		       pause_date, completed_date, completed_character_id, successful_runs,
		       solar_system_id, source, updated_at
		FROM esi_industry_jobs
		WHERE job_id = $1
	`

	var job models.IndustryJob
	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.JobID,
		&job.InstallerID,
		&job.UserID,
		&job.FacilityID,
		&job.StationID,
		&job.ActivityID,
		&job.BlueprintID,
		&job.BlueprintTypeID,
		&job.BlueprintLocationID,
		&job.OutputLocationID,
		&job.Runs,
		&job.Cost,
		&job.LicensedRuns,
		&job.Probability,
		&job.ProductTypeID,
		&job.Status,
		&job.Duration,
		&job.StartDate,
		&job.EndDate,
		&job.PauseDate,
		&job.CompletedDate,
		&job.CompletedCharacterID,
		&job.SuccessfulRuns,
		&job.SolarSystemID,
		&job.Source,
		&job.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get industry job by id")
	}

	job.ActivityName = activityName(job.ActivityID)
	return &job, nil
}

// DeleteOldDeliveredJobs removes delivered/cancelled jobs older than the given cutoff.
func (r *IndustryJobs) DeleteOldDeliveredJobs(ctx context.Context, userID int64, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM esi_industry_jobs
		WHERE user_id = $1
		  AND status IN ('delivered', 'cancelled', 'reverted')
		  AND updated_at < $2
	`, userID, before)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete old delivered jobs")
	}
	return result.RowsAffected()
}

func activityName(activityID int) string {
	switch activityID {
	case 1:
		return "Manufacturing"
	case 3:
		return "TE Research"
	case 4:
		return "ME Research"
	case 5:
		return "Copying"
	case 8:
		return "Invention"
	case 9:
		return "Reaction"
	default:
		return "Unknown"
	}
}
