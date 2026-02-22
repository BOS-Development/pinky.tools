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

// GetContractCreatedWithKeys returns all purchases in contract_created status that have a contract_key set.
func (r *PurchaseTransactions) GetContractCreatedWithKeys(ctx context.Context) ([]*models.PurchaseTransaction, error) {
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
		WHERE pt.status = 'contract_created' AND pt.contract_key IS NOT NULL
		ORDER BY pt.buyer_user_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query contract_created purchases")
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
			return nil, errors.Wrap(err, "failed to scan contract_created purchase")
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// CompleteWithContractID marks a purchase as completed and records the EVE contract ID.
func (r *PurchaseTransactions) CompleteWithContractID(ctx context.Context, purchaseID int64, eveContractID int64) error {
	query := `
		UPDATE purchase_transactions
		SET status = 'completed', contract_key = contract_key || ' [EVE:' || $2::text || ']'
		WHERE id = $1 AND status = 'contract_created'
	`

	result, err := r.db.ExecContext(ctx, query, purchaseID, eveContractID)
	if err != nil {
		return errors.Wrap(err, "failed to complete purchase with contract ID")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("purchase transaction not found or not in contract_created status")
	}

	return nil
}

// GetPendingQuantitiesForSaleContext returns pending purchase quantities
// grouped by type_id, scoped to a specific seller's for-sale context (owner+location+container/division).
// Counts 'pending' purchases and 'contract_created' purchases within the last hour.
// Once a contract is created, EVE removes the items from the seller's hangar, but ESI
// asset caches can take up to an hour to refresh. During that window we still need to
// reserve the quantity to avoid auto-sell creating duplicate listings.
func (r *PurchaseTransactions) GetPendingQuantitiesForSaleContext(
	ctx context.Context,
	sellerUserID int64,
	ownerType string, ownerID, locationID int64,
	containerID *int64, divisionNumber *int,
) (map[int64]int64, error) {
	query := `
		SELECT pt.type_id, SUM(pt.quantity_purchased)
		FROM purchase_transactions pt
		JOIN for_sale_items f ON pt.for_sale_item_id = f.id
		WHERE pt.seller_user_id = $1
			AND (pt.status = 'pending'
				OR (pt.status = 'contract_created'
					AND pt.contract_created_at > NOW() - INTERVAL '1 hour'))
			AND f.owner_type = $2
			AND f.owner_id = $3
			AND f.location_id = $4
			AND COALESCE(f.container_id, 0::bigint) = COALESCE($5::bigint, 0::bigint)
			AND COALESCE(f.division_number, 0) = COALESCE($6, 0)
		GROUP BY pt.type_id
	`

	rows, err := r.db.QueryContext(ctx, query,
		sellerUserID, ownerType, ownerID, locationID, containerID, divisionNumber,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pending quantities for sale context")
	}
	defer rows.Close()

	result := map[int64]int64{}
	for rows.Next() {
		var typeID, quantity int64
		if err := rows.Scan(&typeID, &quantity); err != nil {
			return nil, errors.Wrap(err, "failed to scan pending quantity")
		}
		result[typeID] = quantity
	}

	return result, nil
}

// GetPendingQuantitiesByBuyer returns pending/contract_created purchase quantities
// grouped by type_id for a specific buyer.
func (r *PurchaseTransactions) GetPendingQuantitiesByBuyer(ctx context.Context, buyerUserID int64) (map[int64]int64, error) {
	query := `
		SELECT type_id, SUM(quantity_purchased)
		FROM purchase_transactions
		WHERE buyer_user_id = $1
			AND status IN ('pending', 'contract_created')
		GROUP BY type_id
	`

	rows, err := r.db.QueryContext(ctx, query, buyerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pending quantities by buyer")
	}
	defer rows.Close()

	result := map[int64]int64{}
	for rows.Next() {
		var typeID, quantity int64
		if err := rows.Scan(&typeID, &quantity); err != nil {
			return nil, errors.Wrap(err, "failed to scan pending quantity")
		}
		result[typeID] = quantity
	}

	return result, nil
}

// GetPendingQuantityForBuyOrder returns the total quantity of pending/contract_created
// purchases already created for a specific buy order.
func (r *PurchaseTransactions) GetPendingQuantityForBuyOrder(ctx context.Context, buyOrderID int64) (int64, error) {
	query := `
		SELECT COALESCE(SUM(quantity_purchased), 0)
		FROM purchase_transactions
		WHERE buy_order_id = $1
			AND status IN ('pending', 'contract_created')
	`

	var total int64
	err := r.db.QueryRowContext(ctx, query, buyOrderID).Scan(&total)
	if err != nil {
		return 0, errors.Wrap(err, "failed to query pending quantity for buy order")
	}

	return total, nil
}

// UpdateStatus updates the status of a purchase transaction.
// When transitioning to 'contract_created', also sets contract_created_at to NOW().
func (r *PurchaseTransactions) UpdateStatus(ctx context.Context, purchaseID int64, newStatus string) error {
	query := `
		UPDATE purchase_transactions
		SET status = $2,
			contract_created_at = CASE WHEN $3 = 'contract_created' THEN NOW() ELSE contract_created_at END
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, purchaseID, newStatus, newStatus)
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
