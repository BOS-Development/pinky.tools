package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type HaulingRunPnl struct {
	db *sql.DB
}

func NewHaulingRunPnl(db *sql.DB) *HaulingRunPnl {
	return &HaulingRunPnl{db: db}
}

// UpsertPnlEntry upserts a P&L entry for a (run_id, type_id) pair.
// net_profit_isk is a generated column and must NOT be in the INSERT/UPDATE.
func (r *HaulingRunPnl) UpsertPnlEntry(ctx context.Context, entry *models.HaulingRunPnlEntry) error {
	query := `
		INSERT INTO hauling_run_pnl (run_id, type_id, quantity_sold, avg_sell_price_isk, total_revenue_isk, total_cost_isk, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT ON CONSTRAINT hauling_run_pnl_run_type_unique DO UPDATE SET
			quantity_sold     = EXCLUDED.quantity_sold,
			avg_sell_price_isk = EXCLUDED.avg_sell_price_isk,
			total_revenue_isk = EXCLUDED.total_revenue_isk,
			total_cost_isk    = EXCLUDED.total_cost_isk,
			updated_at        = NOW()
		RETURNING id, created_at, updated_at`

	var id int64
	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		entry.RunID, entry.TypeID, entry.QuantitySold,
		entry.AvgSellPriceISK, entry.TotalRevenueISK, entry.TotalCostISK,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return errors.Wrap(err, "failed to upsert pnl entry")
	}
	entry.ID = id
	entry.CreatedAt = createdAt.Format(time.RFC3339)
	entry.UpdatedAt = updatedAt.Format(time.RFC3339)
	return nil
}

// GetPnlByRunID returns all P&L entries for a run, joining type_name from hauling_run_items.
func (r *HaulingRunPnl) GetPnlByRunID(ctx context.Context, runID int64) ([]*models.HaulingRunPnlEntry, error) {
	query := `
		SELECT p.id, p.run_id, p.type_id,
		       COALESCE(i.type_name, '') as type_name,
		       p.quantity_sold, p.avg_sell_price_isk, p.total_revenue_isk,
		       p.total_cost_isk, p.net_profit_isk, p.created_at, p.updated_at
		FROM hauling_run_pnl p
		LEFT JOIN hauling_run_items i ON i.run_id = p.run_id AND i.type_id = p.type_id
		WHERE p.run_id = $1
		ORDER BY p.created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pnl entries")
	}
	defer rows.Close()

	entries := []*models.HaulingRunPnlEntry{}
	for rows.Next() {
		var e models.HaulingRunPnlEntry
		var avgSell, totalRevenue, totalCost, netProfit sql.NullFloat64
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&e.ID, &e.RunID, &e.TypeID, &e.TypeName,
			&e.QuantitySold, &avgSell, &totalRevenue, &totalCost, &netProfit,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan pnl entry")
		}
		if avgSell.Valid {
			e.AvgSellPriceISK = &avgSell.Float64
		}
		if totalRevenue.Valid {
			e.TotalRevenueISK = &totalRevenue.Float64
		}
		if totalCost.Valid {
			e.TotalCostISK = &totalCost.Float64
		}
		if netProfit.Valid {
			e.NetProfitISK = &netProfit.Float64
		}
		e.CreatedAt = createdAt.Format(time.RFC3339)
		e.UpdatedAt = updatedAt.Format(time.RFC3339)
		entries = append(entries, &e)
	}
	return entries, nil
}

// GetPnlSummaryByRunID returns aggregated P&L for a run.
func (r *HaulingRunPnl) GetPnlSummaryByRunID(ctx context.Context, runID int64) (*models.HaulingRunPnlSummary, error) {
	summaryQuery := `
		SELECT
			COALESCE(SUM(total_revenue_isk), 0) as total_revenue,
			COALESCE(SUM(total_cost_isk), 0) as total_cost,
			COALESCE(SUM(net_profit_isk), 0) as net_profit,
			COUNT(*) as items_sold
		FROM hauling_run_pnl
		WHERE run_id = $1`

	var summary models.HaulingRunPnlSummary
	err := r.db.QueryRowContext(ctx, summaryQuery, runID).Scan(
		&summary.TotalRevenueISK,
		&summary.TotalCostISK,
		&summary.NetProfitISK,
		&summary.ItemsSold,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pnl summary")
	}

	// Calculate margin
	if summary.TotalRevenueISK > 0 {
		summary.MarginPct = summary.NetProfitISK / summary.TotalRevenueISK * 100
	}

	// Calculate items pending: items in hauling_run_items not fully sold
	pendingQuery := `
		SELECT COUNT(*)
		FROM hauling_run_items i
		LEFT JOIN hauling_run_pnl p ON p.run_id = i.run_id AND p.type_id = i.type_id
		WHERE i.run_id = $1
		  AND (p.quantity_sold IS NULL OR p.quantity_sold < i.quantity_acquired)`

	err = r.db.QueryRowContext(ctx, pendingQuery, runID).Scan(&summary.ItemsPending)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pending items count")
	}

	return &summary, nil
}
