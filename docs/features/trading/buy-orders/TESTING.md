# Buy Orders Testing Guide

Comprehensive test suite documentation for Phase 5.

## Test Coverage Summary

| Category | Tests | Status |
|----------|-------|--------|
| Repository Tests | 8 | ✅ All Passing |
| Controller Tests | 7 | ✅ All Passing |
| **Total** | **15** | ✅ **100% Passing** |

## Running Tests

### Run All Buy Orders Tests

```bash
# Repository tests
go test -v ./internal/repositories/... -run BuyOrders

# Controller tests
go test -v ./internal/controllers/... -run BuyOrders

# All tests
make test-all
```

### Run Specific Test

```bash
go test -v ./internal/repositories/... -run Test_BuyOrders_CreateAndGet
```

### Run with Coverage

```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html
```

## Repository Tests

### Test_BuyOrders_CreateAndGet

**File:** `internal/repositories/buyOrders_test.go`

**Tests:** Basic CRUD operations - create and retrieve a buy order.

```go
func Test_BuyOrders_CreateAndGet(t *testing.T) {
    // Create user and item type
    // Create buy order
    // Verify order created with correct values
    // Get by ID and verify
    // Check TypeName populated from join
}
```

**Validates:**
- Order creation with all fields
- Auto-generated ID, timestamps
- Foreign key relationships (user, type)
- JOIN with asset_item_types for typeName

---

### Test_BuyOrders_GetByUser

**Tests:** Retrieving all buy orders for a specific user.

```go
func Test_BuyOrders_GetByUser(t *testing.T) {
    // Create multiple buy orders for user
    // Get by user ID
    // Verify all orders returned
    // Check ordering (DESC by created_at)
    // Verify type names populated
}
```

**Validates:**
- Multiple orders returned correctly
- Correct ordering (newest first)
- Type names populated
- User isolation (only user's orders)

---

### Test_BuyOrders_Update

**Tests:** Updating buy order fields.

```go
func Test_BuyOrders_Update(t *testing.T) {
    // Create buy order
    // Update quantity, price, notes
    // Verify updates applied
    // Check updated_at changed
}
```

**Validates:**
- Quantity updates
- Price updates
- Notes updates (including null)
- Timestamp updates

---

### Test_BuyOrders_Delete

**Tests:** Soft delete functionality.

```go
func Test_BuyOrders_Delete(t *testing.T) {
    // Create buy order
    // Delete order
    // Verify is_active = false
    // Verify not in active results
}
```

**Validates:**
- Soft delete (is_active = false)
- Order still exists in database
- Doesn't appear in active queries
- User ownership check

---

### Test_BuyOrders_GetDemandForSeller

**Tests:** Permission-based demand viewing.

```go
func Test_BuyOrders_GetDemandForSeller(t *testing.T) {
    // Create buyer and seller users
    // Create contact relationship
    // Grant for_sale_browse permission
    // Create buy orders
    // Seller queries demand
    // Verify only active, permitted orders returned
}
```

**Validates:**
- Permission system integration
- Contact relationship required
- Only active orders shown
- Correct JOIN with permissions table
- Type names populated

---

### Test_BuyOrders_GetByID_NotFound

**Tests:** Error handling for non-existent orders.

```go
func Test_BuyOrders_GetByID_NotFound(t *testing.T) {
    // Query non-existent ID
    // Verify error returned
    // Check error message
}
```

**Validates:**
- Proper error for missing orders
- Error message clarity

---

### Test_BuyOrders_Update_NotFound

**Tests:** Error handling when updating non-existent order.

```go
func Test_BuyOrders_Update_NotFound(t *testing.T) {
    // Attempt to update non-existent order
    // Verify error returned
}
```

**Validates:**
- Update fails gracefully for missing orders
- Appropriate error message

---

### Test_BuyOrders_Delete_NotFound

**Tests:** Error handling when deleting non-existent order.

```go
func Test_BuyOrders_Delete_NotFound(t *testing.T) {
    // Attempt to delete non-existent order
    // Verify error returned
}
```

**Validates:**
- Delete fails gracefully for missing orders
- User ownership check

## Controller Tests

### Test_BuyOrders_CreateOrder_Success

**File:** `internal/controllers/buyOrders_test.go`

**Tests:** Successful buy order creation via API.

```go
func Test_BuyOrders_CreateOrder_Success(t *testing.T) {
    // Setup database and users
    // POST /v1/buy-orders with valid data
    // Verify 200 response
    // Check returned order data
}
```

**Validates:**
- Request parsing
- Validation passes
- Order created in database
- Correct response format
- Logging

---

### Test_BuyOrders_CreateOrder_InvalidQuantity

**Tests:** Validation for invalid quantity.

```go
func Test_BuyOrders_CreateOrder_InvalidQuantity(t *testing.T) {
    // POST with quantityDesired = -100
    // Verify 400 Bad Request
    // Check error message
}
```

**Validates:**
- Quantity must be positive
- 400 status code
- Clear error message

---

### Test_BuyOrders_GetMyOrders

**Tests:** Retrieving user's buy orders.

```go
func Test_BuyOrders_GetMyOrders(t *testing.T) {
    // Create multiple buy orders
    // GET /v1/buy-orders
    // Verify all orders returned
    // Check type names populated
}
```

**Validates:**
- User authentication
- All user's orders returned
- Type names populated
- Correct JSON format

---

### Test_BuyOrders_UpdateOrder_Success

**Tests:** Successful order update.

```go
func Test_BuyOrders_UpdateOrder_Success(t *testing.T) {
    // Create order
    // PUT /v1/buy-orders/{id} with updates
    // Verify 200 response
    // Check updates applied
}
```

**Validates:**
- URL parameter parsing
- Update validation
- Owner verification
- Response format
- Logging

---

### Test_BuyOrders_UpdateOrder_NotOwner

**Tests:** Permission check for updates.

```go
func Test_BuyOrders_UpdateOrder_NotOwner(t *testing.T) {
    // User A creates order
    // User B tries to update
    // Verify 403 Forbidden
    // Check error message
}
```

**Validates:**
- Owner-only updates
- 403 status code
- Clear error message

---

### Test_BuyOrders_DeleteOrder_Success

**Tests:** Successful order deletion.

```go
func Test_BuyOrders_DeleteOrder_Success(t *testing.T) {
    // Create order
    // DELETE /v1/buy-orders/{id}
    // Verify 200 response
    // Check is_active = false
}
```

**Validates:**
- Soft delete via API
- Owner verification
- Response format
- Logging

---

### Test_BuyOrders_GetDemand

**Tests:** Permission-based demand endpoint.

```go
func Test_BuyOrders_GetDemand(t *testing.T) {
    // Setup buyer and seller
    // Create contact with permission
    // Create buy orders
    // Seller GET /v1/buy-orders/demand
    // Verify only permitted orders returned
}
```

**Validates:**
- Permission system integration
- Contact relationship check
- Only active orders
- Type names populated
- Correct filtering

## Test Database Setup

Each test creates an isolated PostgreSQL database:

```go
func setupDatabase() (*sql.DB, error) {
    databaseName := "testDatabase_" + strconv.Itoa(rand.Int())

    // Connection pooling limits to prevent exhaustion
    db.SetMaxOpenConns(5)
    db.SetMaxIdleConns(2)

    // Migrate database
    settings.MigrateUp()

    return db, nil
}
```

**Features:**
- Isolated databases per test
- Automatic migration
- Connection pooling (5 max open, 2 idle)
- Cleanup after tests

## Mock Objects

### MockRouter

Mocks the router for controller tests:

```go
type MockRouter struct{}

func (m *MockRouter) RegisterRestAPIRoute(
    path string,
    auth web.AuthAccess,
    handler func(*web.HandlerArgs) (any, *web.HttpError),
    methods ...string,
) {
    // No-op for tests
}
```

### MockContactPermissionsRepository

Mocks permission checks:

```go
type MockContactPermissionsRepository struct {
    mock.Mock
}

func (m *MockContactPermissionsRepository) CheckPermission(
    ctx context.Context,
    grantingUserID, receivingUserID int64,
    serviceType string,
) (bool, error) {
    args := m.Called(ctx, grantingUserID, receivingUserID, serviceType)
    return args.Bool(0), args.Error(1)
}
```

## Common Test Patterns

### Setup Pattern

```go
func Test_Something(t *testing.T) {
    // 1. Setup database
    db, err := setupDatabase()
    assert.NoError(t, err)

    // 2. Create test data
    userRepo := repositories.NewUserRepository(db)
    user := &repositories.User{ID: 1000, Name: "Test User"}
    assert.NoError(t, userRepo.Add(context.Background(), user))

    // 3. Create repository/controller
    repo := repositories.NewBuyOrders(db)

    // 4. Execute test
    // 5. Assert results
}
```

### Assertion Pattern

```go
// Assert no error
assert.NoError(t, err)

// Assert value equality
assert.Equal(t, expected, actual)

// Assert true/false
assert.True(t, condition)
assert.False(t, condition)

// Assert not zero/nil
assert.NotZero(t, value)
assert.NotNil(t, pointer)

// Assert collection length
assert.Len(t, collection, expectedLength)

// Assert contains
assert.Contains(t, string, substring)
```

## Test Data Conventions

### User IDs
- Repository tests: 5000-5999
- Controller tests: 6000-6999
- Character IDs: User ID * 10 (e.g., user 5000 → char 50000)

### Type IDs
- Tritanium: 34
- Pyerite: 35
- Mexallon: 36
- Isogen: 37
- Other minerals: 38+

### System IDs
- Jita: 30000142
- Amarr: 30002187

## Debugging Tests

### Run with Verbose Output

```bash
go test -v ./internal/repositories/... -run BuyOrders
```

### Run Single Test

```bash
go test -v ./internal/repositories/... -run Test_BuyOrders_CreateAndGet
```

### Print Debug Info

Add temporary debug prints:

```go
func Test_Something(t *testing.T) {
    // ... test code ...

    // Debug output
    t.Logf("Order: %+v", order)
    t.Logf("Count: %d", len(orders))
}
```

### Check Database State

```go
func Test_Something(t *testing.T) {
    // ... test code ...

    // Query database directly
    var count int
    db.QueryRow("SELECT COUNT(*) FROM buy_orders").Scan(&count)
    t.Logf("Buy orders in database: %d", count)
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: make test-all
```

### Test Coverage Reporting

```bash
# Generate coverage
go test -coverprofile=coverage.out ./internal/...

# View in browser
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out
```

## Common Issues

### Too Many Connections

**Error:** `pq: sorry, too many clients already`

**Solution:** Connection pooling limits added to test setup:
```go
db.SetMaxOpenConns(5)
db.SetMaxIdleConns(2)
```

### Foreign Key Violations

**Error:** `violates foreign key constraint`

**Solution:** Ensure test data created in correct order:
1. Users
2. Characters
3. Item types
4. Regions/Systems
5. Buy orders

### Test Isolation Issues

**Problem:** Tests pass individually but fail when run together.

**Solutions:**
- Use unique user IDs per test
- Random database names
- Proper cleanup
- Transaction rollback where appropriate

## Best Practices

1. **One Assertion Per Test** - Focus each test on a single behavior
2. **Descriptive Names** - Test names should describe what they test
3. **Arrange-Act-Assert** - Follow AAA pattern
4. **Independent Tests** - Tests shouldn't depend on each other
5. **Clean Test Data** - Use unique IDs to prevent conflicts
6. **Test Edge Cases** - Not just happy paths
7. **Mock External Dependencies** - Don't call real external APIs
8. **Fast Tests** - Keep tests quick (< 100ms each)

## Next Steps

- Add integration tests for full user flows
- Add performance/load tests
- Add frontend component tests
- Test error scenarios more thoroughly
- Add API contract tests
