# Buy Orders Implementation Summary

Technical architecture and implementation details for Phase 5.

## Overview

Phase 5 extends the Contact & Marketplace system with buy orders and demand tracking. This allows users to express interest in purchasing items, with permission-based visibility to potential sellers.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    User Interface (React/MUI)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │  BuyOrders.tsx       │      │  DemandViewer.tsx    │        │
│  │                      │      │                      │        │
│  │  - Create buy order  │      │  - View aggregated   │        │
│  │  - Item autocomplete │      │    demand            │        │
│  │  - CRUD operations   │      │  - Individual orders │        │
│  │  - Form validation   │      │  - Search/filter     │        │
│  └──────────┬───────────┘      └──────────┬───────────┘        │
│             │                              │                    │
└─────────────┼──────────────────────────────┼────────────────────┘
              │                              │
              │  Next.js API Routes (Proxy)  │
              │                              │
┌─────────────▼──────────────────────────────▼────────────────────┐
│                    Backend (Go)                                  │
├─────────────────────────────────────────────────────────────────┤
│  Controllers Layer                                               │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │  BuyOrders           │      │  ItemTypes           │        │
│  │  - CreateOrder       │      │  - SearchItemTypes   │        │
│  │  - GetMyOrders       │      │                      │        │
│  │  - UpdateOrder       │      │  Returns item data   │        │
│  │  - DeleteOrder       │      │  for autocomplete    │        │
│  │  - GetDemand         │      │                      │        │
│  └──────────┬───────────┘      └──────────┬───────────┘        │
│             │                              │                    │
├─────────────┼──────────────────────────────┼────────────────────┤
│  Repository Layer                          │                    │
│  ┌──────────▼───────────┐      ┌──────────▼───────────┐        │
│  │  BuyOrders           │      │  ItemTypes           │        │
│  │  - Create            │      │  - SearchItemTypes   │        │
│  │  - GetByID           │      │  - GetItemTypeByName │        │
│  │  - GetByUser         │      │                      │        │
│  │  - GetDemandForSeller│      │  (Enhanced for       │        │
│  │  - Update            │      │   autocomplete)      │        │
│  │  - Delete            │      │                      │        │
│  └──────────┬───────────┘      └──────────┬───────────┘        │
│             │                              │                    │
└─────────────┼──────────────────────────────┼────────────────────┘
              │                              │
              │  PostgreSQL Database         │
              │                              │
┌─────────────▼──────────────────────────────▼────────────────────┐
│  ┌────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
│  │  buy_orders    │  │ asset_item_types │  │ contact_perms   │ │
│  │                │  │                  │  │                 │ │
│  │  id            │  │ type_id          │  │ contact_id      │ │
│  │  buyer_user_id │  │ type_name        │  │ granting_user   │ │
│  │  type_id ──────┼──┼─►               │  │ receiving_user  │ │
│  │  quantity      │  │ volume           │  │ service_type    │ │
│  │  price         │  │ icon_id          │  │ can_access      │ │
│  │  notes         │  └──────────────────┘  └─────────────────┘ │
│  │  is_active     │                                             │
│  │  created_at    │                                             │
│  │  updated_at    │                                             │
│  └────────────────┘                                             │
└─────────────────────────────────────────────────────────────────┘
```

## Database Layer

### Migration 12: buy_orders

**File:** `internal/database/migrations/12_buy_orders.up.sql`

```sql
CREATE TABLE buy_orders (
    id BIGSERIAL PRIMARY KEY,
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_desired BIGINT NOT NULL,
    max_price_per_unit BIGINT NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT buy_order_positive_quantity CHECK (quantity_desired > 0),
    CONSTRAINT buy_order_positive_price CHECK (max_price_per_unit >= 0)
);

CREATE INDEX idx_buy_orders_buyer ON buy_orders(buyer_user_id);
CREATE INDEX idx_buy_orders_type ON buy_orders(type_id);
CREATE INDEX idx_buy_orders_active ON buy_orders(is_active);
```

**Design Decisions:**
- **Soft Delete**: `is_active` flag instead of hard delete preserves history
- **Constraints**: CHECK constraints ensure data integrity at database level
- **Indexes**: Optimized for common queries (by buyer, by type, by status)
- **Timestamps**: Auto-managed created_at and updated_at
- **Foreign Keys**: Ensure referential integrity to users and item types

### Enhanced ItemType Repository

**File:** `internal/repositories/itemType.go`

Added two new methods for autocomplete functionality:

```go
func (r *ItemTypeRepository) SearchItemTypes(
    ctx context.Context,
    query string,
    limit int,
) ([]models.EveInventoryType, error) {
    // Case-insensitive LIKE search
    // Relevance-based sorting:
    //   1. Exact match
    //   2. Starts with query
    //   3. Contains query
    // Limited to 20 results
}

func (r *ItemTypeRepository) GetItemTypeByName(
    ctx context.Context,
    typeName string,
) (*models.EveInventoryType, error) {
    // Exact name match
    // Used for validation
}
```

**Query Optimization:**
```sql
-- Relevance-based ordering
ORDER BY
    CASE
        WHEN LOWER(type_name) = LOWER($3) THEN 1
        WHEN LOWER(type_name) LIKE LOWER($3) || '%' THEN 2
        ELSE 3
    END,
    type_name
LIMIT $2
```

## Repository Layer

### BuyOrders Repository

**File:** `internal/repositories/buyOrders.go`

Six methods implementing data access:

#### 1. Create

```go
func (r *BuyOrders) Create(
    ctx context.Context,
    order *models.BuyOrder,
) error
```

**Features:**
- Inserts new buy order
- Returns auto-generated ID and timestamps
- Uses RETURNING clause for efficiency

#### 2. GetByID

```go
func (r *BuyOrders) GetByID(
    ctx context.Context,
    id int64,
) (*models.BuyOrder, error)
```

**Features:**
- LEFT JOIN with asset_item_types for type_name
- Single query with all data
- Returns error if not found

#### 3. GetByUser

```go
func (r *BuyOrders) GetByUser(
    ctx context.Context,
    userID int64,
) ([]*models.BuyOrder, error)
```

**Features:**
- Returns all orders for a user (active and inactive)
- Ordered by created_at DESC (newest first)
- Includes type names via JOIN

#### 4. GetDemandForSeller

```go
func (r *BuyOrders) GetDemandForSeller(
    ctx context.Context,
    sellerUserID int64,
) ([]*models.BuyOrder, error)
```

**Features:**
- INNER JOIN with contact_permissions
- Filters by service_type = 'for_sale_browse'
- Only active orders (is_active = true)
- Permission-based access control

**Query:**
```sql
SELECT DISTINCT
    bo.id, bo.buyer_user_id, bo.type_id, it.type_name,
    bo.quantity_desired, bo.max_price_per_unit, bo.notes,
    bo.is_active, bo.created_at, bo.updated_at
FROM buy_orders bo
LEFT JOIN asset_item_types it ON bo.type_id = it.type_id
INNER JOIN contact_permissions cp
    ON cp.granting_user_id = bo.buyer_user_id
    AND cp.receiving_user_id = $1
    AND cp.service_type = 'for_sale_browse'
    AND cp.can_access = true
WHERE bo.is_active = true
ORDER BY bo.created_at DESC
```

#### 5. Update

```go
func (r *BuyOrders) Update(
    ctx context.Context,
    order *models.BuyOrder,
) error
```

**Features:**
- Updates quantity, price, notes, is_active
- Auto-updates updated_at timestamp
- Uses RETURNING for new timestamp

#### 6. Delete

```go
func (r *BuyOrders) Delete(
    ctx context.Context,
    id int64,
    userID int64,
) error
```

**Features:**
- Soft delete (sets is_active = false)
- Verifies user ownership via WHERE clause
- Returns error if no rows affected

## Controller Layer

### BuyOrders Controller

**File:** `internal/controllers/buyOrders.go`

Five HTTP endpoints with validation and error handling:

#### 1. CreateOrder

```go
POST /v1/buy-orders
```

**Validation:**
- typeId required
- quantityDesired > 0
- maxPricePerUnit >= 0

**Flow:**
1. Parse JSON request
2. Validate inputs
3. Call repository.Create()
4. Log creation
5. Return created order

#### 2. GetMyOrders

```go
GET /v1/buy-orders
```

**Flow:**
1. Get user ID from auth context
2. Call repository.GetByUser()
3. Return orders (or empty array)

#### 3. UpdateOrder

```go
PUT /v1/buy-orders/{id}
```

**Validation:**
- ID must be valid integer
- User must own the order
- quantityDesired > 0 if provided
- maxPricePerUnit >= 0 if provided

**Flow:**
1. Parse URL parameter (ID)
2. Parse JSON body
3. Get existing order
4. Verify ownership
5. Apply updates
6. Call repository.Update()
7. Log update
8. Return updated order

#### 4. DeleteOrder

```go
DELETE /v1/buy-orders/{id}
```

**Flow:**
1. Parse URL parameter (ID)
2. Call repository.Delete() with user ID
3. Repository verifies ownership
4. Log deletion
5. Return deleted order

#### 5. GetDemand

```go
GET /v1/buy-orders/demand
```

**Flow:**
1. Get user ID from auth context
2. Call repository.GetDemandForSeller()
3. Return filtered orders (or empty array)

### ItemTypes Controller

**File:** `internal/controllers/itemTypes.go`

One endpoint for autocomplete:

```go
GET /v1/item-types/search?q={query}
```

**Flow:**
1. Get query parameter
2. Return empty array if query empty
3. Call repository.SearchItemTypes()
4. Return up to 20 results

## Frontend Layer

### Component Architecture

```
marketplace.tsx (Page)
├── Tabs
    ├── My Listings
    ├── Browse
    ├── Pending Sales
    ├── History
    ├── My Buy Orders ← BuyOrders.tsx
    └── Demand ← DemandViewer.tsx
```

### BuyOrders Component

**File:** `frontend/packages/components/marketplace/BuyOrders.tsx`

**State Management:**

```typescript
const [orders, setOrders] = useState<BuyOrder[]>([]);
const [loading, setLoading] = useState(true);
const [dialogOpen, setDialogOpen] = useState(false);
const [selectedOrder, setSelectedOrder] = useState<BuyOrder | null>(null);
const [formData, setFormData] = useState<Partial<BuyOrderFormData>>({});
const [itemOptions, setItemOptions] = useState<ItemType[]>([]);
const [itemSearchLoading, setItemSearchLoading] = useState(false);
const [selectedItem, setSelectedItem] = useState<ItemType | null>(null);
const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);
```

**Key Features:**

1. **Autocomplete with Debounce**
```typescript
const handleItemSearch = (value: string) => {
  if (searchTimeoutRef.current) {
    clearTimeout(searchTimeoutRef.current);
  }

  searchTimeoutRef.current = setTimeout(() => {
    searchItems(value);
  }, 300);
};
```

2. **Icon Rendering**
```typescript
<Avatar
  src={getItemIconUrl(option.TypeID, 32)}
  alt={option.TypeName}
  sx={{ width: 32, height: 32 }}
  variant="square"
/>
```

3. **Form Validation**
```typescript
if (!formData.typeId || !formData.quantityDesired || formData.maxPricePerUnit === undefined) {
  showSnackbar('Please fill in all required fields', 'error');
  return;
}
```

### DemandViewer Component

**File:** `frontend/packages/components/marketplace/DemandViewer.tsx`

**Demand Aggregation:**

```typescript
const aggregatedDemand = filteredDemand.reduce((acc, order) => {
  const key = order.typeId;
  if (!acc[key]) {
    acc[key] = {
      typeId: order.typeId,
      typeName: order.typeName,
      totalQuantity: 0,
      maxPrice: 0,
      orderCount: 0,
      orders: [],
    };
  }
  acc[key].totalQuantity += order.quantityDesired;
  acc[key].maxPrice = Math.max(acc[key].maxPrice, order.maxPricePerUnit);
  acc[key].orderCount += 1;
  acc[key].orders.push(order);
  return acc;
}, {});
```

**Two Tables:**
1. **Aggregated Summary** - Totals per item type
2. **Individual Orders** - Detailed view

### EVE Images Utility

**File:** `frontend/packages/utils/eveImages.ts`

Centralized helper functions for EVE image URLs:

```typescript
export function getItemIconUrl(
  typeId: number,
  size: 32 | 64 | 128 | 256 | 512 = 64
): string {
  return `https://images.evetech.net/types/${typeId}/icon?size=${size}`;
}

export function getItemRenderUrl(
  typeId: number,
  size: 32 | 64 | 128 | 256 | 512 = 512
): string {
  return `https://images.evetech.net/types/${typeId}/render?size=${size}`;
}

// Also: getCharacterPortraitUrl, getCorporationLogoUrl, getAllianceLogoUrl
```

**Benefits:**
- Consistent URL generation
- Type-safe size parameters
- Reusable across components
- Easy to update if CDN changes

## API Routes (Next.js)

### Proxy Pattern

All frontend API routes proxy to the Go backend:

```typescript
// /api/buy-orders/index.ts
export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  if (req.method === "GET") {
    const response = await fetch(backend + "v1/buy-orders", {
      method: "GET",
      headers: getHeaders(session.providerAccountId),
    });

    if (response.status !== 200) {
      return res.status(response.status).json({ error: "Failed to get buy orders" });
    }

    const data = await response.json();
    return res.status(200).json(data);
  }

  // Handle POST...
}
```

**Files:**
- `/api/buy-orders/index.ts` - GET, POST
- `/api/buy-orders/[id].ts` - PUT, DELETE
- `/api/buy-orders/demand.ts` - GET
- `/api/item-types/search.ts` - GET

## Security & Authorization

### Permission Model

```
Buyer creates buy order
    ↓
Buyer grants "for_sale_browse" to Seller
    ↓
Seller can view buyer's buy orders in Demand tab
```

### Controller-Level Checks

```go
// Owner verification in Update
existingOrder, err := c.repository.GetByID(args.Request.Context(), id)
if existingOrder.BuyerUserID != *args.User {
    return nil, &web.HttpError{
        StatusCode: 403,
        Error:      errors.New("you do not own this buy order"),
    }
}
```

### Repository-Level Filtering

```sql
-- Delete: ownership verified by WHERE clause
UPDATE buy_orders
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND buyer_user_id = $2
```

## Performance Optimizations

### 1. Database Indexes

- `idx_buy_orders_buyer` - Fast user order queries
- `idx_buy_orders_type` - Fast type-based queries
- `idx_buy_orders_active` - Filter active orders efficiently

### 2. Connection Pooling

```go
// Test setup
db.SetMaxOpenConns(5)
db.SetMaxIdleConns(2)
```

Prevents PostgreSQL connection exhaustion.

### 3. Frontend Debouncing

```typescript
// 300ms debounce prevents API spam
searchTimeoutRef.current = setTimeout(() => {
  searchItems(value);
}, 300);
```

### 4. Query Optimization

- Use JOINs instead of separate queries
- LIMIT autocomplete results to 20
- SELECT only needed columns
- Use DISTINCT for permission queries

### 5. Caching (Future)

Potential caching targets:
- Item type search results
- User's buy orders (invalidate on CUD)
- Aggregated demand calculations

## Error Handling

### Repository Layer

```go
if err == sql.ErrNoRows {
    return nil, errors.New("buy order not found")
}
if err != nil {
    return nil, errors.Wrap(err, "failed to get buy order")
}
```

### Controller Layer

```go
if quantityDesired <= 0 {
    return nil, &web.HttpError{
        StatusCode: 400,
        Error:      errors.New("quantityDesired must be positive"),
    }
}
```

### Frontend Layer

```typescript
try {
  const response = await fetch('/api/buy-orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(formData)
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to save buy order');
  }

  showSnackbar('Buy order created successfully', 'success');
} catch (error: any) {
  showSnackbar(error.message || 'Failed to save buy order', 'error');
}
```

## Logging

Strategic logging for debugging and monitoring:

```go
log.Info("buy order created", "orderId", order.ID, "userId", *args.User, "typeId", req.TypeID)
log.Info("buy order updated", "orderId", id, "userId", *args.User)
log.Info("buy order deleted", "orderId", id, "userId", *args.User)
```

## Testing Infrastructure

### Test Database Isolation

Each test gets a unique database:

```go
databaseName := "testDatabase_" + strconv.Itoa(rand.Int())
```

### Connection Pooling in Tests

```go
db.SetMaxOpenConns(5)
db.SetMaxIdleConns(2)
```

### PostgreSQL Configuration

```yaml
command: postgres -c max_connections=200 -c shared_buffers=256MB
```

## Future Enhancements

### Phase 6: Sales Metrics Dashboard
- Sales analytics
- Revenue charts
- Top selling items
- Demand vs. inventory comparison

### Potential Features
- **Auto-Matching**: Notify when for-sale items match buy orders
- **Price Suggestions**: Recommend prices based on buy orders
- **Bulk Operations**: Create multiple buy orders at once
- **Expiration**: Auto-deactivate old buy orders
- **Quantity Tracking**: Track partially filled orders
- **Notifications**: Alert when new matching demand appears

## Design Patterns Used

1. **Repository Pattern**: Data access abstraction
2. **Controller Pattern**: HTTP request handling
3. **Proxy Pattern**: Frontend API routes
4. **Factory Pattern**: Repository/controller constructors
5. **Strategy Pattern**: Different sorting strategies
6. **Observer Pattern**: Permission-based filtering

## Key Learnings

1. **Relevance Sorting**: CASE expressions in SQL for custom ordering
2. **Debouncing**: Essential for autocomplete to reduce API calls
3. **Connection Pooling**: Critical for test stability
4. **Soft Deletes**: Preserve history while hiding records
5. **Permission Integration**: JOIN with permissions for authorization
6. **Icon URLs**: Centralized helpers prevent URL inconsistencies

## Related Documentation

- [API Documentation](API.md) - Detailed endpoint specs
- [Quick Start Guide](QUICK_START.md) - Getting started
- [Testing Guide](TESTING.md) - Test suite details
