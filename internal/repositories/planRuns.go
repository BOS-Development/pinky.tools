package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type PlanRuns struct {
	db *sql.DB
}

func NewPlanRuns(db *sql.DB) *PlanRuns {
	return &PlanRuns{db: db}
}

func (r *PlanRuns) Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error) {
	query := `
		INSERT INTO production_plan_runs (plan_id, user_id, quantity)
		VALUES ($1, $2, $3)
		RETURNING id, plan_id, user_id, quantity, created_at
	`

	var created models.ProductionPlanRun
	err := r.db.QueryRowContext(ctx, query,
		run.PlanID,
		run.UserID,
		run.Quantity,
	).Scan(
		&created.ID,
		&created.PlanID,
		&created.UserID,
		&created.Quantity,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create plan run")
	}

	return &created, nil
}

func (r *PlanRuns) GetByPlan(ctx context.Context, planID, userID int64) ([]*models.ProductionPlanRun, error) {
	query := `
		SELECT r.id, r.plan_id, r.user_id, r.quantity, r.created_at,
		       COALESCE(p.name, ''),
		       COALESCE(t.type_name, ''),
		       COALESCE(counts.total, 0),
		       COALESCE(counts.planned, 0),
		       COALESCE(counts.active, 0),
		       COALESCE(counts.completed, 0),
		       COALESCE(counts.cancelled, 0),
		       CASE
		         WHEN COALESCE(counts.total, 0) = 0 THEN 'pending'
		         WHEN COALESCE(counts.completed, 0) + COALESCE(counts.cancelled, 0) = counts.total THEN 'completed'
		         WHEN COALESCE(counts.active, 0) > 0 OR COALESCE(counts.completed, 0) > 0 THEN 'in_progress'
		         ELSE 'pending'
		       END
		FROM production_plan_runs r
		JOIN production_plans p ON p.id = r.plan_id
		LEFT JOIN asset_item_types t ON t.type_id = p.product_type_id
		LEFT JOIN LATERAL (
		  SELECT count(*) AS total,
		         count(*) FILTER (WHERE q.status = 'planned') AS planned,
		         count(*) FILTER (WHERE q.status = 'active') AS active,
		         count(*) FILTER (WHERE q.status = 'completed') AS completed,
		         count(*) FILTER (WHERE q.status = 'cancelled') AS cancelled
		  FROM industry_job_queue q
		  WHERE q.plan_run_id = r.id
		) counts ON true
		WHERE r.plan_id = $1 AND r.user_id = $2
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, planID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plan runs")
	}
	defer rows.Close()

	runs := []*models.ProductionPlanRun{}
	for rows.Next() {
		var run models.ProductionPlanRun
		var summary models.PlanRunJobSummary
		err = rows.Scan(
			&run.ID,
			&run.PlanID,
			&run.UserID,
			&run.Quantity,
			&run.CreatedAt,
			&run.PlanName,
			&run.ProductName,
			&summary.Total,
			&summary.Planned,
			&summary.Active,
			&summary.Completed,
			&summary.Cancelled,
			&run.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan plan run")
		}
		run.JobSummary = &summary
		runs = append(runs, &run)
	}

	return runs, nil
}

func (r *PlanRuns) GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlanRun, error) {
	query := `
		SELECT r.id, r.plan_id, r.user_id, r.quantity, r.created_at,
		       COALESCE(p.name, ''),
		       COALESCE(t.type_name, ''),
		       COALESCE(counts.total, 0),
		       COALESCE(counts.planned, 0),
		       COALESCE(counts.active, 0),
		       COALESCE(counts.completed, 0),
		       COALESCE(counts.cancelled, 0),
		       CASE
		         WHEN COALESCE(counts.total, 0) = 0 THEN 'pending'
		         WHEN COALESCE(counts.completed, 0) + COALESCE(counts.cancelled, 0) = counts.total THEN 'completed'
		         WHEN COALESCE(counts.active, 0) > 0 OR COALESCE(counts.completed, 0) > 0 THEN 'in_progress'
		         ELSE 'pending'
		       END
		FROM production_plan_runs r
		JOIN production_plans p ON p.id = r.plan_id
		LEFT JOIN asset_item_types t ON t.type_id = p.product_type_id
		LEFT JOIN LATERAL (
		  SELECT count(*) AS total,
		         count(*) FILTER (WHERE q.status = 'planned') AS planned,
		         count(*) FILTER (WHERE q.status = 'active') AS active,
		         count(*) FILTER (WHERE q.status = 'completed') AS completed,
		         count(*) FILTER (WHERE q.status = 'cancelled') AS cancelled
		  FROM industry_job_queue q
		  WHERE q.plan_run_id = r.id
		) counts ON true
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plan runs by user")
	}
	defer rows.Close()

	runs := []*models.ProductionPlanRun{}
	for rows.Next() {
		var run models.ProductionPlanRun
		var summary models.PlanRunJobSummary
		err = rows.Scan(
			&run.ID,
			&run.PlanID,
			&run.UserID,
			&run.Quantity,
			&run.CreatedAt,
			&run.PlanName,
			&run.ProductName,
			&summary.Total,
			&summary.Planned,
			&summary.Active,
			&summary.Completed,
			&summary.Cancelled,
			&run.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan plan run")
		}
		run.JobSummary = &summary
		runs = append(runs, &run)
	}

	return runs, nil
}

func (r *PlanRuns) CancelPlannedJobs(ctx context.Context, runID, userID int64) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE industry_job_queue
		SET status = 'cancelled', updated_at = now()
		WHERE plan_run_id = $1 AND user_id = $2 AND status = 'planned'
	`, runID, userID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to cancel planned jobs for run")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get rows affected")
	}

	return rows, nil
}

func (r *PlanRuns) GetByID(ctx context.Context, runID, userID int64) (*models.ProductionPlanRun, error) {
	query := `
		SELECT r.id, r.plan_id, r.user_id, r.quantity, r.created_at,
		       COALESCE(p.name, ''),
		       COALESCE(t.type_name, ''),
		       COALESCE(counts.total, 0),
		       COALESCE(counts.planned, 0),
		       COALESCE(counts.active, 0),
		       COALESCE(counts.completed, 0),
		       COALESCE(counts.cancelled, 0),
		       CASE
		         WHEN COALESCE(counts.total, 0) = 0 THEN 'pending'
		         WHEN COALESCE(counts.completed, 0) + COALESCE(counts.cancelled, 0) = counts.total THEN 'completed'
		         WHEN COALESCE(counts.active, 0) > 0 OR COALESCE(counts.completed, 0) > 0 THEN 'in_progress'
		         ELSE 'pending'
		       END
		FROM production_plan_runs r
		JOIN production_plans p ON p.id = r.plan_id
		LEFT JOIN asset_item_types t ON t.type_id = p.product_type_id
		LEFT JOIN LATERAL (
		  SELECT count(*) AS total,
		         count(*) FILTER (WHERE q.status = 'planned') AS planned,
		         count(*) FILTER (WHERE q.status = 'active') AS active,
		         count(*) FILTER (WHERE q.status = 'completed') AS completed,
		         count(*) FILTER (WHERE q.status = 'cancelled') AS cancelled
		  FROM industry_job_queue q
		  WHERE q.plan_run_id = r.id
		) counts ON true
		WHERE r.id = $1 AND r.user_id = $2
	`

	var run models.ProductionPlanRun
	var summary models.PlanRunJobSummary
	err := r.db.QueryRowContext(ctx, query, runID, userID).Scan(
		&run.ID,
		&run.PlanID,
		&run.UserID,
		&run.Quantity,
		&run.CreatedAt,
		&run.PlanName,
		&run.ProductName,
		&summary.Total,
		&summary.Planned,
		&summary.Active,
		&summary.Completed,
		&summary.Cancelled,
		&run.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get plan run")
	}
	run.JobSummary = &summary

	// Fetch jobs for this run
	jobQuery := `
		SELECT q.id, q.user_id, q.character_id, q.blueprint_type_id, q.activity, q.runs,
		       q.me_level, q.te_level, q.system_id, q.facility_tax, q.status, q.esi_job_id,
		       q.product_type_id, q.estimated_cost, q.estimated_duration, q.notes,
		       q.plan_run_id, q.plan_step_id, q.created_at, q.updated_at,
		       COALESCE(bp.type_name, ''),
		       COALESCE(prod.type_name, ''),
		       COALESCE(c.name, ''),
		       COALESCE(ss.name, ''),
		       j.end_date,
		       COALESCE(j.source, '')
		FROM industry_job_queue q
		LEFT JOIN asset_item_types bp ON bp.type_id = q.blueprint_type_id
		LEFT JOIN asset_item_types prod ON prod.type_id = q.product_type_id
		LEFT JOIN characters c ON c.id = q.character_id
		LEFT JOIN solar_systems ss ON ss.solar_system_id = q.system_id
		LEFT JOIN esi_industry_jobs j ON j.job_id = q.esi_job_id
		WHERE q.plan_run_id = $1
		ORDER BY q.created_at ASC
	`

	jobRows, err := r.db.QueryContext(ctx, jobQuery, runID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plan run jobs")
	}
	defer jobRows.Close()

	run.Jobs = []*models.IndustryJobQueueEntry{}
	for jobRows.Next() {
		var entry models.IndustryJobQueueEntry
		err = jobRows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.CharacterID,
			&entry.BlueprintTypeID,
			&entry.Activity,
			&entry.Runs,
			&entry.MELevel,
			&entry.TELevel,
			&entry.SystemID,
			&entry.FacilityTax,
			&entry.Status,
			&entry.EsiJobID,
			&entry.ProductTypeID,
			&entry.EstimatedCost,
			&entry.EstimatedDuration,
			&entry.Notes,
			&entry.PlanRunID,
			&entry.PlanStepID,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.BlueprintName,
			&entry.ProductName,
			&entry.CharacterName,
			&entry.SystemName,
			&entry.EsiJobEndDate,
			&entry.EsiJobSource,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan plan run job")
		}
		run.Jobs = append(run.Jobs, &entry)
	}

	return &run, nil
}

func (r *PlanRuns) Delete(ctx context.Context, runID, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM production_plan_runs
		WHERE id = $1 AND user_id = $2
	`, runID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete plan run")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for delete plan run")
	}
	if rows == 0 {
		return errors.New("plan run not found")
	}

	return nil
}
