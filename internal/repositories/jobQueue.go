package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type JobQueue struct {
	db *sql.DB
}

func NewJobQueue(db *sql.DB) *JobQueue {
	return &JobQueue{db: db}
}

func (r *JobQueue) Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	query := `
		INSERT INTO industry_job_queue
			(user_id, character_id, blueprint_type_id, activity, runs,
			 me_level, te_level, system_id, facility_tax, status,
			 product_type_id, estimated_cost, estimated_duration, notes,
			 plan_run_id, plan_step_id, transport_job_id,
			 sort_order, station_name, input_location, output_location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'planned', $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id, user_id, character_id, blueprint_type_id, activity, runs,
		          me_level, te_level, system_id, facility_tax, status, esi_job_id,
		          product_type_id, estimated_cost, estimated_duration, notes,
		          plan_run_id, plan_step_id, transport_job_id,
		          sort_order, station_name, input_location, output_location,
		          created_at, updated_at
	`

	var created models.IndustryJobQueueEntry
	err := r.db.QueryRowContext(ctx, query,
		entry.UserID,
		entry.CharacterID,
		entry.BlueprintTypeID,
		entry.Activity,
		entry.Runs,
		entry.MELevel,
		entry.TELevel,
		entry.SystemID,
		entry.FacilityTax,
		entry.ProductTypeID,
		entry.EstimatedCost,
		entry.EstimatedDuration,
		entry.Notes,
		entry.PlanRunID,
		entry.PlanStepID,
		entry.TransportJobID,
		entry.SortOrder,
		entry.StationName,
		entry.InputLocation,
		entry.OutputLocation,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.CharacterID,
		&created.BlueprintTypeID,
		&created.Activity,
		&created.Runs,
		&created.MELevel,
		&created.TELevel,
		&created.SystemID,
		&created.FacilityTax,
		&created.Status,
		&created.EsiJobID,
		&created.ProductTypeID,
		&created.EstimatedCost,
		&created.EstimatedDuration,
		&created.Notes,
		&created.PlanRunID,
		&created.PlanStepID,
		&created.TransportJobID,
		&created.SortOrder,
		&created.StationName,
		&created.InputLocation,
		&created.OutputLocation,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create job queue entry")
	}

	return &created, nil
}

func (r *JobQueue) GetByUser(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	query := `
		SELECT q.id, q.user_id, q.character_id, q.blueprint_type_id, q.activity, q.runs,
		       q.me_level, q.te_level, q.system_id, q.facility_tax, q.status, q.esi_job_id,
		       q.product_type_id, q.estimated_cost, q.estimated_duration, q.notes,
		       q.plan_run_id, q.plan_step_id, q.transport_job_id,
		       q.sort_order, q.station_name, q.input_location, q.output_location,
		       q.created_at, q.updated_at,
		       COALESCE(bp.type_name, ''),
		       COALESCE(prod.type_name, ''),
		       COALESCE(c.name, installer.name, ''),
		       COALESCE(ss.name, ''),
		       j.end_date,
		       COALESCE(j.source, ''),
		       COALESCE(origin_st.name, ''),
		       COALESCE(dest_st.name, ''),
		       COALESCE(tj.transport_method, ''),
		       COALESCE(tj.fulfillment_type, ''),
		       COALESCE(tj.total_volume_m3, 0),
		       COALESCE(tj.jumps, 0),
		       COALESCE((
		           SELECT string_agg(ait.type_name || ' x' || tji.quantity::text, ', ' ORDER BY ait.type_name)
		           FROM transport_job_items tji
		           JOIN asset_item_types ait ON ait.type_id = tji.type_id
		           WHERE tji.transport_job_id = q.transport_job_id
		       ), '')
		FROM industry_job_queue q
		LEFT JOIN asset_item_types bp ON bp.type_id = q.blueprint_type_id
		LEFT JOIN asset_item_types prod ON prod.type_id = q.product_type_id
		LEFT JOIN characters c ON c.id = q.character_id
		LEFT JOIN solar_systems ss ON ss.solar_system_id = q.system_id
		LEFT JOIN esi_industry_jobs j ON j.job_id = q.esi_job_id
		LEFT JOIN characters installer ON installer.id = j.installer_id
		LEFT JOIN transport_jobs tj ON tj.id = q.transport_job_id
		LEFT JOIN stations origin_st ON origin_st.station_id = tj.origin_station_id
		LEFT JOIN stations dest_st ON dest_st.station_id = tj.destination_station_id
		WHERE q.user_id = $1
		  AND q.status NOT IN ('cancelled', 'completed')
		ORDER BY q.sort_order DESC, q.created_at ASC
	`

	return r.queryEntries(ctx, query, userID)
}

// GetPlannedJobs returns job queue entries with status='planned' for a user.
func (r *JobQueue) GetPlannedJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	query := `
		SELECT q.id, q.user_id, q.character_id, q.blueprint_type_id, q.activity, q.runs,
		       q.me_level, q.te_level, q.system_id, q.facility_tax, q.status, q.esi_job_id,
		       q.product_type_id, q.estimated_cost, q.estimated_duration, q.notes,
		       q.plan_run_id, q.plan_step_id, q.transport_job_id,
		       q.sort_order, q.station_name, q.input_location, q.output_location,
		       q.created_at, q.updated_at,
		       '', '', '', '',
		       CAST(NULL AS timestamptz),
		       '',
		       '', '', '', '', 0, 0,
		       ''
		FROM industry_job_queue q
		WHERE q.user_id = $1
		  AND q.status = 'planned'
		ORDER BY q.sort_order DESC, q.created_at ASC
	`

	return r.queryEntries(ctx, query, userID)
}

func (r *JobQueue) Update(ctx context.Context, id, userID int64, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	query := `
		UPDATE industry_job_queue
		SET character_id = $3,
		    blueprint_type_id = $4,
		    activity = $5,
		    runs = $6,
		    me_level = $7,
		    te_level = $8,
		    system_id = $9,
		    facility_tax = $10,
		    product_type_id = $11,
		    estimated_cost = $12,
		    estimated_duration = $13,
		    notes = $14,
		    updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status = 'planned'
		RETURNING id, user_id, character_id, blueprint_type_id, activity, runs,
		          me_level, te_level, system_id, facility_tax, status, esi_job_id,
		          product_type_id, estimated_cost, estimated_duration, notes,
		          plan_run_id, plan_step_id, transport_job_id,
		          sort_order, station_name, input_location, output_location,
		          created_at, updated_at
	`

	var updated models.IndustryJobQueueEntry
	err := r.db.QueryRowContext(ctx, query,
		id,
		userID,
		entry.CharacterID,
		entry.BlueprintTypeID,
		entry.Activity,
		entry.Runs,
		entry.MELevel,
		entry.TELevel,
		entry.SystemID,
		entry.FacilityTax,
		entry.ProductTypeID,
		entry.EstimatedCost,
		entry.EstimatedDuration,
		entry.Notes,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.CharacterID,
		&updated.BlueprintTypeID,
		&updated.Activity,
		&updated.Runs,
		&updated.MELevel,
		&updated.TELevel,
		&updated.SystemID,
		&updated.FacilityTax,
		&updated.Status,
		&updated.EsiJobID,
		&updated.ProductTypeID,
		&updated.EstimatedCost,
		&updated.EstimatedDuration,
		&updated.Notes,
		&updated.PlanRunID,
		&updated.PlanStepID,
		&updated.TransportJobID,
		&updated.SortOrder,
		&updated.StationName,
		&updated.InputLocation,
		&updated.OutputLocation,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update job queue entry")
	}

	return &updated, nil
}

func (r *JobQueue) Cancel(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE industry_job_queue
		SET status = 'cancelled', updated_at = now()
		WHERE id = $1 AND user_id = $2 AND status IN ('planned', 'active')
	`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to cancel job queue entry")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for cancel")
	}
	if rows == 0 {
		return errors.New("job queue entry not found or not cancellable")
	}

	return nil
}

// LinkToEsiJob links a planned queue entry to an active ESI job.
func (r *JobQueue) LinkToEsiJob(ctx context.Context, queueID, esiJobID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE industry_job_queue
		SET esi_job_id = $2, status = 'active', updated_at = now()
		WHERE id = $1
	`, queueID, esiJobID)
	if err != nil {
		return errors.Wrap(err, "failed to link queue entry to ESI job")
	}
	return nil
}

// CompleteJob marks a queue entry as completed.
func (r *JobQueue) CompleteJob(ctx context.Context, queueID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE industry_job_queue
		SET status = 'completed', updated_at = now()
		WHERE id = $1
	`, queueID)
	if err != nil {
		return errors.Wrap(err, "failed to complete job queue entry")
	}
	return nil
}

// GetLinkedActiveJobs returns queue entries that are linked to ESI jobs (status='active').
func (r *JobQueue) GetLinkedActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	query := `
		SELECT q.id, q.user_id, q.character_id, q.blueprint_type_id, q.activity, q.runs,
		       q.me_level, q.te_level, q.system_id, q.facility_tax, q.status, q.esi_job_id,
		       q.product_type_id, q.estimated_cost, q.estimated_duration, q.notes,
		       q.plan_run_id, q.plan_step_id, q.transport_job_id,
		       q.sort_order, q.station_name, q.input_location, q.output_location,
		       q.created_at, q.updated_at,
		       '', '', '', '',
		       CAST(NULL AS timestamptz),
		       '',
		       '', '', '', '', 0, 0,
		       ''
		FROM industry_job_queue q
		WHERE q.user_id = $1
		  AND q.status = 'active'
		  AND q.esi_job_id IS NOT NULL
		ORDER BY q.sort_order DESC, q.created_at ASC
	`

	return r.queryEntries(ctx, query, userID)
}

// GetSlotUsage returns a nested map of characterID -> activity -> count for
// all planned and active queue entries that have a character assigned.
func (r *JobQueue) GetSlotUsage(ctx context.Context, userID int64) (map[int64]map[string]int, error) {
	query := `
		select character_id, activity, count(*) as slot_count
		from industry_job_queue
		where user_id = $1
		  and character_id is not null
		  and status in ('planned', 'active')
		group by character_id, activity
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query slot usage")
	}
	defer rows.Close()

	result := map[int64]map[string]int{}
	for rows.Next() {
		var characterID int64
		var activity string
		var count int
		if err := rows.Scan(&characterID, &activity, &count); err != nil {
			return nil, errors.Wrap(err, "failed to scan slot usage row")
		}
		if _, ok := result[characterID]; !ok {
			result[characterID] = map[string]int{}
		}
		result[characterID][activity] = count
	}

	return result, nil
}

// ReassignCharacter updates the character_id on a planned queue entry.
// Pass nil for characterID to unassign. Returns an error if the entry is not
// found, belongs to a different user, or is not in 'planned' status.
func (r *JobQueue) ReassignCharacter(ctx context.Context, id, userID int64, characterID *int64) error {
	var result sql.Result
	var err error

	if characterID != nil {
		result, err = r.db.ExecContext(ctx, `
			UPDATE industry_job_queue
			SET character_id = $3, updated_at = now()
			WHERE id = $1 AND user_id = $2 AND status = 'planned'
		`, id, userID, *characterID)
	} else {
		result, err = r.db.ExecContext(ctx, `
			UPDATE industry_job_queue
			SET character_id = NULL, updated_at = now()
			WHERE id = $1 AND user_id = $2 AND status = 'planned'
		`, id, userID)
	}
	if err != nil {
		return errors.Wrap(err, "reassigning character")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected")
	}
	if rows == 0 {
		return errors.New("queue entry not found or not in planned status")
	}

	return nil
}

func (r *JobQueue) queryEntries(ctx context.Context, query string, args ...interface{}) ([]*models.IndustryJobQueueEntry, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query job queue entries")
	}
	defer rows.Close()

	entries := []*models.IndustryJobQueueEntry{}
	for rows.Next() {
		var entry models.IndustryJobQueueEntry
		err = rows.Scan(
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
			&entry.TransportJobID,
			&entry.SortOrder,
			&entry.StationName,
			&entry.InputLocation,
			&entry.OutputLocation,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			// Enriched fields from JOINs
			&entry.BlueprintName,
			&entry.ProductName,
			&entry.CharacterName,
			&entry.SystemName,
			&entry.EsiJobEndDate,
			&entry.EsiJobSource,
			&entry.TransportOriginName,
			&entry.TransportDestName,
			&entry.TransportMethod,
			&entry.TransportFulfillment,
			&entry.TransportVolumeM3,
			&entry.TransportJumps,
			&entry.TransportItemsSummary,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan job queue entry")
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}
