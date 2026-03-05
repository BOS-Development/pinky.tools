# Station Markets

## Status

**Phase**: 1 - Implemented (Issue #162)

**Scope**: Station-level market selection for hauling scanner, NPC station presets, player-owned structure markets, structure market caching

## Overview

Station Markets extends the hauling scanner to support granular location selection beyond region-wide scans. Players can now select from preset NPC stations (Jita, Amarr, Dodixie, Rens, Hek), add their own player-owned trading structures (citadels, engineering complexes), and scan structure markets for arbitrage opportunities.

The system caches structure market data separately from region snapshots, using a 30-minute TTL to minimize ESI API calls while keeping structure prices fresh. Structure access is validated via ESI (403 = no access), and users see visual indicators (⚠) for structures they cannot read. The UI shows contextual feedback including a "My Structures" group with empty-state messaging and helper CTAs when no structures are configured.

## Key Decisions

1. **Station vs Structure terminology** — NPC stations stored in `trading_stations` with `is_preset=true` for Jita 4-4. Player structures stored in `user_trading_structures` (user_id scoped). Both use same location picker UI with grouped selector.

2. **NPC station preset seeding** — Five major NPC stations seeded as `is_preset=true` in `trading_stations` and cannot be deleted:
   - Jita 4-4 (station_id=60003760, system_id=30000142, region_id=10000002)
   - Amarr VIII (station_id=60008494, system_id=30002187, region_id=10000043)
   - Dodixie V (station_id=60011866, system_id=30002659, region_id=10000032)
   - Rens VI (station_id=60004588, system_id=30002053, region_id=10000030)
   - Hek VIII (station_id=60005686, system_id=30002502, region_id=10000033)

   Seeded via migration; acts as default reference locations across major trading hubs.

3. **Structure discovery via ESI** — Users add structures by structure_id. On add, we call `GET /v2/universe/structures/{id}` to validate and fetch name/location. Returns `{accessOk: false}` if 403 (access denied). Frontend shows ⚠ badge; structure remains in list but is not scannable until access is granted (scope/permissions).

4. **Per-user structure ownership** — `user_trading_structures` tracks character_id (which character has access). Multiple users can add the same structure; each gets their own row. Enables shared corporate structures.

5. **Separate market snapshot table** — `hauling_structure_snapshots` uses PK `(type_id, structure_id)` parallel to `(type_id, region_id, system_id)` regional snapshots. Same price/volume fields; same 30-minute TTL and cache refresh logic.

6. **Scanner unified picker** — Replaced two region dropdowns with single location picker showing: NPC Stations group (Jita), Regions group (scrollable dropdown), My Structures group (user's saved structures with access indicator). Picker emits `sourceStructureId`, `destStructureId`, `destSystemId` (vs. old region_id params).

7. **Structure scan trigger** — POST `/v1/hauling/structures/{id}/scan` calls updater to fetch fresh market data for that structure. Responds immediately; no async job. Frontend shows loading spinner during scan.

8. **Access validation on scan** — If structure access is denied (403), `access_ok` flag is set to `false`, and endpoint returns empty snapshots (or 403). Scanner UI shows warning if trying to scan inaccessible structure.

9. **Source vs destination flexibility** — Hauling run can now specify:
   - `from_station_id` (NPC or structure) + optional `from_system_id`
   - `to_station_id` (NPC or structure) + optional `to_system_id`
   - Falls back to region scan if source/dest are regions (backward compatible)

10. **No region inference** — If user selects a structure, we do NOT auto-infer region_id. Region is stored for reference but source/dest comparison is structure-to-structure or structure-to-region depending on picker selection.

11. **UX polish for structure management** — The location picker always displays a "My Structures" group, with an empty-state placeholder and helper CTA when no structures are configured. Scan progress indicator shows contextual text ("Scanning structure market..." vs "Fetching market data from ESI..."). Add dialog validation improved to clearly message when a character has no accessible asset structures.

## Database Schema

### `trading_stations`

NPC station presets.

- `id` (bigint, PK)
- `station_id` (bigint, unique) — ESI station ID
- `name` (varchar(255)) — Station name (e.g., "Jita IV - Moon 4 - Caldari Navy Assembly Plant")
- `system_id` (bigint) — Solar system ID
- `region_id` (bigint) — Region ID
- `is_preset` (boolean, default false) — true only for Jita 4-4; prevents deletion

**Indexes:**
- `(is_preset)` — fast filter for presets

### `user_trading_structures`

Player-owned trading structures (citadels, engineering complexes, etc.).

- `id` (bigint, PK)
- `user_id` (bigint, FK users CASCADE) — structure owner (user who added it)
- `structure_id` (bigint, unique within user) — ESI structure ID
- `name` (varchar(255)) — Structure name (cached from ESI)
- `system_id` (bigint) — Solar system ID (cached from ESI)
- `region_id` (bigint) — Region ID (cached from ESI)
- `character_id` (bigint, FK characters) — which character has access (ESI token source)
- `access_ok` (boolean, default true) — false if last scan returned 403
- `last_scanned_at` (timestamptz, nullable) — when structure market data was last refreshed
- `created_at`, `updated_at` (timestamps)

**Unique Constraint:** `(user_id, structure_id)`

**Indexes:**
- `(user_id, access_ok)` — filter by user and access status
- FK: user_id (cascade), character_id

**Computed fields (application level):**
- `access_status` — 'ok' | 'denied' (from access_ok flag)

### `hauling_structure_snapshots`

Cached market data per type/structure (parallel to regional snapshots).

- `type_id` (bigint)
- `structure_id` (bigint)
- `buy_price` (numeric(12,2)) — highest buy order price in structure
- `sell_price` (numeric(12,2)) — lowest sell order price in structure
- `volume_available` (bigint) — units available at best sell price
- `avg_daily_volume` (bigint) — market liquidity estimate
- `updated_at` (timestamp)

**Primary Key:** `(type_id, structure_id)`

**Cache Policy:**
- Records older than 30 minutes are candidates for refresh
- Refresh triggered by POST `/v1/hauling/structures/{id}/scan` (on-demand) or background job
- Same 30-min TTL as regional snapshots for consistency

### `hauling_runs` (Updated)

Added optional station tracking to existing hauling_runs table.

- `from_station_id` (bigint, nullable, FK trading_stations) — source NPC station
- `to_station_id` (bigint, nullable, FK trading_stations) — destination NPC station

**Migration:** `20260305_hauling_runs_station_tracking.up.sql`

**Backward Compatibility:** from_region_id/to_region_id remain; scanner falls back to region comparison if stations are not set.

## API Endpoints

### Station Management (New)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/hauling/stations` | None | List preset NPC stations (Jita only) |

**Response:**
```json
[
  {
    "id": 1,
    "stationId": 60003760,
    "name": "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
    "systemId": 30000142,
    "regionId": 10000002,
    "isPreset": true
  }
]
```

### Structure Management (New)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/hauling/structures` | User | List user's trading structures |
| POST | `/v1/hauling/structures` | User | Add new trading structure by structure_id |
| DELETE | `/v1/hauling/structures/{id}` | User | Remove structure from user's list |
| POST | `/v1/hauling/structures/{id}/scan` | User | Refresh structure market data |

**POST /structures body:**
```json
{
  "structureId": 1021000000001,
  "characterId": 1001002003
}
```

**Response (success):**
```json
{
  "id": 42,
  "structureId": 1021000000001,
  "name": "Jita Engineering Complex Alpha",
  "systemId": 30000142,
  "regionId": 10000002,
  "characterId": 1001002003,
  "accessOk": true,
  "lastScannedAt": null
}
```

**Response (403 access denied):**
```json
{
  "id": 42,
  "structureId": 1021000000001,
  "name": "Jita Engineering Complex Alpha",
  "systemId": 30000142,
  "regionId": 10000002,
  "characterId": 1001002003,
  "accessOk": false,
  "lastScannedAt": null
}
```

**GET /structures response:**
```json
[
  {
    "id": 42,
    "structureId": 1021000000001,
    "name": "Jita Engineering Complex Alpha",
    "systemId": 30000142,
    "regionId": 10000002,
    "characterId": 1001002003,
    "accessOk": true,
    "lastScannedAt": "2026-03-05T14:30:00Z"
  }
]
```

**POST /structures/{id}/scan response:**
```json
{
  "success": true,
  "itemsScanned": 1200,
  "lastScannedAt": "2026-03-05T14:35:00Z"
}
```

### Scanner Updates

Updated `/v1/hauling/scanner` and `/v1/hauling/scanner/scan` to support station-level selection.

| Method | Path | Query Params | Description |
|--------|------|--------------|-------------|
| GET | `/v1/hauling/scanner` | `source_structure_id`, `dest_structure_id`, `dest_system_id` | Arbitrage opportunities with optional structure params (region params still supported for backward compatibility) |
| POST | `/v1/hauling/scanner/scan` | N/A | Async market scan |

**GET /scanner query params:**
- `source_region_id` (optional if source_structure_id present)
- `source_structure_id` (optional) — structure ID to scan from
- `dest_region_id` (optional if dest_structure_id present)
- `dest_structure_id` (optional) — structure ID to scan to
- `dest_system_id` (optional) — system-level destination scan
- `sort_by`, `min_spread_pct`, `page`, `limit` (existing params)

**POST /scanner/scan body:**
```json
{
  "sourceStructureId": 1021000000001,
  "destStructureId": 1021000000002,
  "destSystemId": 30000142,
  "charToken": "character_token_for_structure_access"
}
```

**Backward Compatibility:** If source_region_id and dest_region_id are provided without structure IDs, behaves as before.

## File Structure

### Backend

**Migrations:**
- `internal/database/migrations/20260305_create_trading_stations.{up,down}.sql` — trading_stations table
- `internal/database/migrations/20260305_create_user_trading_structures.{up,down}.sql` — user_trading_structures table
- `internal/database/migrations/20260305_create_hauling_structure_snapshots.{up,down}.sql` — hauling_structure_snapshots table
- `internal/database/migrations/20260305_hauling_runs_station_tracking.{up,down}.sql` — adds from_station_id, to_station_id to hauling_runs
- `internal/database/migrations/20260305115121_add_npc_station_presets.{up,down}.sql` — seeds Amarr, Dodixie, Rens, Hek presets into trading_stations

**Models:**
- `internal/models/models.go` — TradingStation, UserTradingStructure (added)

**Repositories:**
- `internal/repositories/tradingStations.go` + test — CRUD: GetPresetStations, CreateStation, DeleteStation (with is_preset check)
- `internal/repositories/userTradingStructures.go` + test — CRUD: CreateStructure, GetStructuresByUser, GetStructureByID, DeleteStructure, UpdateAccessStatus, UpdateLastScannedAt
- `internal/repositories/haulingStructures.go` + test — Cache: UpsertSnapshot, GetSnapshotsByStructure, GetSnapshotsOlderThanByStructure
- `internal/repositories/haulingMarketCombined.go` (new) — Abstraction layer combining regional + structure snapshots for scanner

**Updaters:**
- `internal/updaters/haulingMarket.go` — added:
  - `ScanStructure(structureID, characterToken)` method
  - Structure market fetch logic (calls ESI GetStructureMarketOrders)
  - Cache refresh for structures (same 30-min TTL)

**Controllers:**
- `internal/controllers/tradingStructures.go` + test — HTTP handlers:
  - `GetStations` (GET /v1/hauling/stations)
  - `ListStructures` (GET /v1/hauling/structures)
  - `CreateStructure` (POST /v1/hauling/structures)
  - `DeleteStructure` (DELETE /v1/hauling/structures/{id})
  - `ScanStructure` (POST /v1/hauling/structures/{id}/scan)

**Client:**
- `internal/client/esiClient.go` — added:
  - `GetStructureInfo(structureID)` — GET /v2/universe/structures/{id}
  - `GetStructureMarketOrders(structureID, typeID)` — GET /v5/structures/{id}/orders (needs appropriate token)

### Frontend

**Pages:**
- `frontend/pages/api/hauling/stations.ts` — GET handler (proxy to backend)
- `frontend/pages/api/hauling/structures.ts` — GET/POST handler (proxy to backend)
- `frontend/pages/api/hauling/structure-scan.ts` — POST handler for manual scan trigger

**Components:**
- `frontend/packages/components/hauling/UserStructuresDialog.tsx` — Dialog for adding/removing structures:
  - Structure ID input
  - Character dropdown (for access token selection)
  - List of user's structures with delete buttons
  - ⚠ access denied indicator
  - Manual scan button per structure
- `frontend/packages/components/hauling/MarketScanner.tsx` — Updated to include:
  - Unified location picker (Stations group, Regions group, My Structures group)
  - Grouping and visual separation
  - Structure access status badges
  - Fallback to region selection if no structure selected

**API Client:**
- `frontend/packages/client/api.ts` — added:
  - `getStations()` — GET /v1/hauling/stations
  - `getStructures()` — GET /v1/hauling/structures
  - `createStructure(structureId, characterId)` — POST /v1/hauling/structures
  - `deleteStructure(id)` — DELETE /v1/hauling/structures/{id}
  - `scanStructure(id)` — POST /v1/hauling/structures/{id}/scan
  - Updated `getHaulingScanner()` to accept structure params

### Tests

**Backend Tests:**
- `internal/repositories/tradingStations_test.go` — preset listing, deletion protection
- `internal/repositories/userTradingStructures_test.go` — CRUD, access_ok flag, user scoping
- `internal/repositories/haulingStructures_test.go` — cache ops, structure snapshot upsert, expiry
- `internal/controllers/tradingStructures_test.go` — endpoint handlers, 403 scenarios, deletion protection
- `internal/updaters/haulingMarket_test.go` — added ScanStructure tests, structure cache refresh

**E2E Tests:**
- `e2e/tests/21-station-markets.spec.ts` (22 tests):
  - Stations dropdown loads Jita preset
  - Add structure dialog validates structure_id
  - Add structure handles 403 (access denied) → displays ⚠
  - List user's structures
  - Delete structure
  - Scanner with structure source/destination
  - Manual structure scan (POST /scan)
  - Cache refresh at 30 minutes
  - Backward compatibility: region-only scanner still works
  - Station tracking on hauling_runs table

**Mock ESI:**
- `cmd/mock-esi/main.go` — updated:
  - Added GET /v2/universe/structures/{id} endpoint (returns structure info or 403)
  - Added GET /v5/structures/{id}/orders endpoint (market data)

## Phase 2 (Future)

- **Automated structure discovery** — Option to auto-discover accessible structures from character data
- **Multi-hub arbitrage** — Scanner supports >2 locations for three-way trades
- **Structure grouping** — User-defined folders/categories for structures

## Open Questions

- [ ] Should structure access re-validation happen on every scan, or only on-demand? Currently on-demand (POST /scan).
- [ ] Should preset stations be hardcoded to Jita, or seedable via migration? Currently hardcoded in migration.
- [ ] Should structure snapshots use character-specific pricing (buy/sell from their corp wallet), or market-wide? Currently market-wide.
- [ ] Future: Should we cache structure info (name, location) separately and refresh on a longer schedule (1 hour)?
