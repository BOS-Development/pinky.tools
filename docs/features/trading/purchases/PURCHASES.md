# Purchase System Documentation (Phase 4)

## Overview

The Purchase System enables users to buy items from their contacts' for-sale listings. It provides a complete transaction workflow from purchase initiation through contract creation and completion, with full atomicity guarantees and permission enforcement.

**Key Features:**
- Atomic purchase transactions (quantity updates + purchase records)
- Multi-stage workflow: pending → contract_created → completed
- Permission-based access control
- Cancel and restore functionality
- Contract key grouping for batch purchases
- Comprehensive transaction history

---

## Architecture

### Database Schema

#### purchase_transactions table
Immutable transaction log for all purchases.

```sql
CREATE TABLE purchase_transactions (
    id BIGSERIAL PRIMARY KEY,
    for_sale_item_id BIGINT NOT NULL REFERENCES for_sale_items(id),
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    seller_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_purchased BIGINT NOT NULL,
    price_per_unit BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    contract_key VARCHAR(100),
    location_id BIGINT NOT NULL,
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT purchase_positive_quantity CHECK (quantity_purchased > 0),
    CONSTRAINT purchase_different_users CHECK (buyer_user_id != seller_user_id)
);
```

**Indexes:**
- `idx_purchase_buyer` - Buyer purchase history (buyer_user_id, purchased_at DESC)
- `idx_purchase_seller` - Seller sales history (seller_user_id, purchased_at DESC)
- `idx_purchase_item` - Item-based lookups (for_sale_item_id)
- `idx_purchase_status` - Status filtering (status)

**Status Values:**
- `pending` - Purchase created, awaiting seller action
- `contract_created` - Seller created in-game contract
- `completed` - Buyer completed the contract
- `cancelled` - Purchase cancelled, quantity restored

---

## Backend Implementation

### Repository Layer

**File:** `/internal/repositories/purchaseTransactions.go`

#### Key Methods

```go
// Create records a new purchase transaction (within transaction)
func (r *PurchaseTransactions) Create(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error

// GetByID retrieves a purchase by ID with type name populated
func (r *PurchaseTransactions) GetByID(ctx context.Context, id int64) (*models.PurchaseTransaction, error)

// UpdateStatus changes purchase status
func (r *PurchaseTransactions) UpdateStatus(ctx context.Context, id int64, status string) error

// UpdateContractKeys sets contract key for multiple purchases (batch operations)
func (r *PurchaseTransactions) UpdateContractKeys(ctx context.Context, purchaseIDs []int64, contractKey string) error

// GetByBuyer returns purchase history for buyer (DESC by purchase date)
func (r *PurchaseTransactions) GetByBuyer(ctx context.Context, buyerUserID int64) ([]*models.PurchaseTransaction, error)

// GetBySeller returns sales history for seller
func (r *PurchaseTransactions) GetBySeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error)

// GetPendingForSeller returns pending purchases with buyer name and location populated
func (r *PurchaseTransactions) GetPendingForSeller(ctx context.Context, sellerUserID int64) ([]*models.PurchaseTransaction, error)
```

**Important Notes:**
- `Create` must be called within a database transaction
- `GetPendingForSeller` only returns status='pending' (excludes contract_created, completed, cancelled)
- Purchase records are immutable (no updates except status and contract_key)

---

### Controller Layer

**File:** `/internal/controllers/purchases.go`

#### API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/purchases` | Purchase item | Yes |
| POST | `/v1/purchases/{id}/mark-contract-created` | Seller marks contract created | Yes |
| POST | `/v1/purchases/{id}/complete` | Buyer completes purchase | Yes |
| POST | `/v1/purchases/{id}/cancel` | Cancel purchase and restore quantity | Yes |
| GET | `/v1/purchases/buyer` | Get buyer's purchase history | Yes |
| GET | `/v1/purchases/seller` | Get seller's sales history | Yes |
| GET | `/v1/purchases/pending` | Get seller's pending sales | Yes |

---

### Endpoint Details

#### POST /v1/purchases - Purchase Item

**Request Body:**
```json
{
  "forSaleItemId": 123,
  "quantityPurchased": 50
}
```

**Response (201):**
```json
{
  "id": 456,
  "forSaleItemId": 123,
  "buyerUserId": 1001,
  "sellerUserId": 1002,
  "typeId": 34,
  "quantityPurchased": 50,
  "pricePerUnit": 100,
  "totalPrice": 5000,
  "status": "pending",
  "purchasedAt": "2026-02-15T10:30:00Z"
}
```

**Business Logic:**
1. Retrieve for-sale item
2. **Permission check:** Verify buyer has `for_sale_browse` permission from seller (403 if denied)
3. **Self-purchase prevention:** Verify buyer ≠ seller (403 if same user)
4. **Quantity validation:** Verify quantity ≤ available (400 if exceeded)
5. **Begin database transaction**
6. Calculate new quantity: `newQuantity = available - purchased`
7. Update for-sale item quantity (marks inactive if newQuantity ≤ 0)
8. Create purchase transaction record
9. **Commit transaction** (rollback on any error)

**Error Responses:**
- `400` - Invalid request, quantity exceeded, or missing fields
- `403` - No permission or self-purchase attempt
- `404` - For-sale item not found
- `500` - Internal error

**Atomicity Guarantee:**
Both the quantity update and purchase record creation succeed or fail together. Row-level locking prevents overselling in concurrent scenarios.

---

#### POST /v1/purchases/{id}/mark-contract-created - Mark Contract Created

**Authorization:** Only seller can call this

**Request Body:**
```json
{
  "contractKey": "PT-1001-30000142-1234567890"
}
```

**Response (200):**
```json
{
  "id": 456,
  "status": "contract_created",
  "contractKey": "PT-1001-30000142-1234567890"
}
```

**Business Logic:**
1. Retrieve purchase transaction
2. **Authorization check:** Verify caller is seller (403 if not)
3. Validate contract key provided (400 if missing)
4. Update purchase status to `contract_created`
5. Store contract key

**Use Case:**
Seller creates an in-game EVE contract and stores the contract key for buyer reference.

---

#### POST /v1/purchases/{id}/complete - Complete Purchase

**Authorization:** Only buyer can call this

**Response (200):**
```json
{
  "id": 456,
  "status": "completed"
}
```

**Business Logic:**
1. Retrieve purchase transaction
2. **Authorization check:** Verify caller is buyer (403 if not)
3. Update purchase status to `completed`

**Use Case:**
Buyer accepts the in-game contract and marks the transaction complete.

---

#### POST /v1/purchases/{id}/cancel - Cancel Purchase

**Authorization:** Both buyer and seller can cancel

**Response (200):**
```json
{
  "id": 456,
  "status": "cancelled"
}
```

**Business Logic:**
1. Retrieve purchase transaction
2. **Authorization check:** Verify caller is buyer OR seller (403 if neither)
3. **Begin database transaction**
4. Update purchase status to `cancelled`
5. Retrieve original for-sale item
6. Restore quantity: `newQuantity = current + quantityPurchased`
7. Reactivate for-sale item (`is_active = true`)
8. **Commit transaction**

**Atomicity Guarantee:**
Status update and quantity restoration succeed or fail together.

---

#### GET /v1/purchases/buyer - Buyer Purchase History

**Response (200):**
```json
[
  {
    "id": 456,
    "forSaleItemId": 123,
    "buyerUserId": 1001,
    "sellerUserId": 1002,
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityPurchased": 50,
    "pricePerUnit": 100,
    "totalPrice": 5000,
    "status": "completed",
    "contractKey": "PT-1001-30000142-1234567890",
    "locationId": 30000142,
    "locationName": "Jita",
    "purchasedAt": "2026-02-15T10:30:00Z"
  }
]
```

**Notes:**
- Results ordered by `purchasedAt DESC` (newest first)
- Includes all statuses (pending, contract_created, completed, cancelled)

---

#### GET /v1/purchases/seller - Seller Sales History

**Response:** Same format as buyer history

**Notes:**
- Shows all sales made by the authenticated user
- Ordered by `purchasedAt DESC`

---

#### GET /v1/purchases/pending - Pending Sales (Seller View)

**Response (200):**
```json
[
  {
    "id": 456,
    "forSaleItemId": 123,
    "buyerUserId": 1001,
    "buyerName": "Buyer Character",
    "sellerUserId": 1002,
    "typeId": 34,
    "typeName": "Tritanium",
    "quantityPurchased": 50,
    "pricePerUnit": 100,
    "totalPrice": 5000,
    "status": "pending",
    "locationId": 30000142,
    "locationName": "Jita",
    "purchasedAt": "2026-02-15T10:30:00Z"
  }
]
```

**Notes:**
- **Only returns `status = 'pending'`** (excludes contract_created, completed, cancelled)
- Includes buyer character name for easy identification
- Includes location name for contract creation

**Use Case:**
Seller views pending purchases to create in-game contracts.

---

## Special Cases & Edge Cases

### Purchasing Entire Quantity

**Problem:** Database constraint `for_sale_positive_quantity CHECK (quantity_available > 0)` prevents setting quantity to 0.

**Solution:** Modified `UpdateQuantity` in forSaleItems repository:
```sql
UPDATE for_sale_items
SET
    quantity_available = CASE
        WHEN $2 > 0 THEN $2
        ELSE quantity_available  -- Keep original if newQuantity ≤ 0
    END,
    is_active = CASE
        WHEN $2 > 0 THEN true
        ELSE false  -- Mark inactive if newQuantity ≤ 0
    END,
    updated_at = NOW()
WHERE id = $1
```

**Verified by:** `Test_ForSaleItemsUpdateQuantityToZero_ShouldMarkInactive`

---

### Self-Purchase Prevention

Users cannot purchase their own items. However, users also cannot create contacts with themselves.

**Implementation:**
- Permission check runs first, checking if seller granted `for_sale_browse` permission to buyer
- Since users cannot create self-contacts, this check fails with 403 (no permission)
- Explicit self-purchase check (buyer_id == seller_id) is redundant but included as defensive programming

**Error Response:** 403 Forbidden with message "you do not have permission to purchase from this seller"

**Verified by:** `Test_PurchaseItem_SelfPurchase_Rejected`

---

### Concurrent Purchases

**Scenario:** Two buyers attempt to purchase the last 10 units simultaneously.

**Protection:**
1. PostgreSQL row-level locking during transaction
2. Each transaction acquires lock on for-sale item row
3. First transaction succeeds, second waits
4. Second transaction sees updated quantity and either:
   - Succeeds with reduced quantity (if some remains)
   - Fails with 400 error (if quantity exceeded)

**No overselling possible.**

---

### Contract Key Grouping

Sellers can create a single in-game contract for multiple purchase transactions.

**API Usage:**
```bash
# Mark multiple purchases with same contract key
POST /v1/purchases/bulk-mark-contract-created
{
  "purchaseIds": [456, 457, 458],
  "contractKey": "PT-1001-30000142-1234567890"
}
```

**Backend Implementation:**
```go
repo.UpdateContractKeys(ctx, []int64{456, 457, 458}, contractKey)
```

**Use Case:**
Buyer purchases Tritanium (x2), Pyerite (x1), and Mexallon (x1) from same seller at same location. Seller creates one contract containing all items.

---

## Testing

See [TESTING_PURCHASES.md](TESTING_PURCHASES.md) for complete test suite documentation.

**Test Coverage Summary:**
- 11 repository tests (CRUD, status management, queries, edge cases)
- 8 controller integration tests (full workflows, permissions, validations)
- **Total: 19 tests, all passing**

**Run Tests:**
```bash
# All purchase tests
go test ./internal/repositories -run 'PurchaseTransactions|ForSaleItemsUpdateQuantity' -v
go test ./internal/controllers -run 'Purchase|MarkContractCreated|CompletePurchase|CancelPurchase' -v

# Via Docker (recommended)
docker-compose -f docker-compose.test.yaml run --rm backend-test \
  sh -c "go test -v ./internal/repositories -run 'PurchaseTransactions|ForSaleItemsUpdateQuantity'"
```

---

## Usage Examples

### Example 1: Complete Purchase Flow

```bash
# 1. Buyer browses marketplace and purchases 100 Tritanium
POST /v1/purchases
{
  "forSaleItemId": 123,
  "quantityPurchased": 100
}
# Response: { "id": 456, "status": "pending", ... }

# 2. Seller views pending sales
GET /v1/purchases/pending
# Response: [{ "id": 456, "buyerName": "John Doe", "typeName": "Tritanium", ... }]

# 3. Seller creates in-game contract and marks it
POST /v1/purchases/456/mark-contract-created
{
  "contractKey": "PT-1001-30000142-1234567890"
}
# Response: { "id": 456, "status": "contract_created", ... }

# 4. Buyer accepts contract in-game and completes purchase
POST /v1/purchases/456/complete
# Response: { "id": 456, "status": "completed" }
```

---

### Example 2: Cancelling a Purchase

```bash
# Buyer changes mind before contract is created
POST /v1/purchases/456/cancel
# Response: { "id": 456, "status": "cancelled" }

# Quantity is automatically restored to for-sale item
# Item is reactivated if it was marked inactive
```

---

### Example 3: Batch Contract Creation

```bash
# Buyer purchases multiple items from same seller/location
POST /v1/purchases
{ "forSaleItemId": 123, "quantityPurchased": 100 }  # Tritanium
# Response: { "id": 456 }

POST /v1/purchases
{ "forSaleItemId": 124, "quantityPurchased": 50 }   # Pyerite
# Response: { "id": 457 }

# Seller creates single contract for both items
POST /v1/purchases/bulk-mark-contract-created
{
  "purchaseIds": [456, 457],
  "contractKey": "PT-1001-30000142-1234567890"
}
```

---

## Integration with For-Sale Items

The purchase system tightly integrates with the for-sale items system:

**Quantity Management:**
- Purchase decreases `quantity_available`
- Cancel restores `quantity_available`
- When quantity reaches 0, item marked `is_active = false`
- When quantity restored from 0, item reactivated

**Permission Enforcement:**
- Only buyers with `for_sale_browse` permission can purchase
- Permission checked on every purchase attempt
- Permission revocation prevents new purchases (existing purchases unaffected)

---

## Error Handling

### Common Error Scenarios

| Error | Status | Cause | Solution |
|-------|--------|-------|----------|
| "for-sale item not found" | 404 | Invalid ID or item deleted | Verify item exists |
| "you do not have permission" | 403 | No for_sale_browse permission | Request permission from seller |
| "cannot purchase your own items" | 403 | buyer_id == seller_id | Contact another seller |
| "quantity exceeds available" | 400 | Requested > available | Reduce quantity or cancel |
| "purchase transaction not found" | 404 | Invalid purchase ID | Verify purchase exists |
| "only the seller can mark contract created" | 403 | Wrong user | Seller must perform action |
| "only the buyer can complete purchase" | 403 | Wrong user | Buyer must perform action |

---

## Database Migrations

**Migration File:** `/internal/database/migrations/10_purchase_transactions.up.sql`

```sql
CREATE TABLE purchase_transactions (
    id BIGSERIAL PRIMARY KEY,
    for_sale_item_id BIGINT NOT NULL REFERENCES for_sale_items(id),
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    seller_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_purchased BIGINT NOT NULL,
    price_per_unit BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    contract_key VARCHAR(100),
    location_id BIGINT NOT NULL,
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT purchase_positive_quantity CHECK (quantity_purchased > 0),
    CONSTRAINT purchase_different_users CHECK (buyer_user_id != seller_user_id)
);

CREATE INDEX idx_purchase_buyer ON purchase_transactions(buyer_user_id, purchased_at DESC);
CREATE INDEX idx_purchase_seller ON purchase_transactions(seller_user_id, purchased_at DESC);
CREATE INDEX idx_purchase_item ON purchase_transactions(for_sale_item_id);
CREATE INDEX idx_purchase_status ON purchase_transactions(status);
```

**Run Migration:**
```bash
migrate -path internal/database/migrations -database "postgresql://user:pass@localhost:5432/dbname" up
```

---

## Performance Considerations

### Indexes

All critical query paths are indexed:
- Buyer history: `idx_purchase_buyer` (buyer_user_id, purchased_at DESC)
- Seller history: `idx_purchase_seller` (seller_user_id, purchased_at DESC)
- Item lookups: `idx_purchase_item` (for_sale_item_id)
- Status filtering: `idx_purchase_status` (status)

### Query Optimization

**GetPendingForSeller:**
```sql
SELECT pt.*,
       it.type_name,
       buyer_char.name as buyer_name,
       ss.name as location_name
FROM purchase_transactions pt
LEFT JOIN asset_item_types it ON pt.type_id = it.type_id
LEFT JOIN characters buyer_char ON pt.buyer_user_id = buyer_char.user_id
LEFT JOIN solar_systems ss ON pt.location_id = ss.solar_system_id
WHERE pt.seller_user_id = $1
  AND pt.status = 'pending'
ORDER BY pt.purchased_at DESC
```

**Note:** JOIN on `buyer_char.user_id` (not `buyer_char.id`) because purchase stores user_id, not character_id.

---

## Security

### Authorization Checks

Every endpoint enforces authorization:
- Purchase: Requires `for_sale_browse` permission
- Mark contract created: Only seller
- Complete purchase: Only buyer
- Cancel: Buyer OR seller
- View history: Own purchases/sales only

### SQL Injection Prevention

All queries use parameterized statements:
```go
db.QueryRowContext(ctx, "SELECT * FROM purchase_transactions WHERE id = $1", id)
```

### Transaction Safety

Critical operations wrapped in database transactions:
- Purchase (quantity update + record creation)
- Cancel (status update + quantity restoration)

Rollback on any error ensures consistency.

---

## Future Enhancements (Phase 5+)

**Buy Orders:**
- Buyers place orders for out-of-stock items
- Sellers see demand and can create listings

**Analytics:**
- Sales volume charts
- Revenue tracking
- Top-selling items
- Buyer analytics

**Automation:**
- Auto-accept contracts via ESI
- Price alerts
- Inventory restocking suggestions

---

## Troubleshooting

### Purchase Fails with "no permission"

**Check:**
1. Is contact relationship accepted? (status = 'accepted')
2. Did seller grant `for_sale_browse` permission?
3. Query: `SELECT * FROM contact_permissions WHERE granting_user_id = <seller> AND receiving_user_id = <buyer> AND service_type = 'for_sale_browse'`

### Quantity Not Updating

**Check:**
1. Was purchase successful? (verify transaction committed)
2. Query: `SELECT * FROM for_sale_items WHERE id = <item_id>`
3. Check `quantity_available` and `is_active` fields

### Buyer Name Shows as "User XXX" in Pending Sales

**Root Cause:** Character not found for buyer's user_id

**Fix:**
1. Verify buyer has character: `SELECT * FROM characters WHERE user_id = <buyer_id>`
2. If missing, buyer needs to add character via ESI

---

## Related Documentation

- [TESTING_PURCHASES.md](TESTING_PURCHASES.md) - Test suite documentation
- [Contact System Plan](/home/benjamin/.claude/plans/serialized-growing-starfish.md) - Overall marketplace architecture
- `/internal/repositories/purchaseTransactions.go` - Repository implementation
- `/internal/controllers/purchases.go` - Controller implementation

---

## Changelog

**Phase 4 (2026-02-15):**
- ✅ Initial purchase system implementation
- ✅ Multi-stage workflow (pending → contract_created → completed)
- ✅ Cancel and restore functionality
- ✅ Contract key grouping
- ✅ Comprehensive test suite (19 tests)
- ✅ Fixed: Purchasing entire quantity constraint violation
- ✅ Fixed: Pending sales filtering (only 'pending' status)
- ✅ Fixed: Character JOIN in GetPendingForSeller query
