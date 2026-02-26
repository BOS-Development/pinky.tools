# Planetary Industry (PI) Management

## Overview

Unified dashboard for managing multi-character PI chains in EVE Online. Tracks production rates, detects stalled extractors and factories, calculates profit, and provides supply chain visibility across all characters.

**ESI Scope**: `esi-planets.manage_planets.v1` (already requested in character auth flow)

**Key Constraint**: ESI PI data only updates when the player views the colony in-game. The `last_update` timestamp on each planet tells us when data was last refreshed by the player.

---

## Status

### Phase 1 — Data Fetch, Storage, Display
- [x] Database migration (pi_planets, pi_pins, pi_pin_contents, pi_routes, pi_tax_config)
- [x] ESI client methods (GetCharacterPlanets, GetCharacterPlanetDetails)
- [x] PI repository (piPlanets.go, piTaxConfig.go)
- [x] PI updater with pin classification
- [x] Dedicated PI background runner (configurable interval)
- [x] PI controller with stall detection + production rates
- [x] Wired in root.go and settings.go
- [x] Frontend (overview page with planet cards, search, stats, item/planet icons, route-based input display)

### Phase 2 — Profit Calculation + POCO Tax
- [x] Tier classification (walk sde_planet_schematic_types: R0→P1→P2→P3→P4)
- [x] Per-factory profit calculation (output value - input cost - export tax - import tax)
- [x] Price source selector (Jita sell/buy/split)
- [x] POCO tax integration (global + per-planet rates from pi_tax_config)
- [x] Profit tab in frontend (per-planet summary table with expandable factory breakdown)
- [x] Summary cards (total revenue, costs, taxes, profit per hour)

### Phase 3 — Launchpad Naming + Loading Screen
- [x] pi_launchpad_labels table
- [x] Label CRUD endpoints (POST/DELETE /v1/pi/launchpad-labels)
- [x] Launchpad detail endpoint (GET /v1/pi/launchpad-detail)
- [x] Connected factory input tracking with depletion times
- [x] Inline label editing in drawer UI
- [x] Clickable launchpads in planet cards → detail drawer

### Phase 4 — Discord Stall Alerts
- [x] `pi_stall` notification event type
- [x] `NotifyPiStall` in notifications updater with Discord embed (red alert, character/planet/issue fields)
- [x] Stall detection in PI updater with state transition tracking (`last_stall_notified_at` dedup)
- [x] `pi_stall` added to Discord settings frontend EVENT_TYPES

### Phase 5 — Supply Chain Analysis
- [x] Cross-character input/output aggregation (extractor outputs, factory inputs/outputs)
- [x] Stockpile integration (bought supply from stockpile markers)
- [x] Depletion time calculation for net-deficit items
- [x] Supply Chain tab in frontend (tier filter, search, expandable producer/consumer detail)
- [x] GET /v1/pi/supply-chain endpoint

---

## Key Design Decisions

1. **Dedicated PI tables** — Colony structure (pins, routes, schematics) doesn't fit the asset model
2. **Dedicated background runner** — Separate from asset refresh, configurable via `PI_UPDATE_INTERVAL_SEC` (default 3600s)
3. **Pin category stored at write time** — Classified during ESI fetch, avoids SDE join on every read
4. **Scope check per character** — Skip PI fetch for characters without `esi-planets.manage_planets.v1` in `esi_scopes`
5. **Global + per-planet tax** — Two-level hierarchy matching real POCO variation
6. **Stall dedup via `last_stall_notified_at`** — Only alert on state transition (running → stalled), not repeatedly

---

## Database Schema

```sql
-- Core PI colony data (one row per character+planet)
create table pi_planets (
    id bigserial primary key,
    character_id bigint not null,
    user_id bigint not null,
    planet_id bigint not null,
    planet_type varchar(20) not null,
    solar_system_id bigint not null,
    upgrade_level int not null default 0,
    num_pins int not null default 0,
    last_update timestamp not null,
    last_stall_notified_at timestamp,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    unique(character_id, planet_id)
);

-- Individual pins on each planet
create table pi_pins (
    id bigserial primary key,
    character_id bigint not null,
    planet_id bigint not null,
    pin_id bigint not null,
    type_id bigint not null,
    schematic_id int,
    latitude double precision,
    longitude double precision,
    install_time timestamp,
    expiry_time timestamp,
    last_cycle_start timestamp,
    extractor_cycle_time int,
    extractor_head_radius double precision,
    extractor_product_type_id bigint,
    extractor_qty_per_cycle int,
    extractor_num_heads int,
    pin_category varchar(20) not null,
    updated_at timestamp not null default now(),
    unique(character_id, planet_id, pin_id)
);

-- Contents of storage/launchpad pins
create table pi_pin_contents (
    character_id bigint not null,
    planet_id bigint not null,
    pin_id bigint not null,
    type_id bigint not null,
    amount bigint not null,
    primary key(character_id, planet_id, pin_id, type_id)
);

-- Routes between pins
create table pi_routes (
    character_id bigint not null,
    planet_id bigint not null,
    route_id bigint not null,
    source_pin_id bigint not null,
    destination_pin_id bigint not null,
    content_type_id bigint not null,
    quantity bigint not null,
    primary key(character_id, planet_id, route_id)
);

-- User POCO tax configuration
create table pi_tax_config (
    id bigserial primary key,
    user_id bigint not null references users(id),
    planet_id bigint,
    tax_rate double precision not null default 10.0,
    unique(user_id, planet_id)
);

-- User-defined launchpad labels
create table pi_launchpad_labels (
    user_id bigint not null references users(id),
    character_id bigint not null,
    planet_id bigint not null,
    pin_id bigint not null,
    label varchar(100) not null,
    primary key(user_id, character_id, planet_id, pin_id)
);
```

---

## ESI Endpoints

| Method | ESI Path | Purpose |
|--------|----------|---------|
| GET | `/v1/characters/{id}/planets/` | Planet list (type, system, upgrade, num_pins, last_update) |
| GET | `/v3/characters/{id}/planets/{planet_id}/` | Colony detail (pins with contents/extractor/factory details, routes) |

---

## Pin Classification

Pins are classified during ESI fetch using a combination of ESI detail fields and known type_id sets:

| Category | Detection |
|----------|-----------|
| extractor | `extractor_details` present |
| factory | `factory_details` present |
| launchpad | type_id in {2256, 2542, 2543, 2544} |
| storage | type_id in {2257, 2535, 2536, 2541} |
| command_center | type_id in {2524-2531} |

---

## Stall Detection

Computed at read time in the controller (not stored):

| Condition | Logic | Status |
|-----------|-------|--------|
| Extractor expired | `expiry_time < now()` | `"expired"` |
| Factory idle | `last_cycle_start + (cycle_time × 2) < now()` | `"stalled"` |
| Data stale | `planet.last_update` older than 48 hours | `"stale_data"` |

Planet-level status is the worst of its pin statuses.

### Discord Stall Alerts (Phase 4)

After each PI data refresh, the updater runs stall detection using the same logic as the controller. Notifications are deduped using `pi_planets.last_stall_notified_at`:

- **Running → Stalled**: Send `pi_stall` Discord notification, set `last_stall_notified_at = now()`
- **Stalled → Stalled**: No notification (already notified)
- **Stalled → Running**: Clear `last_stall_notified_at` (ready for future alerts)

The Discord embed includes character name, planet type, solar system, and a summary of stalled pins (e.g., "2 extractors expired, 1 factory stalled").

---

## Production Rates

- **Extractors**: `qty_per_cycle / cycle_time_seconds × 3600` = units/hour
- **Factories**: From `sde_planet_schematics`: `output_quantity / cycle_time_seconds × 3600` = units/hour

---

## API Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/v1/pi/planets` | User | All planets with status, extractors, factories, launchpads |
| GET | `/v1/pi/profit?priceSource=sell` | User | Per-planet profit breakdown (sell/buy/split pricing) |
| GET | `/v1/pi/tax` | User | Tax config (global + overrides) |
| POST | `/v1/pi/tax` | User | Upsert tax config |
| DELETE | `/v1/pi/tax` | User | Delete tax config entry |
| POST | `/v1/pi/launchpad-labels` | User | Upsert launchpad label |
| DELETE | `/v1/pi/launchpad-labels` | User | Delete launchpad label |
| GET | `/v1/pi/launchpad-detail` | User | Launchpad detail with connected factories + depletion |
| GET | `/v1/pi/supply-chain` | User | Cross-character supply chain aggregation with stockpile integration |

---

## PI Tax Formula (Phase 2)

- **Export**: `base_cost_per_unit × quantity × (tax_rate / 100)`
- **Import**: `base_cost_per_unit × quantity × (tax_rate / 100) × 0.5`
- Base costs: R0=5, P1=400, P2=7,200, P3=60,000, P4=1,200,000 ISK

---

## Supply Chain Analysis (Phase 5)

Aggregates production and consumption rates across all characters to show the net balance of every PI material in the user's chain.

### Data Sources

| Source | What it provides |
|--------|-----------------|
| Extractor pins | `product_type_id` at `qty_per_cycle / cycle_time * 3600` units/hour |
| Factory outputs | Schematic output at `output_qty / cycle_time * 3600` units/hour |
| Factory inputs (demand) | Schematic inputs at `input_qty / cycle_time * 3600` units/hour |
| Stockpile markers | Purchased/available inventory (e.g., user buys P1 from Jita) |

### Source Classification

Each item is classified based on how it enters the chain:
- **Extracted**: Produced by extractors (R0 raw resources)
- **Produced**: Output of factories
- **Bought**: Not produced by any planet but available in stockpile markers
- **Extracted + Produced**: Both extractor output and factory output

### Depletion Time

For items with net deficit (`consumed > produced`):
- `depletion_hours = stockpile_qty / (consumed_per_hour - produced_per_hour)`
- Only calculated when stockpile quantity > 0 and there is a net deficit
- Frontend color-codes: red (< 24h), yellow (< 72h), default otherwise

### Expandable Detail

Each row expands to show which planets produce and consume the item, with character name, solar system, planet type, and per-planet rate.

---

## File Structure

### Backend

| File | Purpose |
|------|---------|
| `internal/database/migrations/20260221000943_create_pi_tables.{up,down}.sql` | Schema |
| `internal/database/migrations/20260220131419_create_pi_launchpad_labels.{up,down}.sql` | Launchpad labels schema |
| `internal/repositories/piPlanets.go` | CRUD for pi_planets, pi_pins, pi_pin_contents, pi_routes |
| `internal/repositories/piTaxConfig.go` | CRUD for pi_tax_config |
| `internal/repositories/piLaunchpadLabels.go` | CRUD for pi_launchpad_labels |
| `internal/updaters/pi.go` | Fetch ESI data, classify pins, upsert, stall detection + notification |
| `internal/runners/pi.go` | Dedicated background runner |
| `internal/controllers/pi.go` | HTTP handlers with stall detection + production rates + profit + launchpad detail |

### Frontend

| File | Phase | Purpose |
|------|-------|---------|
| `frontend/pages/pi.tsx` | 1 | Thin page wrapper |
| `frontend/packages/pages/pi.tsx` | 1+2+5 | Main PI page with tabs (Overview, Profit, Supply Chain) |
| `frontend/packages/components/pi/PlanetOverview.tsx` | 1 | Grid of planet cards with search/stats |
| `frontend/packages/components/pi/ProfitTable.tsx` | 2 | Per-product profit table with factory breakdown |
| `frontend/packages/components/pi/LaunchpadDetail.tsx` | 3 | Drawer with factory inputs, depletion, editable labels |
| `frontend/pages/api/pi/planets.ts` | 1 | API route proxy |
| `frontend/pages/api/pi/profit.ts` | 2 | API route proxy for profit endpoint |
| `frontend/pages/api/pi/tax.ts` | 1 | API route proxy |
| `frontend/pages/api/pi/launchpad-labels.ts` | 3 | API route proxy for label CRUD |
| `frontend/pages/api/pi/launchpad-detail.ts` | 3 | API route proxy for launchpad detail |
| `frontend/packages/components/pi/SupplyChain.tsx` | 5 | Supply chain analysis table with tier filter and expandable detail |
| `frontend/pages/api/pi/supply-chain.ts` | 5 | API route proxy for supply chain endpoint |

### Modified Files

| File | Change |
|------|--------|
| `internal/models/models.go` | PI model structs |
| `internal/client/esiClient.go` | GetCharacterPlanets, GetCharacterPlanetDetails |
| `internal/repositories/character.go` | GetNames method |
| `internal/repositories/solarSystems.go` | GetNames method |
| `internal/repositories/sdeData.go` | GetAllSchematics, GetAllSchematicTypes |
| `cmd/industry-tool/cmd/root.go` | Wire PI components (updater, runner, controller, stall notifier) |
| `cmd/industry-tool/cmd/settings.go` | PiUpdateIntervalSec setting |
| `internal/repositories/marketPrices.go` | GetAllJitaPrices (used by profit endpoint) |
| `internal/updaters/notifications.go` | `PiStallNotifier` interface, `NotifyPiStall`, `buildPiStallEmbed` |
| `frontend/packages/components/settings/DiscordSettings.tsx` | Added `pi_stall` event type |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PI_UPDATE_INTERVAL_SEC` | `3600` | How often to refresh PI data from ESI |
