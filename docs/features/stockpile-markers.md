# Stockpile Markers Feature

## Overview

Stockpile markers let players set desired quantity targets on individual items in their inventory and track "deficits" — items whose current quantity falls below the target. The feature spans the full stack: a database table for marker storage, backend CRUD + deficit aggregation endpoints, and two frontend surfaces (inline markers on the assets page, dedicated deficits dashboard).

**Key pages:**
- `/inventory` — set, edit, and delete stockpile markers on any asset row; add markers for items not yet in inventory via section-level "Add Stockpile" icons
- `/stockpiles` — view all items below target with deficit quantities and ISK costs (includes orphan markers with zero quantity)

## Business Context

Industrial players in EVE Online maintain material buffers (minerals, components, ships) across multiple characters, corporations, stations, and containers. Manually tracking what needs restocking is tedious. Stockpile markers turn the asset list into a live "shopping list" by flagging items that need replenishment.

### User Stories

**As an industrial player, I want to:**
1. Set a desired quantity on any item in my inventory so I know when I'm running low
2. See at a glance which items are below target, grouped by location
3. Know the ISK cost to fill all my deficits at current Jita buy prices
4. Export my deficit list to Janice for quick appraisal
5. Set markers on both personal character assets and corporation division assets
6. Create stockpile markers for items I don't have yet (e.g., target 10,000 Tritanium at Jita even with zero in stock)
7. See items with stockpile markers but zero current quantity on the inventory page as phantom rows

## Architecture

### Data Flow

```
User sets marker on /inventory page
    ↓
POST /api/stockpiles/upsert → POST /v1/stockpiles (backend)
    ↓
StockpileMarkers repository → UPSERT stockpile_markers table
    ↓
Asset queries LEFT JOIN stockpile_markers → inline delta display
    ↓
GET /v1/stockpiles/deficits → CTE query across all asset types
    ↓
/stockpiles page renders deficit table with ISK costs
```

### Key Design Decisions

#### 1. Marker Granularity

**Decision:** Markers are scoped to the combination of `(user, type, owner, location, container, division)`.

A player can set different targets for the same item type in different locations. For example:
- 50,000 Tritanium at Jita IV in character hangar
- 100,000 Tritanium at Jita IV in corp division "Production Materials"
- 10,000 Tritanium at Amarr VIII in character hangar

This uses a composite unique index with `COALESCE` for nullable columns (container_id, division_number).

#### 2. Inline Display vs Separate Page

**Decision:** Both. Markers are set and displayed inline on the `/inventory` page (next to each asset row), and a dedicated `/stockpiles` page shows only items below target.

- **Inline** — set markers while browsing assets, see delta immediately
- **Dedicated** — view all deficits at once, export to Janice, see total ISK cost

#### 3. Deficit Cost Calculation

**Decision:** Use Jita buy price (best bid) for deficit cost.

When you need to acquire missing items, you'll pay the buy price, not the sell price. The deficit query LEFT JOINs `market_prices` and multiplies `ABS(deficit) * buy_price`.

#### 4. Owner Type Support

**Decision:** Markers support both `character` and `corporation` owner types.

- **Character assets** — matched by `owner_id = character_id`
- **Corporation assets** — matched by `owner_id = corporation_id` + `division_number`

#### 5. Orphan Markers (Items Not in Inventory)

**Decision:** Support creating stockpile markers for items that don't exist in the asset tables yet.

- **Problem:** The original assets-first query (`FROM character_assets LEFT JOIN stockpile_markers`) silently drops markers with no matching asset row, making them invisible on both the inventory page and the deficits page.
- **Solution — Inventory page:** `InjectOrphanStockpileRows` post-processes the `GetUserAssets` response, querying all stockpile markers and injecting phantom `Asset` rows (quantity=0) for markers that have no matching asset.
- **Solution — Deficits page:** Two additional `UNION ALL` sections (sections 5 and 6) in `GetStockpileDeficits` use `NOT EXISTS` anti-joins against the asset tables to surface orphan markers with full deficit.
- **Frontend dialog:** An "Add Stockpile" icon button appears on Personal Hangar, Container, and Corp Hanger section headers in the inventory page. The dialog auto-populates location/owner context from the section and provides item type search via `/api/item-types/search`. On save, a phantom asset is optimistically injected into local state.

## Database Schema

### Table: `stockpile_markers`

```sql
CREATE TABLE stockpile_markers (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    owner_type VARCHAR(20) NOT NULL,       -- 'character' or 'corporation'
    owner_id BIGINT NOT NULL,              -- character ID or corporation ID
    location_id BIGINT NOT NULL,           -- station ID
    container_id BIGINT,                   -- NULL for hangar items, item_id for container items
    division_number INT,                   -- NULL for characters, 1-7 for corp divisions
    desired_quantity BIGINT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_stockpile_unique ON stockpile_markers(
    user_id, type_id, owner_type, owner_id, location_id,
    COALESCE(container_id, 0), COALESCE(division_number, 0)
);

CREATE INDEX idx_stockpile_user ON stockpile_markers(user_id);
CREATE INDEX idx_stockpile_type ON stockpile_markers(type_id);
```

### Schema Notes

- **Composite unique index** uses `COALESCE` to handle NULL container_id and division_number — two markers with NULL container are considered the same location
- **owner_type** distinguishes character vs corporation ownership (different JOIN paths in queries)
- **division_number** maps to EVE's CorpSAG1-7 location flags for corporation hangars
- **notes** is optional free-text for user annotations (not currently displayed in UI)

### Migration Files

- **Up:** `internal/database/migrations/20250101120000_stockpile_markers.up.sql`
- **Down:** `internal/database/migrations/20250101120000_stockpile_markers.down.sql`

## Backend Implementation

### Model

**File:** `internal/models/models.go`

```go
type StockpileMarker struct {
    UserID          int64   `json:"userId"`
    TypeID          int64   `json:"typeId"`
    OwnerType       string  `json:"ownerType"`
    OwnerID         int64   `json:"ownerId"`
    LocationID      int64   `json:"locationId"`
    ContainerID     *int64  `json:"containerId"`
    DivisionNumber  *int    `json:"divisionNumber"`
    DesiredQuantity int64   `json:"desiredQuantity"`
    Notes           *string `json:"notes"`
}
```

### Repository: StockpileMarkers

**File:** `internal/repositories/stockpileMarkers.go`

CRUD operations for the `stockpile_markers` table.

| Method | SQL Operation | Notes |
|--------|--------------|-------|
| `GetByUser(ctx, userID)` | `SELECT ... WHERE user_id = $1` | Returns all markers for a user |
| `Upsert(ctx, marker)` | `INSERT ... ON CONFLICT DO UPDATE` | Creates or updates marker, conflict on composite unique index |
| `Delete(ctx, marker)` | `DELETE ... WHERE` | Matches on all composite key columns using `COALESCE` for NULLs |

### Deficit Aggregation

**File:** `internal/repositories/assets.go` — `GetStockpileDeficits(ctx, user)`

This is the most complex query in the system. It uses a CTE (`WITH all_deficits AS`) that combines six UNION ALL subqueries:

1. **Personal hangar items** — `character_assets` WHERE `location_flag IN ('Hangar', 'Deliveries', 'AssetSafety')`
2. **Personal container items** — `character_assets` WHERE `location_type = 'item'` (items inside containers)
3. **Corporation hangar items** — `corporation_asset_locations` WHERE `location_flag LIKE 'CorpSAG%'`
4. **Corporation container items** — `corporation_asset_locations` WHERE `location_type = 'item'`
5. **Orphan character markers** — `stockpile_markers` WHERE `owner_type = 'character'` with `NOT EXISTS` against `character_assets` (markers with no matching asset)
6. **Orphan corporation markers** — `stockpile_markers` WHERE `owner_type = 'corporation'` with `NOT EXISTS` against `corporation_asset_locations` (markers with no matching asset)

Subqueries 1-4 are asset-first (LEFT JOIN markers):
- LEFT JOINs `stockpile_markers` (matching on type, location, container, owner)
- LEFT JOINs `market_prices` (Jita region 10000002) for deficit cost
- Filters to `(quantity - desired_quantity) < 0` (only deficit items)
- Calculates `deficit_value = ABS(delta) * buy_price`

Subqueries 5-6 are marker-first (NOT EXISTS anti-join):
- Sets `quantity = 0`, `stockpile_delta = -desired_quantity`
- Full deficit value calculated as `desired_quantity * buy_price`

Results are ordered by `deficit_value DESC` (most expensive deficits first).

### Summary Aggregation

**File:** `internal/repositories/assets.go` — `GetUserAssetsSummary(ctx, user)`

Used by the landing page to show total deficit ISK. Similar four-way UNION ALL but returns only aggregate `SUM(deficit_value)`.

### Controllers

**File:** `internal/controllers/stockpileMarkers.go`

| Endpoint | Method | Handler | Description |
|----------|--------|---------|-------------|
| `/v1/stockpiles` | GET | `GetStockpiles` | List all markers for authenticated user |
| `/v1/stockpiles` | POST | `UpsertStockpile` | Create or update a marker |
| `/v1/stockpiles` | DELETE | `DeleteStockpile` | Delete a marker |

All endpoints require `web.AuthAccessUser`. The `UserID` is always set from the auth context (not the request body) for security.

**File:** `internal/controllers/stockpiles.go`

| Endpoint | Method | Handler | Description |
|----------|--------|---------|-------------|
| `/v1/stockpiles/deficits` | GET | `GetDeficits` | Get all items below target with costs |

## Frontend Implementation

### Assets Page Integration

**File:** `frontend/packages/components/assets/AssetsList.tsx`

Each asset row in the table shows:
- **Current quantity** and **desired quantity** side by side (e.g., "1,000 / 5,000")
- **Delta indicator** — green (+) if above target, red if below
- **Set/Edit stockpile** button (pencil icon) — opens a dialog to set desired quantity
- **Delete stockpile** button (trash icon) — confirms with `window.confirm()` then removes marker
- **Below target filter** — toggle to show only items with negative delta

Section headers have an **"Add Stockpile"** icon button (Inventory icon with Tooltip) on:
- **Personal Hangar** — extracts unique owners from hangar assets, opens dialog with `locationId` and owner picker
- **Container** — single owner from container context, opens dialog with `locationId` and `containerId`
- **Corp Hanger** — single corporation owner, opens dialog with `locationId` and `divisionNumber`

The stockpile dialog collects:
- `desiredQuantity` — target quantity
- Automatically populates `typeId`, `ownerId`, `ownerType`, `locationId`, `containerId`, `divisionNumber` from the asset context

After upsert/delete, the local state is updated immediately (optimistic update) without refetching all assets.

### Add Stockpile Dialog

**File:** `frontend/packages/components/assets/AddStockpileDialog.tsx`

MUI Dialog for creating stockpile markers on items not in inventory:
- **Item type search** — MUI Autocomplete with debounced async search to `/api/item-types/search?q=...`, shows item icon (32px from evetech CDN) + name
- **Owner picker** — MUI Select, shown only when multiple owners exist. Auto-selected when single owner.
- **Desired quantity** — Number TextField
- **Save** — POSTs to `/api/stockpiles/upsert`. On success, injects a phantom Asset (quantity=0) into local state for immediate display.

### Phantom Row Injection

**File:** `internal/repositories/assets.go` — `InjectOrphanStockpileRows(ctx, userID, response)`

Called by the assets controller after `GetUserAssets`. Queries all stockpile markers for the user, builds a set of existing asset keys, and for each orphan marker (no matching asset), creates a phantom `Asset` with `Quantity=0`, `DesiredQuantity=desired`, `StockpileDelta=-desired`, `DeficitValue=desired*buyPrice` and injects it into the matching structure section.

### Stockpiles Deficits Page

**File:** `frontend/packages/components/stockpiles/StockpilesList.tsx`

Displays a table of all items below target across all characters and corporations.

**Summary cards (sticky header):**
- Items Below Target — count of deficit items
- Total Deficit — sum of all deficit quantities
- Total Volume — estimated volume of deficit items (m3)
- Total Cost (ISK) — sum of deficit values at Jita buy prices

**Table columns:** Item, Structure, Location, Container, Current, Target, Deficit, Cost (ISK), Owner

**Actions:**
- **Search** — filter by item name, structure, solar system, region, or container
- **Copy for Janice** — copies deficit list as "ItemName quantity" text to clipboard
- **Create Janice Appraisal** — POSTs to Janice API and opens appraisal in new tab

**Page file:** `frontend/packages/pages/stockpiles.tsx` (wrapper with auth check)

### Next.js API Routes

**Files:** `frontend/pages/api/stockpiles/`

| Route | Method | Backend Proxy |
|-------|--------|---------------|
| `/api/stockpiles/upsert` | POST | `POST /v1/stockpiles` |
| `/api/stockpiles/delete` | DELETE | `DELETE /v1/stockpiles` |
| `/api/stockpiles/deficits` | GET | `GET /v1/stockpiles/deficits` |

## Testing

### Backend Unit Tests

**Controller tests:** `internal/controllers/stockpiles_test.go` (6 tests)
- Get deficits success, empty result, repository error, multiple deficits, route registration, context propagation

**Marker controller tests:** `internal/controllers/stockpileMarkers_test.go` (7 tests)
- Get/upsert/delete success, invalid JSON, repository errors

**Repository tests:** `internal/repositories/stockpileMarkers_test.go`
- Upsert, get by user, delete, conflict handling

**Deficit tests:** `internal/repositories/stockpileDeficits_test.go`
- Deficit query correctness across character and corporation assets

### E2E Tests

**File:** `e2e/tests/05-stockpiles.spec.ts` (5 tests)
- Empty state initially
- Set stockpile marker from assets page (via dialog)
- Stockpiles page shows deficit after marker is set
- Edit stockpile marker (change desired quantity)
- Delete stockpile marker (confirm dialog, verify removal)

## Files

### Backend
- `internal/models/models.go` — `StockpileMarker` struct
- `internal/database/migrations/20250101120000_stockpile_markers.up.sql` — Schema
- `internal/database/migrations/20250101120000_stockpile_markers.down.sql` — Rollback
- `internal/repositories/stockpileMarkers.go` — CRUD repository
- `internal/repositories/assets.go` — `GetStockpileDeficits()`, `GetUserAssetsSummary()`
- `internal/controllers/stockpileMarkers.go` — Marker CRUD controller
- `internal/controllers/stockpiles.go` — Deficit aggregation controller

### Frontend
- `frontend/packages/components/assets/AssetsList.tsx` — Inline marker UI + section-level "Add Stockpile" icons
- `frontend/packages/components/assets/AddStockpileDialog.tsx` — Dialog for adding markers to items not in inventory
- `frontend/packages/client/data/models.ts` — `EveInventoryType` type for item search results
- `frontend/packages/components/stockpiles/StockpilesList.tsx` — Deficits dashboard
- `frontend/packages/pages/stockpiles.tsx` — Page wrapper
- `frontend/pages/stockpiles.tsx` — Next.js page entry
- `frontend/pages/api/stockpiles/upsert.ts` — API proxy
- `frontend/pages/api/stockpiles/delete.ts` — API proxy
- `frontend/pages/api/stockpiles/deficits.ts` — API proxy

### Tests
- `internal/controllers/stockpiles_test.go` — Deficit controller tests
- `internal/controllers/stockpileMarkers_test.go` — Marker controller tests
- `internal/controllers/assets_test.go` — Assets controller tests (incl. phantom row injection mock)
- `internal/repositories/stockpileMarkers_test.go` — Repository tests
- `internal/repositories/stockpileDeficits_test.go` — Deficit query tests (incl. orphan marker tests)
- `frontend/packages/components/assets/__tests__/AddStockpileDialog.test.tsx` — AddStockpileDialog snapshot and unit tests
- `e2e/tests/05-stockpiles.spec.ts` — End-to-end tests

---

**Status:** Complete

### Changelog

- **v2 — Orphan stockpile markers:** Added ability to create stockpile markers for items not currently in inventory. Backend injects phantom rows into inventory page and deficit query includes orphan markers via NOT EXISTS anti-joins. Frontend AddStockpileDialog opens from section headers on inventory page.
