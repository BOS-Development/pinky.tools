package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type SalesAnalytics struct {
	db *sql.DB
}

func NewSalesAnalytics(db *sql.DB) *SalesAnalytics {
	return &SalesAnalytics{db: db}
}

// GetSalesMetrics returns aggregated sales metrics for a given time period
func (r *SalesAnalytics) GetSalesMetrics(ctx context.Context, sellerUserID int64, periodDays int) (*models.SalesMetrics, error) {
	var startDate time.Time
	if periodDays > 0 {
		startDate = time.Now().AddDate(0, 0, -periodDays)
	}

	metrics := &models.SalesMetrics{}

	// Get overall metrics
	query := `
		SELECT
			COALESCE(SUM(total_price), 0) as total_revenue,
			COUNT(*) as total_transactions,
			COALESCE(SUM(quantity_purchased), 0) as total_quantity_sold,
			COUNT(DISTINCT type_id) as unique_item_types,
			COUNT(DISTINCT buyer_user_id) as unique_buyers
		FROM purchase_transactions
		WHERE seller_user_id = $1
			AND status = 'completed'
	`

	args := []interface{}{sellerUserID}
	if periodDays > 0 {
		query += " AND purchased_at >= $2"
		args = append(args, startDate)
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.TotalRevenue,
		&metrics.TotalTransactions,
		&metrics.TotalQuantitySold,
		&metrics.UniqueItemTypes,
		&metrics.UniqueBuyers,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sales metrics")
	}

	// Get time series data (daily aggregates)
	timeSeriesQuery := `
		SELECT
			DATE(purchased_at) as date,
			COALESCE(SUM(total_price), 0) as revenue,
			COUNT(*) as transactions,
			COALESCE(SUM(quantity_purchased), 0) as quantity_sold
		FROM purchase_transactions
		WHERE seller_user_id = $1
			AND status = 'completed'
	`

	timeSeriesArgs := []interface{}{sellerUserID}
	if periodDays > 0 {
		timeSeriesQuery += " AND purchased_at >= $2"
		timeSeriesArgs = append(timeSeriesArgs, startDate)
	}

	timeSeriesQuery += `
		GROUP BY DATE(purchased_at)
		ORDER BY DATE(purchased_at) ASC
	`

	rows, err := r.db.QueryContext(ctx, timeSeriesQuery, timeSeriesArgs...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get time series data")
	}
	defer rows.Close()

	metrics.TimeSeriesData = []models.TimeSeriesData{}
	for rows.Next() {
		var ts models.TimeSeriesData
		var date time.Time
		err = rows.Scan(&date, &ts.Revenue, &ts.Transactions, &ts.QuantitySold)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan time series data")
		}
		ts.Date = date.Format("2006-01-02")
		metrics.TimeSeriesData = append(metrics.TimeSeriesData, ts)
	}

	// Get top items
	topItems, err := r.GetTopItems(ctx, sellerUserID, periodDays, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get top items")
	}
	metrics.TopItems = topItems

	return metrics, nil
}

// GetTopItems returns the top selling items by revenue
func (r *SalesAnalytics) GetTopItems(ctx context.Context, sellerUserID int64, periodDays int, limit int) ([]models.ItemSalesData, error) {
	var startDate time.Time
	if periodDays > 0 {
		startDate = time.Now().AddDate(0, 0, -periodDays)
	}

	query := `
		SELECT
			pt.type_id,
			t.type_name,
			COALESCE(SUM(pt.quantity_purchased), 0) as quantity_sold,
			COALESCE(SUM(pt.total_price), 0) as revenue,
			COUNT(*) as transaction_count,
			COALESCE(AVG(pt.price_per_unit), 0) as avg_price_per_unit
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		WHERE pt.seller_user_id = $1
			AND pt.status = 'completed'
	`

	args := []interface{}{sellerUserID}
	if periodDays > 0 {
		query += " AND pt.purchased_at >= $2"
		args = append(args, startDate)
	}

	query += `
		GROUP BY pt.type_id, t.type_name
		ORDER BY revenue DESC
		LIMIT $` + getNextParamNum(args)

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get top items")
	}
	defer rows.Close()

	items := []models.ItemSalesData{}
	for rows.Next() {
		var item models.ItemSalesData
		err = rows.Scan(
			&item.TypeID,
			&item.TypeName,
			&item.QuantitySold,
			&item.Revenue,
			&item.TransactionCount,
			&item.AveragePricePerUnit,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item sales data")
		}
		items = append(items, item)
	}

	return items, nil
}

// GetBuyerAnalytics returns analytics about buyers
func (r *SalesAnalytics) GetBuyerAnalytics(ctx context.Context, sellerUserID int64, periodDays int, limit int) ([]models.BuyerAnalytics, error) {
	var startDate time.Time
	if periodDays > 0 {
		startDate = time.Now().AddDate(0, 0, -periodDays)
	}

	query := `
		SELECT
			pt.buyer_user_id,
			COALESCE(MAX(c.name), CONCAT('User ', pt.buyer_user_id)) as buyer_name,
			COALESCE(SUM(pt.total_price), 0) as total_spent,
			COUNT(*) as total_purchases,
			COALESCE(SUM(pt.quantity_purchased), 0) as total_quantity,
			MIN(pt.purchased_at) as first_purchase_date,
			MAX(pt.purchased_at) as last_purchase_date,
			CASE WHEN COUNT(*) > 1 THEN true ELSE false END as repeat_customer
		FROM purchase_transactions pt
		LEFT JOIN characters c ON pt.buyer_user_id = c.user_id
		WHERE pt.seller_user_id = $1
			AND pt.status = 'completed'
	`

	args := []interface{}{sellerUserID}
	if periodDays > 0 {
		query += " AND pt.purchased_at >= $2"
		args = append(args, startDate)
	}

	query += `
		GROUP BY pt.buyer_user_id
		ORDER BY total_spent DESC
		LIMIT $` + getNextParamNum(args)

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get buyer analytics")
	}
	defer rows.Close()

	buyers := []models.BuyerAnalytics{}
	for rows.Next() {
		var buyer models.BuyerAnalytics
		err = rows.Scan(
			&buyer.BuyerUserID,
			&buyer.BuyerName,
			&buyer.TotalSpent,
			&buyer.TotalPurchases,
			&buyer.TotalQuantity,
			&buyer.FirstPurchaseDate,
			&buyer.LastPurchaseDate,
			&buyer.RepeatCustomer,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan buyer analytics")
		}
		buyers = append(buyers, buyer)
	}

	return buyers, nil
}

// GetItemSalesHistory returns sales history for a specific item type
func (r *SalesAnalytics) GetItemSalesHistory(ctx context.Context, sellerUserID int64, typeID int64, periodDays int) (*models.ItemSalesData, error) {
	var startDate time.Time
	if periodDays > 0 {
		startDate = time.Now().AddDate(0, 0, -periodDays)
	}

	query := `
		SELECT
			pt.type_id,
			t.type_name,
			COALESCE(SUM(pt.quantity_purchased), 0) as quantity_sold,
			COALESCE(SUM(pt.total_price), 0) as revenue,
			COUNT(*) as transaction_count,
			COALESCE(AVG(pt.price_per_unit), 0) as avg_price_per_unit
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		WHERE pt.seller_user_id = $1
			AND pt.type_id = $2
			AND pt.status = 'completed'
	`

	args := []interface{}{sellerUserID, typeID}
	if periodDays > 0 {
		query += " AND pt.purchased_at >= $3"
		args = append(args, startDate)
	}

	query += " GROUP BY pt.type_id, t.type_name"

	var item models.ItemSalesData
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&item.TypeID,
		&item.TypeName,
		&item.QuantitySold,
		&item.Revenue,
		&item.TransactionCount,
		&item.AveragePricePerUnit,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("no sales data found for this item")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item sales history")
	}

	return &item, nil
}

// Helper function to get next parameter number for SQL query
func getNextParamNum(args []interface{}) string {
	switch len(args) {
	case 0:
		return "1"
	case 1:
		return "2"
	case 2:
		return "3"
	default:
		return "4"
	}
}
