package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type PurchaseTransactions struct {
	db *sql.DB
}

func NewPurchaseTransactions(db *sql.DB) *PurchaseTransactions {
	return &PurchaseTransactions{db: db}
}

// Create records a new purchase transaction (within transaction)
func (r *PurchaseTransactions) Create(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error {
	query := `
		INSERT INTO purchase_transactions
		(for_sale_item_id, buyer_user_id, seller_user_id, type_id, quantity_purchased,
		 price_per_unit, total_price, status, transaction_notes, buy_order_id, is_auto_fulfilled, purchased_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, purchased_at
	`

	err := tx.QueryRowContext(ctx, query,
		purchase.ForSaleItemID,
		purchase.BuyerUserID,
		purchase.SellerUserID,
		purchase.TypeID,
		purchase.QuantityPurchased,
		purchase.PricePerUnit,
		purchase.TotalPrice,
		purchase.Status,
		purchase.TransactionNotes,
		purchase.BuyOrderID,
		purchase.IsAutoFulfilled,
	).Scan(&purchase.ID, &purchase.PurchasedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create purchase transaction")
	}

	return nil
}

// CreateAutoFulfill records an auto-fulfilled purchase transaction (within transaction)
func (r *PurchaseTransactions) CreateAutoFulfill(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error {
	purchase.IsAutoFulfilled = true
	return r.Create(ctx, tx, purchase)
}

// UpdateContractKeys updates contract keys for multiple purchase IDs
func (r *PurchaseTransactions) UpdateContractKeys(ctx context.Context, purchaseIDs []int64, contractKey string) error {
	if len(purchaseIDs) == 0 {
		return nil
	}

	// Convert slice to PostgreSQL array format
	query := `UPDATE purchase_transactions SET contract_key = $1 WHERE id = ANY($2)`

	result, err := r.db.ExecContext(ctx, query, contractKey, pq.Array(purchaseIDs))
	if err != nil {
		return errors.Wrap(err, "failed to update contract keys")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("no purchase transactions found to update")
	}

	return nil
}

// GetByBuyer returns purchase history for buyer
func (r *PurchaseTransactions) GetByBuyer(ctx context.Context, buyerUserID int64) ([]*models.PurchaseTransaction, error) {
	query := `
		SELECT
			pt.id,
			pt.for_sale_item_id,
			pt.buyer_user_id,
			pt.seller_user_id,
			pt.type_id,
			t.type_name,
			pt.quantity_purchased,
			pt.price_per_unit,
			pt.total_price,
			pt.status,
			pt.contract_key,
			pt.transaction_notes,
			pt.buy_order_id,
			pt.is_auto_fulfilled,
			pt.purchased_at
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		WHERE pt.buyer_user_id = $1
		ORDER BY pt.purchased_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, buyerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query buyer purchase history")
	}
	defer rows.Close()

	transactions := []*models.PurchaseTransaction{}
	for rows.Next() {
		var tx models.PurchaseTransaction
		err = rows.Scan(
			&tx.ID,
			&tx.ForSaleItemID,
			&tx.BuyerUserID,
			&tx.SellerUserID,
			&tx.TypeID,
			&tx.TypeName,
			&tx.QuantityPurchased,
			&tx.PricePerUnit,
			&tx.TotalPrice,
			&tx.Status,
			&tx.ContractKey,
			&tx.TransactionNotes,
			&tx.BuyOrderID,
			&tx.IsAutoFulfilled,
			&tx.PurchasedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan purchase transaction")
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// GetBySeller returns sales history for seller
func (r *PurchaseTransactions) GetBySeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error) {
	query := `
		SELECT
			pt.id,
			pt.for_sale_item_id,
			pt.buyer_user_id,
			pt.seller_user_id,
			pt.type_id,
			t.type_name,
			pt.quantity_purchased,
			pt.price_per_unit,
			pt.total_price,
			pt.status,
			pt.contract_key,
			pt.transaction_notes,
			pt.buy_order_id,
			pt.is_auto_fulfilled,
			pt.purchased_at
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		WHERE pt.seller_user_id = $1
		ORDER BY pt.purchased_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sellerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query seller sales history")
	}
	defer rows.Close()

	transactions := []*models.PurchaseTransaction{}
	for rows.Next() {
		var tx models.PurchaseTransaction
		err = rows.Scan(
			&tx.ID,
			&tx.ForSaleItemID,
			&tx.BuyerUserID,
			&tx.SellerUserID,
			&tx.TypeID,
			&tx.TypeName,
			&tx.QuantityPurchased,
			&tx.PricePerUnit,
			&tx.TotalPrice,
			&tx.Status,
			&tx.ContractKey,
			&tx.TransactionNotes,
			&tx.BuyOrderID,
			&tx.IsAutoFulfilled,
			&tx.PurchasedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan purchase transaction")
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// GetPendingForSeller returns pending purchase requests for seller
func (r *PurchaseTransactions) GetPendingForSeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error) {
	query := `
		SELECT
			pt.id,
			pt.for_sale_item_id,
			pt.buyer_user_id,
			COALESCE(buyer_user.name, CONCAT('User ', pt.buyer_user_id)) AS buyer_name,
			pt.seller_user_id,
			pt.type_id,
			t.type_name,
			fsi.location_id,
			COALESCE(s.name, st.name, 'Unknown Location') AS location_name,
			pt.quantity_purchased,
			pt.price_per_unit,
			pt.total_price,
			pt.status,
			pt.contract_key,
			pt.transaction_notes,
			pt.buy_order_id,
			pt.is_auto_fulfilled,
			pt.purchased_at
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		JOIN for_sale_items fsi ON pt.for_sale_item_id = fsi.id
		LEFT JOIN users buyer_user ON pt.buyer_user_id = buyer_user.id
		LEFT JOIN solar_systems s ON fsi.location_id = s.solar_system_id
		LEFT JOIN stations st ON fsi.location_id = st.station_id
		WHERE pt.seller_user_id = $1 AND pt.status = 'pending'
		ORDER BY fsi.location_id, COALESCE(buyer_user.name, CONCAT('User ', pt.buyer_user_id)), pt.purchased_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sellerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pending sales")
	}
	defer rows.Close()

	transactions := []*models.PurchaseTransaction{}
	for rows.Next() {
		var tx models.PurchaseTransaction
		err = rows.Scan(
			&tx.ID,
			&tx.ForSaleItemID,
			&tx.BuyerUserID,
			&tx.BuyerName,
			&tx.SellerUserID,
			&tx.TypeID,
			&tx.TypeName,
			&tx.LocationID,
			&tx.LocationName,
			&tx.QuantityPurchased,
			&tx.PricePerUnit,
			&tx.TotalPrice,
			&tx.Status,
			&tx.ContractKey,
			&tx.TransactionNotes,
			&tx.BuyOrderID,
			&tx.IsAutoFulfilled,
			&tx.PurchasedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pending sale")
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// GetByID returns a specific purchase transaction by ID
func (r *PurchaseTransactions) GetByID(ctx context.Context, purchaseID int64) (*models.PurchaseTransaction, error) {
	query := `
		SELECT
			pt.id,
			pt.for_sale_item_id,
			pt.buyer_user_id,
			pt.seller_user_id,
			pt.type_id,
			t.type_name,
			pt.quantity_purchased,
			pt.price_per_unit,
			pt.total_price,
			pt.status,
			pt.contract_key,
			pt.transaction_notes,
			pt.buy_order_id,
			pt.is_auto_fulfilled,
			pt.purchased_at
		FROM purchase_transactions pt
		JOIN asset_item_types t ON pt.type_id = t.type_id
		WHERE pt.id = $1
	`

	var tx models.PurchaseTransaction
	err := r.db.QueryRowContext(ctx, query, purchaseID).Scan(
		&tx.ID,
		&tx.ForSaleItemID,
		&tx.BuyerUserID,
		&tx.SellerUserID,
		&tx.TypeID,
		&tx.TypeName,
		&tx.QuantityPurchased,
		&tx.PricePerUnit,
		&tx.TotalPrice,
		&tx.Status,
		&tx.ContractKey,
		&tx.TransactionNotes,
		&tx.BuyOrderID,
		&tx.IsAutoFulfilled,
		&tx.PurchasedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("purchase transaction not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get purchase transaction")
	}

	return &tx, nil
}

// UpdateStatus updates the status of a purchase transaction
func (r *PurchaseTransactions) UpdateStatus(ctx context.Context, purchaseID int64, newStatus string) error {
	query := `
		UPDATE purchase_transactions
		SET status = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, purchaseID, newStatus)
	if err != nil {
		return errors.Wrap(err, "failed to update purchase status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("purchase transaction not found")
	}

	return nil
}
