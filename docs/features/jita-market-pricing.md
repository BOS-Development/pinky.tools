# Jita Market Pricing Feature

## Overview

The Jita Market Pricing feature integrates real-time market data from EVE Online's primary trade hub (Jita) into the industry-tool application, enabling players to:

- **Value their inventory** - See the ISK value of assets at current market prices
- **Calculate stockpile costs** - Determine how much ISK is needed to fill deficits
- **Make informed decisions** - Prioritize purchases based on actual market conditions

## Business Context

Industrial players in EVE Online need to understand the financial value of their assets and the cost to acquire missing materials. Jita (The Forge region, region_id: 10000002) is the game's largest trade hub where most pricing discovery occurs.

### User Stories

**As an industrial player, I want to:**
1. See the ISK value of my current inventory
2. Know how much it will cost to fill my stockpile deficits
3. Update market prices on-demand to get current pricing data

## Architecture

### Data Flow

```
ESI API (/markets/10000002/orders/)
    ↓
ESI Client (GetMarketOrders)
    ↓
Market Prices Updater (calculate best bid/ask)
    ↓
Market Prices Repository (database)
    ↓
Assets Repository (JOIN on type_id)
    ↓
Frontend (display ISK values)
```

### Key Design Decisions

#### 1. Data Source Selection

**Decision**: Use ESI `/markets/{region_id}/orders/` endpoint
- Returns ALL active market orders for a region
- Provides both buy and sell orders
- More accurate than universe-wide averages (`/markets/prices/`)

**Alternative Considered**: `/markets/prices/` endpoint
- Rejected: Provides universe-wide averages, not specific to Jita
- Rejected: Less granular, doesn't distinguish buy vs sell prices

#### 2. Storage Strategy

**Decision**: Store all Jita market data in database table
- Enables caching (reduces ESI API calls)
- Minimal performance impact (asset queries already have 7+ JOINs)
- Provides complete coverage for all items

**Alternative Considered**: Fetch prices on-demand per user
- Rejected: Too many ESI API calls
- Rejected: Slower user experience
- Rejected: Doesn't handle multiple users efficiently

#### 3. Price Types

**Decision**: Store both buy price (best bid) and sell price (best ask)

**Usage:**
- **Sell Price** → Asset valuation (what you could sell items for)
- **Buy Price** → Stockpile deficit cost (what you'd pay to acquire items)

**Rationale:**
- Assets represent inventory you could liquidate → use sell price
- Deficits represent items you need to buy → use buy price
- More accurate than using a single average price

#### 4. Update Strategy

**Decision**: Full market snapshot updates triggered manually
- Fetch ALL market orders on update (not per-item)
- Delete old data, insert new data (full refresh)
- Manual trigger initially, scheduled updates in future

**Alternative Considered**: Incremental updates per user's items
- Rejected: More complex to track which items need updating
- Rejected: Doesn't benefit other users
- Rejected: Harder to maintain data consistency

## Database Schema

### Table: `market_prices`

```sql
CREATE TABLE market_prices (
    type_id BIGINT PRIMARY KEY REFERENCES asset_item_types(type_id),
    region_id BIGINT NOT NULL,
    buy_price DOUBLE PRECISION,
    sell_price DOUBLE PRECISION,
    daily_volume BIGINT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_market_prices_region ON market_prices(region_id);
CREATE INDEX idx_market_prices_updated ON market_prices(updated_at);
```

### Schema Design Notes

- **Primary Key**: `type_id` (one price per item, scoped to Jita initially)
- **region_id**: Allows future multi-region support
- **buy_price/sell_price**: Both nullable (some items may only have buy OR sell orders)
- **daily_volume**: Useful for liquidity analysis (future feature)
- **updated_at**: Cache staleness tracking

### Migration Files

- **Up**: `internal/database/migrations/4_market_prices.up.sql`
- **Down**: `internal/database/migrations/4_market_prices.down.sql`

## Backend Implementation

### Components

#### 1. ESI Client Extension

**File**: `internal/client/esiClient.go`

**New Method**: `GetMarketOrders(ctx, regionID) ([]*MarketOrder, error)`
- Fetches ALL market orders for a region (paginated)
- Public endpoint (no OAuth required)
- Returns buy and sell orders

**New Struct**: `MarketOrder`
```go
type MarketOrder struct {
    TypeID       int64
    Price        float64
    IsBuyOrder   bool
    VolumeRemain int64
    // ... other fields
}
```

#### 2. Market Prices Repository

**File**: `internal/repositories/marketPrices.go` (NEW)

**Methods:**
- `UpsertPrices(ctx, []MarketPrice)` - Batch insert/update prices
- `DeleteAllForRegion(ctx, regionID)` - Clear old data before refresh
- `GetPricesForTypes(ctx, typeIDs, regionID)` - Fetch specific prices (future)

**Pattern**: Follows existing repository patterns (constructor, transaction-based)

#### 3. Market Prices Updater

**File**: `internal/updaters/marketPrices.go` (NEW)

**Main Method**: `UpdateJitaMarket(ctx) error`

**Algorithm:**
1. Fetch all market orders from ESI
2. Group orders by `type_id`
3. For each type, calculate:
   - `best_buy` = MAX(price) WHERE is_buy_order = true
   - `best_sell` = MIN(price) WHERE is_buy_order = false
   - `daily_volume` = SUM(volume_remain)
4. Delete old prices for region
5. Batch upsert new prices

#### 4. Assets Repository Modification

**File**: `internal/repositories/assets.go`

**Changes:**
1. Add fields to `Asset` struct:
   - `UnitPrice *float64` (sell price per unit)
   - `TotalValue *float64` (quantity × sell price)
   - `DeficitValue *float64` (deficit × buy price)

2. Add LEFT JOIN to all 4 asset queries:
   ```sql
   LEFT JOIN market_prices market ON (
       market.type_id = assets.type_id
       AND market.region_id = 10000002
   )
   ```

3. Add calculated fields to SELECT:
   ```sql
   market.sell_price as unit_price,
   (assets.quantity * COALESCE(market.sell_price, 0)) as total_value,
   CASE
       WHEN (assets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
       THEN ABS(...) * COALESCE(market.buy_price, 0)
       ELSE 0
   END as deficit_value
   ```

#### 5. Market Prices Controller

**File**: `internal/controllers/marketPrices.go` (NEW)

**Endpoint**: `POST /v1/market-prices/update`
- Triggers full Jita market data refresh
- User-agnostic (updates all prices)
- Returns 200 OK or error

### Pricing Logic

| Field | Formula | Purpose |
|-------|---------|---------|
| `unit_price` | `sell_price` | What one unit is worth |
| `total_value` | `quantity × sell_price` | Total sellable value |
| `deficit_value` | `abs(deficit) × buy_price` | Cost to fill deficit |

**Example:**
- You have 1,000 Tritanium (current)
- You want 2,000 Tritanium (desired)
- Deficit: -1,000 Tritanium
- Buy price: 5.50 ISK
- Sell price: 5.45 ISK

**Results:**
- `unit_price`: 5.45 ISK (you could sell at this price)
- `total_value`: 1,000 × 5.45 = 5,450 ISK
- `deficit_value`: 1,000 × 5.50 = 5,500 ISK (cost to buy missing units)

## Frontend Implementation

### Changes

#### 1. TypeScript Models

**File**: `frontend/packages/client/data/models.ts`

```typescript
export type Asset = {
  // ... existing fields
  unitPrice?: number;        // NEW
  totalValue?: number;       // NEW
  deficitValue?: number;     // NEW
};
```

#### 2. Assets List Component

**File**: `frontend/packages/components/assets/AssetsList.tsx`

**New Columns:**
- **Unit Price** - Display with 2 decimal places + "ISK" suffix
- **Total Value** - Display with bold font, 0 decimal places
- **Deficit Cost** - Display in red (error color) if > 0

**New Summary Cards:**
- **Total Value** - Sum of all `totalValue` (green icon)
- **Deficit Cost** - Sum of all `deficitValue` (red warning icon)

**New Action:**
- **Refresh Prices** button - Calls `/v1/market-prices/update` endpoint

### UI Mockup

```
┌─────────────────────────────────────────────────────────────┐
│  Assets                                           [Refresh]  │
├─────────────────────────────────────────────────────────────┤
│  📦 1,234 Items  │  📊 89 Types  │  💰 12.5M ISK  │  ⚠️ 3.2M  │
├─────────────────────────────────────────────────────────────┤
│ Name          Qty      Volume   Unit Price  Total Value  Deficit│
│ Tritanium     1,000    10 m³    5.45 ISK    5,450 ISK   5,500   │
│ Pyerite       500      5 m³     12.30 ISK   6,150 ISK   -       │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Database & Models ✅
- [x] Create migrations
- [x] Add `MarketPrice` model
- [x] Add `MarketOrder` struct

### Phase 2: ESI Integration ✅
- [x] Implement `GetMarketOrders()`
- [x] Write ESI client tests
- [x] Create market prices repository
- [x] Write repository tests

### Phase 3: Update Logic ✅
- [x] Create market prices updater
- [x] Implement price calculation (best bid/ask)
- [x] Write updater tests
- [x] Add time-based update throttling (6 hours)

### Phase 4: Backend Integration ✅
- [x] Modify `Asset` struct
- [x] Add LEFT JOIN to asset queries
- [x] Update Scan() calls
- [x] Test asset retrieval includes prices

### Phase 5: API Layer ✅
- [x] Create market prices controller
- [x] Wire dependencies
- [x] Write controller tests

### Phase 6: Frontend (Pending)
- [ ] Update Asset TypeScript type
- [ ] Add ISK columns to table
- [ ] Add summary statistics
- [ ] Add refresh button
- [ ] Test UI rendering

## Testing Strategy

### Unit Tests

#### Updater Tests (`internal/updaters/marketPrices_test.go`)
- ✅ Basic update flow (fetch, calculate, upsert)
- ✅ Skip recent updates (time throttling)
- ✅ ESI client errors
- ✅ Delete operation errors
- ✅ Upsert operation errors
- ✅ Empty orders handling
- ✅ Buy-only orders
- ✅ Sell-only orders
- ✅ Multiple prices (select best)
- ✅ GetLastUpdateTime errors

#### Controller Tests (`internal/controllers/marketPrices_test.go`)
- ✅ Successful update
- ✅ Updater errors
- ✅ Network errors
- ✅ Route registration
- ✅ Context propagation

#### Repository Tests (`internal/repositories/marketPrices_test.go`)
- ✅ Batch upsert
- ✅ Delete by region
- ✅ Get last update time
- ✅ Transaction rollback on error

### Integration Tests

**Manual Testing:**
```bash
# 1. Trigger market update
curl -X POST http://localhost:8080/v1/market-prices/update \
  -H "Cookie: session=..."

# 2. Verify assets include prices
curl http://localhost:8080/v1/assets/ \
  -H "Cookie: session=..." | jq '.structures[0].hangarAssets[0]'

# 3. Check database
psql -c "SELECT COUNT(*) FROM market_prices WHERE region_id = 10000002;"
```

## Performance Considerations

### Database Performance

**Query Impact:**
- Added one LEFT JOIN to asset queries
- Asset queries already have 7+ JOINs
- Indexes on `type_id` (PK) and `region_id`
- Expected impact: < 50ms additional latency

**Optimization:**
- Batch upsert (not individual inserts)
- Transaction-based operations
- Index on `updated_at` for cache staleness checks

### API Performance

**ESI Rate Limits:**
- Public endpoint: 150 requests/second burst
- Pagination: ~20-30 pages for full Jita market
- Full refresh: ~2-5 seconds total

**Caching Strategy:**
- Update throttling: 6 hour minimum between updates
- Manual trigger allows users to force refresh when needed
- Future: Background job for automatic scheduled updates

### Expected Metrics

| Operation | Expected Time |
|-----------|--------------|
| Full market fetch (ESI) | 2-5 seconds |
| Price calculation | < 1 second |
| Database upsert | 1-2 seconds |
| Asset query with prices | < 500ms |
| **Total update time** | **3-8 seconds** |

## Security Considerations

### Authentication

- Update endpoint requires user authentication (`web.AuthAccessUser`)
- Public ESI endpoint (no credentials needed)
- No sensitive data in market prices

### Data Validation

- Type checking on ESI responses
- Null price handling (some items may lack buy/sell orders)
- SQL injection prevention (parameterized queries)

### Rate Limiting

- ESI client respects rate limits
- Update throttling prevents excessive calls
- User-initiated updates only (no automatic spam)

## Future Enhancements

### Multi-Region Support

**Current**: Jita only (region_id: 10000002)

**Future**: Support multiple trade hubs
- Add region selector in UI
- Store prices for multiple regions
- Update schema: composite key `(type_id, region_id)`

### Scheduled Updates

**Current**: Manual trigger via API

**Future**: Background job runner
- Automatic updates every 6 hours
- Use market prices runner (`internal/runners/marketPrices.go`)
- Configurable interval

### Price History

**Current**: Single snapshot (latest prices)

**Future**: Historical tracking
- New table: `market_price_history`
- Track price trends over time
- Enable price charts in UI

### Price Alerts

**Future**: Notify when prices cross thresholds
- "Alert me when Tritanium < 5.00 ISK"
- "Alert me when deficit cost > 10M ISK"
- Email or in-app notifications

### Liquidity Analysis

**Current**: Store `daily_volume` but don't display

**Future**: Show market liquidity
- "Low liquidity" warning for illiquid items
- Volume-weighted average prices
- Market depth analysis

## Monitoring & Observability

### Logs

All operations logged with structured logging:
```go
log.Info("updating market prices", "region_id", regionID)
log.Error("failed to fetch market orders", "error", err)
log.Info("market prices updated successfully", "items_count", len(prices))
```

### Metrics to Track

- Market update frequency
- Update duration (ESI fetch time)
- Number of items with prices
- Failed update attempts
- User-triggered vs scheduled updates

### Health Checks

- Last successful update timestamp
- Stale data detection (> 12 hours old)
- ESI availability monitoring

## API Reference

### Endpoints

#### Update Market Prices

```http
POST /v1/market-prices/update
Authorization: Required (cookie-based session)
```

**Response:**
- `200 OK` - Update successful
- `500 Internal Server Error` - Update failed

**Behavior:**
- Fetches ALL Jita market orders from ESI
- Calculates best bid/ask for each item type
- Deletes old prices and inserts new prices
- Throttled: skips if updated within last 6 hours

## Troubleshooting

### Common Issues

**Issue**: Prices show as blank (-) in UI
- **Cause**: No market orders for that item in Jita
- **Solution**: Expected behavior for rarely-traded items

**Issue**: Update takes a long time
- **Cause**: ESI API latency or many pages of orders
- **Solution**: Normal for full market refresh (3-8 seconds)

**Issue**: Prices seem outdated
- **Cause**: Last update was > 6 hours ago
- **Solution**: Click refresh button in UI

**Issue**: Update fails with ESI error
- **Cause**: ESI API downtime or rate limiting
- **Solution**: Wait and retry; check ESI status page

### Database Queries for Debugging

```sql
-- Check when prices were last updated
SELECT MAX(updated_at) FROM market_prices WHERE region_id = 10000002;

-- Count items with prices
SELECT COUNT(*) FROM market_prices WHERE region_id = 10000002;

-- Find items with highest sell prices
SELECT type_id, sell_price FROM market_prices
WHERE region_id = 10000002
ORDER BY sell_price DESC LIMIT 10;

-- Find items with no buy orders
SELECT type_id, sell_price FROM market_prices
WHERE region_id = 10000002 AND buy_price IS NULL;
```

## References

- **EVE Online ESI Documentation**: https://esi.evetech.net/ui/
- **Market Orders Endpoint**: `GET /markets/{region_id}/orders/`
- **Jita Region ID**: 10000002 (The Forge)

## Files Modified/Created

### Backend
- ✅ `internal/database/migrations/4_market_prices.up.sql` - Schema creation
- ✅ `internal/database/migrations/4_market_prices.down.sql` - Schema teardown
- ✅ `internal/models/models.go` - MarketPrice model
- ✅ `internal/client/esiClient.go` - GetMarketOrders() method + MarketOrder struct
- ✅ `internal/repositories/marketPrices.go` - NEW repository
- ✅ `internal/updaters/marketPrices.go` - NEW updater
- ✅ `internal/runners/marketPrices.go` - NEW runner (scheduled updates)
- ✅ `internal/controllers/marketPrices.go` - NEW controller
- ✅ `internal/repositories/assets.go` - Modified Asset struct + queries
- ✅ `cmd/industry-tool/cmd/root.go` - Dependency wiring

### Tests
- ✅ `internal/updaters/marketPrices_test.go` - Updater tests (10 tests)
- ✅ `internal/controllers/marketPrices_test.go` - Controller tests (5 tests)
- ✅ `internal/runners/marketPrices_test.go` - Runner tests (8 tests)
- ✅ `internal/repositories/marketPrices_test.go` - Repository tests

### Frontend (Pending)
- ⏳ `frontend/packages/client/data/models.ts` - Asset type update
- ⏳ `frontend/packages/components/assets/AssetsList.tsx` - UI updates

---

**Status**: Backend Complete ✅ | Frontend Pending ⏳
**Last Updated**: 2026-02-14
**Version**: 1.0
