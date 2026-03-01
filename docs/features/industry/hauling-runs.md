# Hauling Runs

## Status

- **Phase**: 2 - Discord Alerts, Corp Orders, P&L UI (Implemented)
- **Scope**: Hub-to-hub arbitrage scanning, run creation/tracking, per-item fill progress, Discord notifications, corp order auto-fill, post-run P&L tracking
- **Future**: Phase 3 adds split-run operators, undercut monitoring, Dotlan integration

## Overview

Hauling Runs is a logistics planning feature that helps players identify profitable item-hauling opportunities between trade hubs and track acquisition progress. The system provides a market scanner that identifies arbitrage spreads, a run planner to organize items by destination and fill capacity, and post-run P&L tracking.

**Phase 1** covers the planning workflow: scanning markets for opportunities, creating hauling runs with source/destination regions, adding items to runs with target quantities, and tracking fill progress toward ship capacity.

**Phase 2** adds Discord notifications (Tier 2 fill alerts at ≥80%, run completion alerts, daily digests), automatic corp buy order polling (matching corp wallet orders to run items every 15 minutes), and post-run P&L tracking with revenue, cost, and profit summary.

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

11. **Corp order auto-fill** — Corp buy orders are polled every 15 minutes (esi-corporations.read_orders.v1 scope) and matched to run items by type_id; quantity_acquired is updated automatically when corp orders are partially filled.

12. **Post-run P&L is per-item** — hauling_run_pnl records one entry per (run_id, type_id) with UNIQUE constraint. Total summary is computed as SQL aggregate in PnL summary endpoint.

13. **Tier 2 notification fires once per fill crossing** — The 80% threshold check in UpdateItemAcquired fires a goroutine to send Discord alert; duplicate firing is acceptable (Discord dedup not implemented).

14. **Route safety is frontend-only** — Jump count and kill count data are fetched client-side from public ESI and zKillboard APIs; no server-side caching.

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

### `hauling_run_pnl` (NEW, IMPLEMENTED IN PHASE 2)

Post-run P&L tracking — one entry per (run_id, type_id) pair.

- `id` (bigint, PK)
- `run_id` (bigint, FK hauling_runs CASCADE)
- `type_id` (bigint)
- `quantity_sold` (bigint) — units actually sold
- `avg_sell_price_isk` (numeric(12,2)) — average price per unit sold
- `total_revenue_isk` (numeric(14,2)) — quantity_sold × avg_sell_price
- `total_cost_isk` (numeric(14,2)) — cost to acquire (sum of buys)
- `net_profit_isk` (numeric(14,2)) — GENERATED ALWAYS as total_revenue - total_cost STORED

**Unique constraint:** `(run_id, type_id)` — one P&L record per item type per run

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

### P&L Tracking (Phase 2)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/hauling/runs/{id}/pnl` | Get P&L entries for run (all items) |
| PUT | `/v1/hauling/runs/{id}/pnl` | Upsert P&L entry (quantity_sold, avg_sell_price, total_revenue, total_cost) |
| GET | `/v1/hauling/runs/{id}/pnl/summary` | Get P&L summary (total revenue, cost, profit) |

**PUT /runs/{id}/pnl body:**
```json
{
  "type_id": 1234,
  "quantity_sold": 100,
  "avg_sell_price_isk": 2750000,
  "total_revenue_isk": 275000000,
  "total_cost_isk": 250000000
}
```

**GET /runs/{id}/pnl response:**
```json
[
  {
    "id": 100,
    "run_id": 1,
    "type_id": 1234,
    "quantity_sold": 100,
    "avg_sell_price_isk": 2750000,
    "total_revenue_isk": 275000000,
    "total_cost_isk": 250000000,
    "net_profit_isk": 25000000
  }
]
```

**GET /runs/{id}/pnl/summary response:**
```json
{
  "run_id": 1,
  "total_revenue_isk": 550000000,
  "total_cost_isk": 500000000,
  "total_profit_isk": 50000000,
  "item_count": 2
}
```

### Discord Notifications & Polling (Phase 2)

| Method | Path | Description |
|--------|------|-------------|
| N/A | Tier 2 Fill Alert | Fires when item reaches ≥80% fill and notify_tier2=true |
| N/A | Run Completion Alert | Fires when run status changes to COMPLETE |
| N/A | Daily Digest | Sent once per day to users with runs where daily_digest=true |
| N/A | Corp Order Poll | Background job every 15 minutes: fetches corp buy orders and matches to run items |

**Triggering Conditions:**
- **Tier 2 notification** — item.quantity_acquired / item.quantity_planned ≥ 0.8 AND run.notify_tier2=true
- **Run completion** — run.status changes from any state to COMPLETE
- **Daily digest** — runs with daily_digest=true, sent at configured time
- **Corp order match** — run.corporation_id orders matched by type_id, quantity_acquired auto-updated

## File Structure

### Backend

**Migrations:**
- `internal/database/migrations/20260301010649_create_hauling_runs.up.sql` — Phase 1: runs, items, snapshots
- `internal/database/migrations/20260301010649_create_hauling_runs.down.sql`
- `internal/database/migrations/20260301110717_hauling_run_pnl_unique.up.sql` — Phase 2: adds UNIQUE(run_id, type_id) constraint

**Models:**
- `internal/models/models.go` — HaulingRun, HaulingRunItem, HaulingMarketSnapshot, HaulingArbitrageRow (scanner result DTO)

**Repositories:**
- `internal/repositories/haulingRuns.go` — CRUD: CreateRun, GetRun, UpdateRun, DeleteRun, ListRuns, UpdateStatus; Phase 2 adds ListAccumulatingByUser, ListDigestRunsByUser
- `internal/repositories/haulingRunItems.go` — CRUD: AddItem, UpdateItem, DeleteItem, GetRunItems
- `internal/repositories/haulingMarket.go` — Cache: UpsertSnapshot, GetSnapshot, GetSnapshotsForType, GetSnapshotsOlderThan
- `internal/repositories/haulingRunPnl.go` — Phase 2: UpsertPnlEntry, GetPnlByRunID, GetPnlSummaryByRunID

**Updaters:**
- `internal/updaters/haulingMarket.go` — Market scanning and snapshot refresh:
  - `FetchMarketOrdersForRegion(regionID, systemID)` — ESI market orders
  - `ComputeArbitrage(sourceRegion, destRegion)` — price comparison, spread tiering
  - `RefreshExpiredSnapshots()` — background job (runs hourly)
- `internal/updaters/haulingNotifications.go` — Phase 2: Discord alerts for Tier 2 fill (≥80%), run completion, daily digest
- `internal/updaters/haulingCorpOrders.go` — Phase 2: matches corp buy orders to run items by type_id, updates quantity_acquired

**Controllers:**
- `internal/controllers/haulingRuns.go` — HTTP handlers (14+ endpoints):
  - Phase 1: ListRuns, CreateRun, GetRun, UpdateRun, DeleteRun, UpdateStatus
  - Phase 1: GetRunItems, AddRunItem, UpdateRunItem, DeleteRunItem
  - Phase 1: ScanMarket, GetScanResults, GetFillSuggestions
  - Phase 2: GetPnlByRunID, UpsertPnlEntry, GetPnlSummaryByRunID

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
- `frontend/pages/api/hauling/pnl.ts` — Phase 2: GET/PUT P&L entries
- `frontend/pages/api/hauling/pnl-summary.ts` — Phase 2: GET P&L summary

**Package Components:**
- `frontend/packages/pages/hauling.tsx` — Main page with tabs
- `frontend/packages/pages/hauling/detail.tsx` — Run detail view
- `frontend/packages/pages/hauling/scanner.tsx` — Scanner with sort/filter
- `frontend/packages/components/hauling/HaulingRunsList.tsx` — List view with creation
- `frontend/packages/components/hauling/HaulingRunDetail.tsx` — Detail editor + item list; Phase 2 adds P&L section and Route Safety card
- `frontend/packages/components/hauling/MarketScanner.tsx` — Scanner with result table
- `frontend/packages/components/hauling/FillSuggestionsPanel.tsx` — Top 3 fits + [Add] buttons
- `frontend/packages/components/hauling/HaulingRunPnlSection.tsx` — Phase 2: P&L entry form and summary display
- `frontend/packages/components/hauling/RouteSafetyCard.tsx` — Phase 2: Jump count + kill count from ESI and zKillboard
- `frontend/packages/client/api.ts` — API client methods for all endpoints; Phase 2 adds getPnl, upsertPnl, getPnlSummary

**Navigation:**
- `frontend/packages/components/Navbar.tsx` — Added "Hauling Runs" link to Industry menu

### Tests

**Backend Tests:**
- `internal/repositories/haulingRuns_test.go` — CRUD, list, status transitions
- `internal/repositories/haulingRunItems_test.go` — item CRUD, fill percentage
- `internal/repositories/haulingMarket_test.go` — cache, snapshot upsert, expiry
- `internal/repositories/haulingRunPnl_test.go` — Phase 2: upsert, retrieval, summary aggregation
- `internal/updaters/haulingMarket_test.go` — arbitrage calculation, spread tiering
- `internal/updaters/haulingNotifications_test.go` — Phase 2: Tier 2 threshold, run completion, digest triggers
- `internal/updaters/haulingCorpOrders_test.go` — Phase 2: corp order matching, quantity_acquired updates
- `internal/controllers/haulingRuns_test.go` — endpoint handlers, error cases; Phase 2 adds P&L endpoint tests

**E2E Tests:**
- `e2e/tests/18-hauling-runs.spec.ts` — Phase 1 workflow:
  - Create run, add items
  - Update acquired quantities
  - Verify fill percentages and net profit
  - Market scanner flow
  - Fill suggestions [Add] integration

- `e2e/tests/19-hauling-runs-phase2.spec.ts` — Phase 2 workflow:
  - Discord notification toggles in New Run dialog
  - P&L entry UI on run detail page
  - Route Safety card (jump count + kills)
  - Corp order auto-fill scenarios
  - Daily digest runner behavior

**Mock ESI:**
- `cmd/mock-esi/main.go` — Added market history endpoint for scanner caching validation; Phase 2 adds corp orders endpoint (esi-corporations.read_orders.v1)

## Phase 3 Deferred Items

1. **Split-run operators** — Track which character is responsible for final haul in multi-char runs (operator_character_id on run). Phase 2 treats all runs as single-character.

2. **Undercut monitoring** — Alert if destination sell price undercuts by >X% after initial scan. Phase 2 deferred to Phase 3.

3. **Dotlan integration** — Fetch kill data and player traffic statistics to highlight risky routes. Phase 2 replaced with client-side zKillboard lookup.

## Open Questions

- [ ] Should "Fill Remaining Capacity" be a dialog modal or panel on run detail? Current design uses inline panel.
- [ ] Should Phase 2 auto-populate acquired quantities from corp wallet orders, or require manual confirmation?
- [ ] Should P&L tracking support multi-item post-run summaries, or one entry per type_id?
- [ ] Should haul_threshold_isk alert at run creation or at acquisition milestones (e.g., 50% filled)?
