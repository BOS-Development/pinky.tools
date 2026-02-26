package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type TransportJobs struct {
	db *sql.DB
}

func NewTransportJobs(db *sql.DB) *TransportJobs {
	return &TransportJobs{db: db}
}

func (r *TransportJobs) GetByUser(ctx context.Context, userID int64) ([]*models.TransportJob, error) {
	query := `
		select j.id, j.user_id, j.origin_station_id, j.destination_station_id,
		       j.origin_system_id, j.destination_system_id,
		       j.transport_method, j.route_preference, j.status,
		       j.total_volume_m3, j.total_collateral, j.estimated_cost,
		       j.jumps, j.distance_ly, j.jf_route_id,
		       j.fulfillment_type, j.transport_profile_id,
		       j.plan_run_id, j.plan_step_id, j.queue_entry_id,
		       j.notes, j.created_at, j.updated_at,
		       COALESCE(os.name, ''), COALESCE(ds.name, ''),
		       COALESCE(oss.name, ''), COALESCE(dss.name, ''),
		       COALESCE(tp.name, ''), COALESCE(jr.name, '')
		from transport_jobs j
		left join stations os on os.station_id = j.origin_station_id
		left join stations ds on ds.station_id = j.destination_station_id
		left join solar_systems oss on oss.solar_system_id = j.origin_system_id
		left join solar_systems dss on dss.solar_system_id = j.destination_system_id
		left join transport_profiles tp on tp.id = j.transport_profile_id
		left join jf_routes jr on jr.id = j.jf_route_id
		where j.user_id = $1
		order by j.created_at desc
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query transport jobs")
	}
	defer rows.Close()

	jobs := []*models.TransportJob{}
	for rows.Next() {
		var j models.TransportJob
		if err := rows.Scan(
			&j.ID, &j.UserID, &j.OriginStationID, &j.DestinationStationID,
			&j.OriginSystemID, &j.DestinationSystemID,
			&j.TransportMethod, &j.RoutePreference, &j.Status,
			&j.TotalVolumeM3, &j.TotalCollateral, &j.EstimatedCost,
			&j.Jumps, &j.DistanceLY, &j.JFRouteID,
			&j.FulfillmentType, &j.TransportProfileID,
			&j.PlanRunID, &j.PlanStepID, &j.QueueEntryID,
			&j.Notes, &j.CreatedAt, &j.UpdatedAt,
			&j.OriginStationName, &j.DestinationStationName,
			&j.OriginSystemName, &j.DestinationSystemName,
			&j.ProfileName, &j.JFRouteName,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan transport job")
		}
		j.Items = []*models.TransportJobItem{}
		jobs = append(jobs, &j)
	}

	// Batch load items for all jobs
	if len(jobs) > 0 {
		jobIDs := make([]int64, len(jobs))
		for i, j := range jobs {
			jobIDs[i] = j.ID
		}

		itemQuery := `
			select i.id, i.transport_job_id, i.type_id, i.quantity, i.volume_m3, i.estimated_value,
			       COALESCE(t.type_name, '')
			from transport_job_items i
			left join asset_item_types t on t.type_id = i.type_id
			where i.transport_job_id = ANY($1)
			order by i.transport_job_id, i.id asc
		`

		itemRows, err := r.db.QueryContext(ctx, itemQuery, pq.Array(jobIDs))
		if err != nil {
			return nil, errors.Wrap(err, "failed to batch query transport job items")
		}
		defer itemRows.Close()

		itemMap := map[int64][]*models.TransportJobItem{}
		for itemRows.Next() {
			var item models.TransportJobItem
			if err := itemRows.Scan(
				&item.ID, &item.TransportJobID, &item.TypeID,
				&item.Quantity, &item.VolumeM3, &item.EstimatedValue,
				&item.TypeName,
			); err != nil {
				return nil, errors.Wrap(err, "failed to scan transport job item")
			}
			itemMap[item.TransportJobID] = append(itemMap[item.TransportJobID], &item)
		}

		for _, job := range jobs {
			if items, ok := itemMap[job.ID]; ok {
				job.Items = items
			}
		}
	}

	return jobs, nil
}

func (r *TransportJobs) GetByID(ctx context.Context, id, userID int64) (*models.TransportJob, error) {
	query := `
		select j.id, j.user_id, j.origin_station_id, j.destination_station_id,
		       j.origin_system_id, j.destination_system_id,
		       j.transport_method, j.route_preference, j.status,
		       j.total_volume_m3, j.total_collateral, j.estimated_cost,
		       j.jumps, j.distance_ly, j.jf_route_id,
		       j.fulfillment_type, j.transport_profile_id,
		       j.plan_run_id, j.plan_step_id, j.queue_entry_id,
		       j.notes, j.created_at, j.updated_at,
		       COALESCE(os.name, ''), COALESCE(ds.name, ''),
		       COALESCE(oss.name, ''), COALESCE(dss.name, ''),
		       COALESCE(tp.name, ''), COALESCE(jr.name, '')
		from transport_jobs j
		left join stations os on os.station_id = j.origin_station_id
		left join stations ds on ds.station_id = j.destination_station_id
		left join solar_systems oss on oss.solar_system_id = j.origin_system_id
		left join solar_systems dss on dss.solar_system_id = j.destination_system_id
		left join transport_profiles tp on tp.id = j.transport_profile_id
		left join jf_routes jr on jr.id = j.jf_route_id
		where j.id = $1 and j.user_id = $2
	`

	var j models.TransportJob
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&j.ID, &j.UserID, &j.OriginStationID, &j.DestinationStationID,
		&j.OriginSystemID, &j.DestinationSystemID,
		&j.TransportMethod, &j.RoutePreference, &j.Status,
		&j.TotalVolumeM3, &j.TotalCollateral, &j.EstimatedCost,
		&j.Jumps, &j.DistanceLY, &j.JFRouteID,
		&j.FulfillmentType, &j.TransportProfileID,
		&j.PlanRunID, &j.PlanStepID, &j.QueueEntryID,
		&j.Notes, &j.CreatedAt, &j.UpdatedAt,
		&j.OriginStationName, &j.DestinationStationName,
		&j.OriginSystemName, &j.DestinationSystemName,
		&j.ProfileName, &j.JFRouteName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transport job by ID")
	}

	items, err := r.getItems(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	j.Items = items

	return &j, nil
}

func (r *TransportJobs) Create(ctx context.Context, job *models.TransportJob) (*models.TransportJob, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		insert into transport_jobs
			(user_id, origin_station_id, destination_station_id,
			 origin_system_id, destination_system_id,
			 transport_method, route_preference, status,
			 total_volume_m3, total_collateral, estimated_cost,
			 jumps, distance_ly, jf_route_id,
			 fulfillment_type, transport_profile_id,
			 plan_run_id, plan_step_id, notes)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		returning id, user_id, origin_station_id, destination_station_id,
		          origin_system_id, destination_system_id,
		          transport_method, route_preference, status,
		          total_volume_m3, total_collateral, estimated_cost,
		          jumps, distance_ly, jf_route_id,
		          fulfillment_type, transport_profile_id,
		          plan_run_id, plan_step_id, queue_entry_id,
		          notes, created_at, updated_at
	`

	var created models.TransportJob
	err = tx.QueryRowContext(ctx, query,
		job.UserID, job.OriginStationID, job.DestinationStationID,
		job.OriginSystemID, job.DestinationSystemID,
		job.TransportMethod, job.RoutePreference, "planned",
		job.TotalVolumeM3, job.TotalCollateral, job.EstimatedCost,
		job.Jumps, job.DistanceLY, job.JFRouteID,
		job.FulfillmentType, job.TransportProfileID,
		job.PlanRunID, job.PlanStepID, job.Notes,
	).Scan(
		&created.ID, &created.UserID, &created.OriginStationID, &created.DestinationStationID,
		&created.OriginSystemID, &created.DestinationSystemID,
		&created.TransportMethod, &created.RoutePreference, &created.Status,
		&created.TotalVolumeM3, &created.TotalCollateral, &created.EstimatedCost,
		&created.Jumps, &created.DistanceLY, &created.JFRouteID,
		&created.FulfillmentType, &created.TransportProfileID,
		&created.PlanRunID, &created.PlanStepID, &created.QueueEntryID,
		&created.Notes, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport job")
	}

	// Insert items
	created.Items = []*models.TransportJobItem{}
	if job.Items != nil {
		for _, item := range job.Items {
			var createdItem models.TransportJobItem
			err = tx.QueryRowContext(ctx, `
				insert into transport_job_items (transport_job_id, type_id, quantity, volume_m3, estimated_value)
				values ($1, $2, $3, $4, $5)
				returning id, transport_job_id, type_id, quantity, volume_m3, estimated_value
			`, created.ID, item.TypeID, item.Quantity, item.VolumeM3, item.EstimatedValue,
			).Scan(
				&createdItem.ID, &createdItem.TransportJobID,
				&createdItem.TypeID, &createdItem.Quantity,
				&createdItem.VolumeM3, &createdItem.EstimatedValue,
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create transport job item")
			}
			created.Items = append(created.Items, &createdItem)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transport job")
	}

	return &created, nil
}

func (r *TransportJobs) UpdateStatus(ctx context.Context, id, userID int64, status string) error {
	result, err := r.db.ExecContext(ctx, `
		update transport_jobs
		set status = $3, updated_at = now()
		where id = $1 and user_id = $2
	`, id, userID, status)
	if err != nil {
		return errors.Wrap(err, "failed to update transport job status")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rows == 0 {
		return errors.New("transport job not found")
	}

	return nil
}

func (r *TransportJobs) SetQueueEntryID(ctx context.Context, id int64, queueEntryID int64) error {
	_, err := r.db.ExecContext(ctx, `
		update transport_jobs set queue_entry_id = $2, updated_at = now() where id = $1
	`, id, queueEntryID)
	if err != nil {
		return errors.Wrap(err, "failed to set queue entry ID on transport job")
	}
	return nil
}

func (r *TransportJobs) Cancel(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		update transport_jobs
		set status = 'cancelled', updated_at = now()
		where id = $1 and user_id = $2 and status in ('planned', 'in_transit')
	`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to cancel transport job")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for cancel")
	}
	if rows == 0 {
		return errors.New("transport job not found or not cancellable")
	}

	return nil
}

func (r *TransportJobs) getItems(ctx context.Context, jobID int64) ([]*models.TransportJobItem, error) {
	query := `
		select i.id, i.transport_job_id, i.type_id, i.quantity, i.volume_m3, i.estimated_value,
		       COALESCE(t.type_name, '')
		from transport_job_items i
		left join asset_item_types t on t.type_id = i.type_id
		where i.transport_job_id = $1
		order by i.id asc
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query transport job items")
	}
	defer rows.Close()

	items := []*models.TransportJobItem{}
	for rows.Next() {
		var item models.TransportJobItem
		if err := rows.Scan(
			&item.ID, &item.TransportJobID, &item.TypeID,
			&item.Quantity, &item.VolumeM3, &item.EstimatedValue,
			&item.TypeName,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan transport job item")
		}
		items = append(items, &item)
	}

	return items, nil
}
