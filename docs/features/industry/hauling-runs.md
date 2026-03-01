# Hauling Runs

## Status

- **Phase**: 1 - Market Scanner & Run CRUD (Implemented)
- **Scope**: Hub-to-hub arbitrage scanning, run creation/tracking, per-item fill progress
- **Future**: Phase 2 adds Discord alerts, split-run operators, corp wallet buy orders, P&L UI

## Overview

Hauling Runs is a logistics planning feature that helps players identify profitable item-hauling opportunities between trade hubs and track acquisition progress. Phase 1 provides a market scanner that identifies arbitrage spreads and a run planner to organize items by destination, fill capacity, and compute net profit per haul.

Phase 1 focuses on the planning workflow: scanning markets for opportunities, creating hauling runs with source/destination regions, adding items to runs with target quantities, and tracking fill progress toward ship capacity. Phase 2 will add Discord notifications for fill alerts and post-run P&L tracking.

## Key Decisions

1. **Hub-to-hub arbitrage model** — Scanner compares source region buy prices vs destination region buy orders to identify profitable flips. Spreads are tiered: gap (>15%), markup (5-15%), thin (<5%).

2. **Market snapshot caching** — ESI market order requests are cached per region/system for 30 minutes to reduce API calls and improve UI responsiveness. Snapshots store buy_price, sell_price, volume_available, avg_daily_volume, and estimated days_to_sell.

3. **Three regions per run** — Runs track source_region_id, to_region_id, and optional from_system_id (null = region-wide scan). This allows zone-based planning (Jita to Amarr) and precision system-level pickups.

4. **Capacity tracking** — Runs store max_volume_m3. Per-item acquisition is tracked; overall fill percentage is computed across all run items. Ship capacity acts as the hard constraint for adding more items.

5. **Fill progress per item** — Each run item tracks quantity_planned vs quantity_acquired. Frontend displays both absolute and percentage fill, supporting partial-fill workflows (acquire items as available).

6. **Net profit calculation** — Per-item profit = (sell_price - buy_price) × quantity_acquired, summed across all items. Price inputs are user-controlled fields, allowing manual price entry or Jita snapshot integration.

7. **Run states** — PLANNING (initial), ACCUMULATING (acquiring items), then READY, IN_TRANSIT, SELLING, COMPLETE, CANCELLED. Phase 1 supports PLANNING and ACCUMULATING; later phases unlock remaining states.

8. **Character tracking** — hauling_run_items.character_id optionally tracks which character placed the buy order (for corp-level runs where multiple chars place orders).

9. **Notification schema ready** — notify_tier2, notify_tier3, daily_digest flags stored on run; alerting logic deferred to Phase 2.

10. **P&L scaffold** — hauling_run_pnl table exists for future post-run accounting; Phase 1 does not populate it.

## Schema

### `hauling_runs` (NEW)

Represents a planned hauling operation between regions with fill tracking.

- `id` (bigint, PK)
- `user_id` (bigint, FK users) — owner of the run
- `name` (varchar(255)) — user-friendly name (e.g., "Jita to Amarr T2")
- `status` (varchar) — enum: PLANNING, ACCUMULATING, READY, IN_TRANSIT, SELLING, COMPLETE, CANCELLED
- `from_region_id` (bigint) — source region
- `from_system_id` (bigint, nullable) — optional: scope scan to specific system (null = entire from_region)
- `to_region_id` (bigint) — destination region
- `max_volume_m3` (numeric) — ship cargo capacity m³
- `haul_threshold_isk` (numeric, nullable) — minimum net profit to trigger alert (Phase 2)
- `notify_tier2` (boolean, default false) — notify on Tier 2 fill (Phase 2)
- `notify_tier3` (boolean, default false) — notify on Tier 3 fill (Phase 2)
- `daily_digest` (boolean, default false) — include in daily digest (Phase 2)
- `notes` (text, nullable) — user notes
- `created_at`, `updated_at` (timestamps)

**Indexes:**
- `(user_id, status)` — list user's runs
- CHECK constraint on status values

### `hauling_run_items` (NEW)

Per-item tracking within a hauling run.

- `id` (bigint, PK)
- `run_id` (bigint, FK hauling_runs CASCADE) — parent run
- `type_id` (bigint) — EVE item type ID
- `type_name` (varchar(255)) — cached item name for display
- `quantity_planned` (bigint) — target acquisition quantity
- `quantity_acquired` (bigint) — current acquired quantity
- `buy_price_isk` (numeric(12,2)) — buy price per unit (source)
- `sell_price_isk` (numeric(12,2)) — sell price per unit (destination)
- `volume_m3` (numeric) — volume per unit from SDE
- `character_id` (bigint, nullable, FK characters) — which character placed buy order
- `notes` (text, nullable) — per-item notes
- `created_at`, `updated_at` (timestamps)

**Indexes:**
- UNIQUE `(run_id, type_id)` — prevent duplicates per run
- FK: run_id (cascade), character_id

**Computed fields (application level):**
- fill_percentage = (quantity_acquired / quantity_planned) × 100
- net_profit_isk = (sell_price_isk - buy_price_isk) × quantity_acquired
- total_volume_m3 = volume_m3 × quantity_acquired

### `hauling_market_snapshots` (NEW)

Cached market order data per type/region/system for fast scanner results.

- `type_id` (bigint)
- `region_id` (bigint)
- `system_id` (bigint) — 0 = region-wide average; >0 = specific system
- `buy_price` (numeric(12,2)) — highest buy order price
- `sell_price` (numeric(12,2)) — lowest sell order price
- `volume_available` (bigint) — units available at best sell price
- `avg_daily_volume` (bigint) — market liquidity estimate
- `days_to_sell` (numeric(5,1)) — estimated time to sell at sell_price
- `updated_at` (timestamp)

**Primary Key:** `(type_id, region_id, system_id)`

**Cache policy:**
- Records older than 30 minutes are candidates for refresh on next scanner request
- Full refresh: all items in source region + selected items in destination
- Partial refresh: specific items on user request (e.g., after manual price update)

### `hauling_run_pnl` (NEW, SCAFFOLD ONLY)

Post-run P&L tracking (Phase 2 UI, Phase 1 schema ready).

- `id` (bigint, PK)
- `run_id` (bigint, FK hauling_runs CASCADE, UNIQUE)
- `type_id` (bigint)
- `quantity_sold` (bigint) — units actually sold
- `avg_sell_price_isk` (numeric(12,2)) — average price per unit sold
- `total_revenue_isk` (numeric(14,2)) — quantity_sold × avg_sell_price
- `total_cost_isk` (numeric(14,2)) — cost to acquire (sum of buys)
- `net_profit_isk` (numeric(14,2)) — GENERATED ALWAYS as total_revenue - total_cost STORED

## API Endpoints

### Run Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/hauling/runs` | List user's runs with fill percentages |
| POST | `/v1/hauling/runs` | Create new run |
| GET | `/v1/hauling/runs/{id}` | Get run with all items and fill stats |
| PUT | `/v1/hauling/runs/{id}` | Update run (name, status, notification prefs) |
| DELETE | `/v1/hauling/runs/{id}` | Delete run and cascade items |
| PUT | `/v1/hauling/runs/{id}/status` | Update run status only |

**POST/PUT body:**
```json
{
  "name": "Jita to Amarr",
  "from_region_id": 10000002,
  "from_system_id": 30002187,
  "to_region_id": 10000043,
  "max_volume_m3": 500000,
  "haul_threshold_isk": 50000000,
  "notify_tier2": true,
  "notify_tier3": false,
  "daily_digest": false,
  "notes": "T2 components focus"
}
```

**Response (GET /runs/{id}):**
```json
{
  "id": 1,
  "name": "Jita to Amarr",
  "status": "ACCUMULATING",
  "from_region_id": 10000002,
  "from_system_id": 30002187,
  "to_region_id": 10000043,
  "max_volume_m3": 500000,
  "used_volume_m3": 125000,
  "fill_percentage": 25.0,
  "total_net_profit_isk": 500000000,
  "items": [
    {
      "id": 10,
      "type_id": 1234,
      "type_name": "Fusion Reactor",
      "quantity_planned": 100,
      "quantity_acquired": 75,
      "fill_percentage": 75.0,
      "buy_price_isk": 2500000,
      "sell_price_isk": 2750000,
      "volume_m3": 50,
      "net_profit_isk": 18750000,
      "character_id": null,
      "notes": "Limited supply"
    }
  ],
  "created_at": "2026-03-01T10:30:00Z",
  "updated_at": "2026-03-01T14:45:00Z"
}
```

### Run Items

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/hauling/runs/{id}/items` | Add item to run |
| PUT | `/v1/hauling/runs/{id}/items/{itemId}` | Update acquired quantity and prices |
| DELETE | `/v1/hauling/runs/{id}/items/{itemId}` | Remove item from run |

**POST body:**
```json
{
  "type_id": 1234,
  "type_name": "Fusion Reactor",
  "quantity_planned": 100,
  "buy_price_isk": 2500000,
  "sell_price_isk": 2750000,
  "character_id": null,
  "notes": "Limited supply"
}
```

**PUT body (update acquired quantity):**
```json
{
  "quantity_acquired": 75,
  "buy_price_isk": 2480000,
  "sell_price_isk": 2770000

}
```

### Market Scanner

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/hauling/scanner` | Get arbitrage opportunities with sorted results |
| POST | `/v1/hauling/scanner/scan` | Trigger full market scan (async background job) |

**Query params (GET /scanner):**
- `source_region_id` (required) — scan source region
- `dest_region_id` (required) — scan destination region
- `source_system_id` (optional) — narrow source to specific system
- `sort_by` (optional, default: net_profit_desc) — net_profit_desc, net_profit_asc, spread, volume, days_to_sell
- `min_spread_pct` (optional, default: 0) — filter results by minimum spread percentage
- `page` (optional, default: 1)
- `limit` (optional, default: 50)

**Response:**
```json
{
  "results": [
    {
      "type_id": 1234,
      "type_name": "Fusion Reactor",
      "source_buy_price": 2500000,
      "dest_sell_price": 2750000,
      "spread_pct": 10.0,
      "spread_indicator": "markup",
      "volume_available_at_source": 1000,
      "avg_daily_volume": 50,
      "days_to_sell": 20.0,
      "net_profit_per_unit": 250000,
      "estimated_profit_1_unit": 250000,
      "estimated_profit_10_units": 2500000,
      "estimated_profit_100_units": 25000000
    }
  ],
  "total_count": 1523,
  "page": 1,
  "per_page": 50
}
```

**Spread Indicators:**
- `gap` — spread > 15% (excellent opportunity)
- `markup` — 5% ≤ spread ≤ 15% (good opportunity)
- `thin` — spread < 5% (marginal)

**POST /scanner/scan body (optional filters):**
```json
{
  "source_region_id": 10000002,
  "dest_region_id": 10000043,
  "source_system_id": 30002187,
  "item_type_ids": [1234, 5678]
}
```

Triggers background market fetch; returns job_id for polling.

### Fill Remaining Capacity Suggestions

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/hauling/runs/{id}/fill-suggestions` | Top 3 arbitrage fits for remaining capacity |

**Response:**
```json
{
  "remaining_volume_m3": 375000,
  "suggestions": [
    {
      "type_id": 5678,
      "type_name": "T2 Reactor",
      "source_buy_price": 1200000,
      "dest_sell_price": 1500000,
      "isk_per_m3": 25000,
      "units_fit": 312,
      "net_profit_full_fit": 93600000,
      "volume_m3": 50
    }
  ]
}
```

Frontend [Add] button on suggestion posts to `/v1/hauling/runs/{id}/items` with populated type_id, type_name, quantity_planned = units_fit, buy_price, sell_price.

## File Structure

### Backend

**Migrations:**
- `internal/database/migrations/20260301010649_create_hauling_runs.up.sql`
- `internal/database/migrations/20260301010649_create_hauling_runs.down.sql`

**Models:**
- `internal/models/models.go` — HaulingRun, HaulingRunItem, HaulingMarketSnapshot, HaulingArbitrageRow (scanner result DTO)

**Repositories:**
- `internal/repositories/haulingRuns.go` — CRUD: CreateRun, GetRun, UpdateRun, DeleteRun, ListRuns, UpdateStatus
- `internal/repositories/haulingRunItems.go` — CRUD: AddItem, UpdateItem, DeleteItem, GetRunItems
- `internal/repositories/haulingMarket.go` — Cache: UpsertSnapshot, GetSnapshot, GetSnapshotsForType, GetSnapshotsOlderThan

**Updaters:**
- `internal/updaters/haulingMarket.go` — Market scanning and snapshot refresh:
  - `FetchMarketOrdersForRegion(regionID, systemID)` — ESI market orders
  - `ComputeArbitrage(sourceRegion, destRegion)` — price comparison, spread tiering
  - `RefreshExpiredSnapshots()` — background job (runs hourly)

**Controllers:**
- `internal/controllers/haulingRuns.go` — HTTP handlers (11 endpoints):
  - ListRuns, CreateRun, GetRun, UpdateRun, DeleteRun, UpdateStatus
  - GetRunItems, AddRunItem, UpdateRunItem, DeleteRunItem
  - ScanMarket, GetScanResults, GetFillSuggestions

**Client:**
- `internal/client/esiClient.go` — added MarketHistoryEntry type and GetMarketHistory method for future analytics

### Frontend

**Pages:**
- `frontend/pages/hauling.tsx` — Page router entry
- `frontend/pages/api/hauling/runs.ts` — GET/POST runs
- `frontend/pages/api/hauling/run.ts` — GET/PUT/DELETE specific run
- `frontend/pages/api/hauling/run-items.ts` — POST/PUT/DELETE items
- `frontend/pages/api/hauling/scanner.ts` — GET scan results
- `frontend/pages/api/hauling/scan.ts` — POST to trigger scan
- `frontend/pages/api/hauling/suggestions.ts` — GET fill suggestions

**Package Components:**
- `frontend/packages/pages/hauling.tsx` — Main page with tabs
- `frontend/packages/pages/hauling/detail.tsx` — Run detail view
- `frontend/packages/pages/hauling/scanner.tsx` — Scanner with sort/filter
- `frontend/packages/components/hauling/HaulingRunsList.tsx` — List view with creation
- `frontend/packages/components/hauling/HaulingRunDetail.tsx` — Detail editor + item list
- `frontend/packages/components/hauling/MarketScanner.tsx` — Scanner with result table
- `frontend/packages/components/hauling/FillSuggestionsPanel.tsx` — Top 3 fits + [Add] buttons
- `frontend/packages/client/api.ts` — API client methods for all endpoints

**Navigation:**
- `frontend/packages/components/Navbar.tsx` — Added "Hauling Runs" link to Industry menu

### Tests

**Backend Tests:**
- `internal/repositories/haulingRuns_test.go` — CRUD, list, status transitions
- `internal/repositories/haulingRunItems_test.go` — item CRUD, fill percentage
- `internal/repositories/haulingMarket_test.go` — cache, snapshot upsert, expiry
- `internal/updaters/haulingMarket_test.go` — arbitrage calculation, spread tiering
- `internal/controllers/haulingRuns_test.go` — endpoint handlers, error cases

**E2E Tests:**
- `e2e/tests/18-hauling-runs.spec.ts` — Full workflow:
  - Create run, add items
  - Update acquired quantities
  - Verify fill percentages and net profit
  - Market scanner flow
  - Fill suggestions [Add] integration

**Mock ESI:**
- `cmd/mock-esi/main.go` — Added market history endpoint for scanner caching validation

## Phase 2 Deferred Items

1. **Acquisition alerts** — Discord Tier 1/2/3 notifications when run hits haul_threshold_isk or tier2/tier3 fill flags. Schema ready; alerting logic deferred.

2. **Corp wallet buy order polling** — Integrate `esi-corporations.read_orders.v1` scope to auto-detect buy orders placed by corp characters and auto-fill quantities.

3. **Split-run operators** — Track which character is responsible for final haul in multi-char runs (operator_character_id on run). Phase 1 treats all runs as single-character.

4. **Post-run P&L UI** — Dashboard and detail views for hauling_run_pnl table. Manual or auto-entry of sold quantities and final prices.

5. **Undercut monitoring** — Alert if destination sell price undercuts by >X% after initial scan.

6. **Dotlan integration** — Fetch kill data and player traffic statistics to highlight risky routes.

## Open Questions

- [ ] Should "Fill Remaining Capacity" be a dialog modal or panel on run detail? Current design uses inline panel.
- [ ] Should Phase 2 auto-populate acquired quantities from corp wallet orders, or require manual confirmation?
- [ ] Should P&L tracking support multi-item post-run summaries, or one entry per type_id?
- [ ] Should haul_threshold_isk alert at run creation or at acquisition milestones (e.g., 50% filled)?
