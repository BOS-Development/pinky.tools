package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingAnalytics struct {
	db *sql.DB
}

func NewHaulingAnalytics(db *sql.DB) *HaulingAnalytics {
	return &HaulingAnalytics{db: db}
}

// GetRouteAnalytics returns per-route aggregated P&L stats for completed runs.
func (r *HaulingAnalytics) GetRouteAnalytics(ctx context.Context, userID int64) ([]*models.HaulingRouteAnalytics, error) {
	query := `
		SELECT
			hr.from_region_id,
			hr.to_region_id,
			COUNT(DISTINCT hr.id) AS total_runs,
			COALESCE(SUM(pnl_agg.run_profit), 0) AS total_profit,
			COALESCE(AVG(pnl_agg.run_profit), 0) AS avg_profit,
			COALESCE(AVG(pnl_agg.run_margin_pct), 0) AS avg_margin_pct,
			COALESCE(AVG(pnl_agg.isk_per_m3), 0) AS avg_isk_per_m3,
			COALESCE(MAX(pnl_agg.run_profit), 0) AS best_run_profit,
			COALESCE(MIN(pnl_agg.run_profit), 0) AS worst_run_profit
		FROM hauling_runs hr
		JOIN (
			SELECT
				p.run_id,
				SUM(p.net_profit_isk) AS run_profit,
				CASE WHEN SUM(p.total_revenue_isk) > 0
				     THEN SUM(p.net_profit_isk) / SUM(p.total_revenue_isk) * 100
				     ELSE 0 END AS run_margin_pct,
				CASE WHEN hr2.max_volume_m3 > 0
				     THEN SUM(p.net_profit_isk) / hr2.max_volume_m3
				     ELSE 0 END AS isk_per_m3
			FROM hauling_run_pnl p
			JOIN hauling_runs hr2 ON hr2.id = p.run_id
			GROUP BY p.run_id, hr2.max_volume_m3
		) pnl_agg ON pnl_agg.run_id = hr.id
		WHERE hr.user_id = $1 AND hr.status = 'COMPLETE'
		GROUP BY hr.from_region_id, hr.to_region_id
		ORDER BY total_profit DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query route analytics")
	}
	defer rows.Close()

	results := []*models.HaulingRouteAnalytics{}
	for rows.Next() {
		var row models.HaulingRouteAnalytics
		if err := rows.Scan(
			&row.FromRegionID, &row.ToRegionID,
			&row.TotalRuns, &row.TotalProfitISK, &row.AvgProfitISK,
			&row.AvgMarginPct, &row.AvgIskPerM3,
			&row.BestRunProfitISK, &row.WorstRunProfitISK,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan route analytics row")
		}
		results = append(results, &row)
	}
	return results, nil
}

// GetItemAnalytics returns per-item-type aggregated P&L stats for completed runs.
func (r *HaulingAnalytics) GetItemAnalytics(ctx context.Context, userID int64) ([]*models.HaulingItemAnalytics, error) {
	query := `
		SELECT
			p.type_id,
			COALESCE(hri.type_name, '') AS type_name,
			COUNT(DISTINCT p.run_id) AS total_runs,
			SUM(p.quantity_sold) AS total_qty_sold,
			SUM(p.net_profit_isk) AS total_profit,
			CASE WHEN SUM(p.total_revenue_isk) > 0
			     THEN SUM(p.net_profit_isk) / SUM(p.total_revenue_isk) * 100
			     ELSE 0 END AS avg_margin_pct
		FROM hauling_run_pnl p
		JOIN hauling_runs hr ON hr.id = p.run_id AND hr.user_id = $1 AND hr.status = 'COMPLETE'
		LEFT JOIN hauling_run_items hri ON hri.run_id = p.run_id AND hri.type_id = p.type_id
		GROUP BY p.type_id, hri.type_name
		ORDER BY total_profit DESC
		LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query item analytics")
	}
	defer rows.Close()

	results := []*models.HaulingItemAnalytics{}
	for rows.Next() {
		var row models.HaulingItemAnalytics
		if err := rows.Scan(
			&row.TypeID, &row.TypeName,
			&row.TotalRuns, &row.TotalQtySold,
			&row.TotalProfitISK, &row.AvgMarginPct,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan item analytics row")
		}
		results = append(results, &row)
	}
	return results, nil
}

// GetProfitTimeSeries returns daily profit aggregated by date+route for completed runs.
func (r *HaulingAnalytics) GetProfitTimeSeries(ctx context.Context, userID int64) ([]*models.HaulingProfitDataPoint, error) {
	query := `
		SELECT
			DATE(hr.completed_at) AS date,
			hr.from_region_id,
			hr.to_region_id,
			COALESCE(SUM(p.net_profit_isk), 0) AS profit,
			COUNT(DISTINCT hr.id) AS run_count
		FROM hauling_runs hr
		LEFT JOIN hauling_run_pnl p ON p.run_id = hr.id
		WHERE hr.user_id = $1 AND hr.status = 'COMPLETE' AND hr.completed_at IS NOT NULL
		GROUP BY DATE(hr.completed_at), hr.from_region_id, hr.to_region_id
		ORDER BY date DESC
		LIMIT 90`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query profit time series")
	}
	defer rows.Close()

	results := []*models.HaulingProfitDataPoint{}
	for rows.Next() {
		var row models.HaulingProfitDataPoint
		var date time.Time
		if err := rows.Scan(
			&date, &row.FromRegionID, &row.ToRegionID,
			&row.ProfitISK, &row.RunCount,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan profit time series row")
		}
		row.Date = date.Format("2006-01-02")
		results = append(results, &row)
	}
	return results, nil
}

// GetRunDurationSummary returns summary stats on how long runs take to complete.
func (r *HaulingAnalytics) GetRunDurationSummary(ctx context.Context, userID int64) (*models.HaulingRunDurationSummary, error) {
	query := `
		SELECT
			COUNT(*) AS total_completed_runs,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0), 0) AS avg_duration_days,
			COALESCE(MIN(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0), 0) AS min_duration_days,
			COALESCE(MAX(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0), 0) AS max_duration_days,
			COALESCE(SUM(pnl_total), 0) AS total_profit
		FROM (
			SELECT hr.completed_at, hr.created_at,
			       COALESCE((SELECT SUM(net_profit_isk) FROM hauling_run_pnl WHERE run_id = hr.id), 0) AS pnl_total
			FROM hauling_runs hr
			WHERE hr.user_id = $1 AND hr.status = 'COMPLETE' AND hr.completed_at IS NOT NULL
		) sub`

	var summary models.HaulingRunDurationSummary
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&summary.TotalCompletedRuns,
		&summary.AvgDurationDays,
		&summary.MinDurationDays,
		&summary.MaxDurationDays,
		&summary.TotalProfitISK,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query run duration summary")
	}
	return &summary, nil
}

// GetCompletedRuns returns paginated COMPLETE and CANCELLED runs for the history view.
// Returns the runs and total count.
func (r *HaulingAnalytics) GetCompletedRuns(ctx context.Context, userID int64, limit, offset int) ([]*models.HaulingRun, int64, error) {
	countQuery := `
		SELECT COUNT(*) FROM hauling_runs
		WHERE user_id = $1 AND status IN ('COMPLETE', 'CANCELLED')`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Wrap(err, "failed to count completed runs")
	}

	query := `
		SELECT id, user_id, name, status, from_region_id, from_system_id, to_region_id,
		       max_volume_m3, haul_threshold_isk, notify_tier2, notify_tier3, daily_digest,
		       notes, completed_at, created_at, updated_at
		FROM hauling_runs
		WHERE user_id = $1 AND status IN ('COMPLETE', 'CANCELLED')
		ORDER BY completed_at DESC NULLS LAST, updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to query completed runs")
	}
	defer rows.Close()

	runs := []*models.HaulingRun{}
	for rows.Next() {
		var run models.HaulingRun
		var fromSystemID sql.NullInt64
		var maxVolume, haulThreshold sql.NullFloat64
		var notes sql.NullString
		var completedAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&run.ID, &run.UserID, &run.Name, &run.Status, &run.FromRegionID, &fromSystemID, &run.ToRegionID,
			&maxVolume, &haulThreshold, &run.NotifyTier2, &run.NotifyTier3, &run.DailyDigest,
			&notes, &completedAt, &createdAt, &updatedAt,
		); err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan completed run")
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
		if completedAt.Valid {
			s := completedAt.Time.Format(time.RFC3339)
			run.CompletedAt = &s
		}
		run.CreatedAt = createdAt.Format(time.RFC3339)
		run.UpdatedAt = updatedAt.Format(time.RFC3339)
		run.Items = []*models.HaulingRunItem{}
		runs = append(runs, &run)
	}
	return runs, total, nil
}
