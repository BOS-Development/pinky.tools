# Purchase System Test Suite

This document describes the comprehensive test suite for the purchase transaction system (Phase 4 of the marketplace feature).

**Related Documentation:**
- [PURCHASES.md](PURCHASES.md) - Complete technical documentation
- [API_PURCHASES.md](API_PURCHASES.md) - API reference
- [QUICK_START_PURCHASES.md](QUICK_START_PURCHASES.md) - Quick start guide

---

## Test Coverage

### Repository Tests (`internal/repositories/purchaseTransactions_test.go`)

#### Create and Retrieve Tests
- `Test_PurchaseTransactions_CreateAndGet`: Verifies creating a purchase transaction within a database transaction and retrieving it by ID
- `Test_PurchaseTransactions_GetByID_NotFound`: Tests error handling when purchase doesn't exist

#### Status Management Tests
- `Test_PurchaseTransactions_UpdateStatus`: Tests status transitions (pending → contract_created → completed)
- `Test_PurchaseTransactions_UpdateStatus_NotFound`: Tests error handling for non-existent purchases

#### Contract Key Management Tests
- `Test_PurchaseTransactions_UpdateContractKeys`: Tests bulk contract key updates for multiple purchases (grouped purchases scenario)
- `Test_PurchaseTransactions_UpdateContractKeys_EmptyArray`: Tests edge case with empty purchase ID array

#### Query Tests
- `Test_PurchaseTransactions_GetByBuyer`: Tests buyer purchase history with ordering (DESC by purchase date)
- `Test_PurchaseTransactions_GetBySeller`: Tests seller sales history
- `Test_PurchaseTransactions_GetPendingForSeller`: Tests pending sales query with buyer name and location name population (only returns 'pending' status)

### For-Sale Items Quantity Update Tests (`internal/repositories/forSaleItems_test.go`)

#### Constraint Fix Tests
- `Test_ForSaleItemsUpdateQuantityToZero_ShouldMarkInactive`:
  - Tests the fix for purchasing entire quantity
  - Verifies item is marked inactive when quantity becomes 0
  - Ensures original quantity is preserved (doesn't violate `for_sale_positive_quantity` constraint)

- `Test_ForSaleItemsUpdateQuantityPartial_ShouldUpdateQuantity`:
  - Tests partial quantity updates
  - Verifies item remains active with updated quantity

### Controller Integration Tests (`internal/controllers/purchases_test.go`)

#### Successful Purchase Flow
- `Test_PurchaseItem_Success`:
  - Tests complete purchase flow with permission checking
  - Verifies transaction atomicity (quantity update + purchase record creation)
  - Confirms purchase details are correct

- `Test_PurchaseItem_EntireQuantity_MarksInactive`:
  - Tests purchasing all available quantity
  - Verifies item is marked inactive after complete purchase
  - Ensures no constraint violations

#### Permission and Validation Tests
- `Test_PurchaseItem_NoPermission`:
  - Tests that purchases are blocked without `for_sale_browse` permission
  - Verifies 403 Forbidden response

- `Test_PurchaseItem_SelfPurchase_Rejected`:
  - Tests prevention of buying your own items
  - Verifies 400 Bad Request response

- `Test_PurchaseItem_QuantityExceeded`:
  - Tests quantity validation
  - Ensures cannot purchase more than available

#### Status Transition Tests
- `Test_MarkContractCreated_Success`:
  - Tests seller marking purchase as contract_created
  - Verifies contract key is stored
  - Confirms status transition

- `Test_CompletePurchase_Success`:
  - Tests buyer completing the purchase (contract_created → completed)
  - Verifies final status

#### Cancel and Restore Tests
- `Test_CancelPurchase_RestoresQuantity`:
  - Tests complete purchase and cancel flow
  - Verifies quantity is restored to for-sale item
  - Confirms purchase status changes to 'cancelled'
  - Tests transaction atomicity

## Running the Tests

### Run All Tests
```bash
go test ./internal/repositories/... -v
go test ./internal/controllers/... -v
```

### Run Specific Test Suite
```bash
# Purchase transaction repository tests
go test ./internal/repositories -run TestPurchaseTransactions -v

# For-sale items quantity tests
go test ./internal/repositories -run Test_ForSaleItemsUpdateQuantity -v

# Purchase controller integration tests
go test ./internal/controllers -run Test_Purchase -v
```

### Run Single Test
```bash
go test ./internal/controllers -run Test_PurchaseItem_Success -v
```

## Database Setup

Tests use the pattern from `common_test.go`:
- Each test creates a new temporary database
- Migrations are run automatically
- Database name format: `testDatabase_{random_number}`

### Environment Variables
```bash
DATABASE_HOST=localhost      # Default: localhost
DATABASE_PORT=5432          # Default: 5432
DATABASE_USER=postgres      # Default: postgres
DATABASE_PASSWORD=postgres  # Default: postgres
```

## Key Test Patterns

### Transaction Testing
Tests that involve database transactions use this pattern:
```go
tx, err := db.BeginTx(context.Background(), nil)
defer tx.Rollback()

// Perform operations within transaction
err = repo.Create(ctx, tx, item)

// Commit if successful
err = tx.Commit()
```

### Setup Helpers
- `setupPurchaseTestData()`: Creates users, characters, items, locations, permissions
- `setupPurchasesTestDB()`: Creates base location data (regions, constellations, systems)
- `setupForSaleTestData()`: Creates minimal test data for for-sale items

## Test Coverage Summary

| Component | Tests | Coverage |
|-----------|-------|----------|
| PurchaseTransactions Repository | 8 | Create, Update, Query, Edge cases |
| ForSaleItems UpdateQuantity | 2 | Zero quantity, Partial quantity |
| Purchases Controller | 9 | Full flow, Permissions, Validations, Status transitions |

## Critical Scenarios Tested

1. ✅ **Atomic Purchase**: Quantity update and purchase record creation in single transaction
2. ✅ **Constraint Compliance**: Purchasing entire quantity doesn't violate database constraints
3. ✅ **Permission Enforcement**: Cannot purchase without browse permission
4. ✅ **Self-Purchase Prevention**: Cannot buy your own items
5. ✅ **Quantity Validation**: Cannot purchase more than available
6. ✅ **Status Workflow**: pending → contract_created → completed
7. ✅ **Contract Key Grouping**: Multiple purchases share same contract key
8. ✅ **Cancel and Restore**: Cancelled purchases restore quantity
9. ✅ **Pending Sales Filter**: Only 'pending' status appears in pending sales view

## Bug Fixes Verified

### Issue: Database Constraint Violation
**Problem**: When purchasing entire quantity, `UpdateQuantity` tried to set `quantity_available = 0`, violating `for_sale_positive_quantity` constraint.

**Fix**: Modified UPDATE query to use CASE statement:
- When `newQuantity > 0`: Update to new value
- When `newQuantity <= 0`: Keep original value, mark as inactive

**Test**: `Test_ForSaleItemsUpdateQuantityToZero_ShouldMarkInactive` verifies the fix works.

### Issue: Contract_Created Items in Pending Sales
**Problem**: Items with status 'contract_created' appeared in pending sales view.

**Fix**: Changed `GetPendingForSeller` query to only return `status = 'pending'`.

**Test**: `Test_PurchaseTransactions_GetPendingForSeller` verifies only pending items are returned.
