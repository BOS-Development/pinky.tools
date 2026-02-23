# Transportation System

## Status

- **Concept**: In review — no implementation yet

## Overview

Transport jobs track the movement of items between EVE stations. They're created automatically (from plan generation or stockpile deficits) or manually, and appear alongside manufacturing/reaction jobs in the unified job queue. Each transport job has a cost estimate based on the transport method (freighter vs jump freighter), route, volume, and collateral.

---

## Transport Methods

**Freighter** (gate-to-gate, high volume)
- Travels via stargates, route calculated by ESI `GET /route/{origin}/{destination}/`
- Supports routing preferences: shortest, secure (highsec only), insecure
- Cost driven by: jump count, cargo volume, collateral
- Restrictions: slow, vulnerable to ganking in highsec, cannot enter lowsec/nullsec safely without escort
- Ships: Charon, Obelisk, Fenrir, Providence
- Typical cargo: ~435,000–1,000,000+ m³ (with expanders)

**Jump Freighter** (cynosural jumps, high volume)
- Jumps directly between systems using cynosural fields
- Uses **pre-designated routes** — user configures named JF routes with specific cyno waypoints (e.g., "Jita → Ignoitton → GE-8JV" with 3 cyno stops)
- Cost driven by: fuel (isotopes) consumed per leg, collateral
- Fuel formula per leg: `base_fuel_per_LY × distance_LY × (1 - JFC_skill_level × 0.10)`
- Each race uses different isotopes (Nitrogen, Hydrogen, Helium, Oxygen)
- Restrictions: maximum jump range per jump (skills-dependent), requires cyno alt at each waypoint
- Ships: Rhea, Ark, Nomad, Anshar
- Typical cargo: ~30,000–40,000 m³

**Deep Space Transport (DST)** (gate-to-gate, medium volume, tanky)
- Travels via stargates like freighters but with much smaller cargo hold
- Has fleet hangar (~60,000 m³) + cargo bay, and can fit a Micro Jump Drive
- Significantly more survivable than freighters (can use MJD + cloak trick)
- Cost model: same as freighter (per-m³-per-jump + collateral)
- Ships: Impel, Bustard, Occator, Mastodon
- Typical cargo: ~50,000–62,500 m³ (fleet hangar)

**Blockade Runner** (gate-to-gate, small volume, covert)
- Travels via stargates with covert ops cloak — virtually uncatchable
- Very small cargo capacity but ideal for high-value, low-volume items
- Cost model: same as freighter (per-m³-per-jump + collateral)
- Ships: Prorator, Crane, Viator, Prowler
- Typical cargo: ~4,000–12,000 m³

---

## Data Model

### `transport_jobs` (new table)

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | PK |
| `user_id` | bigint | FK → users |
| `origin_station_id` | bigint | Source station/structure |
| `destination_station_id` | bigint | Destination station/structure |
| `origin_system_id` | bigint | Source solar system (for route calc) |
| `destination_system_id` | bigint | Destination solar system |
| `transport_method` | text | `freighter`, `jump_freighter`, `dst`, `blockade_runner`, `manual` |
| `route_preference` | text | `shortest`, `secure`, `insecure` (gate methods only) |
| `status` | text | `planned`, `in_transit`, `delivered`, `cancelled` |
| `total_volume_m3` | float8 | Sum of item volumes |
| `total_collateral` | float8 | Estimated value of cargo (for insurance) |
| `estimated_cost` | float8 | Calculated transport cost |
| `jumps` | int | Gate jumps (gate methods) or cyno jumps (jump freighter) |
| `distance_ly` | float8 | Light-year distance (jump freighter only) |
| `jf_route_id` | bigint | FK → jf_routes (nullable, jump freighter only) |
| `fulfillment_type` | text | `self_haul`, `courier_contract`, `contact_haul` |
| `transport_profile_id` | bigint | FK → transport_profiles (nullable, for self-haul) |
| `plan_run_id` | bigint | FK → production_plan_runs (nullable, if auto-created) |
| `plan_step_id` | bigint | Soft ref to originating step (nullable) |
| `queue_entry_id` | bigint | FK → industry_job_queue (nullable, linked queue entry) |
| `notes` | text | Optional |
| `created_at` | timestamptz | |
| `updated_at` | timestamptz | |

### `transport_job_items` (new table)

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | PK |
| `transport_job_id` | bigint | FK → transport_jobs ON DELETE CASCADE |
| `type_id` | bigint | Item type |
| `quantity` | int | Number of items |
| `volume_m3` | float8 | Per-unit volume × quantity |
| `estimated_value` | float8 | Per-unit value × quantity |

### `jf_routes` (new table — user-defined jump freighter routes)

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | PK |
| `user_id` | bigint | FK → users |
| `name` | text | Route name (e.g., "Jita → GE-8JV") |
| `origin_system_id` | bigint | First system |
| `destination_system_id` | bigint | Last system |
| `total_distance_ly` | float8 | Sum of all leg distances |
| `created_at` | timestamptz | |

### `jf_route_waypoints` (new table — ordered cyno stops)

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | PK |
| `route_id` | bigint | FK → jf_routes ON DELETE CASCADE |
| `sequence` | int | Order in route (0 = origin, 1 = first cyno, ...) |
| `system_id` | bigint | Solar system for this waypoint |
| `distance_ly` | float8 | LY distance from previous waypoint (calculated from SDE coordinates) |

### `transport_profiles` (new table — per-ship/character transport configurations)

A user can have multiple transport profiles — one per ship or alt. For example: "Rhea Alt (JFC 5)", "Charon Main", "Viator Scout".

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | PK |
| `user_id` | bigint | FK → users |
| `name` | text | Profile name (e.g., "Rhea Alt - JFC 5") |
| `transport_method` | text | `freighter`, `jump_freighter`, `dst`, `blockade_runner` |
| `character_id` | bigint | FK → characters (nullable, links to a specific alt) |
| `cargo_m3` | float8 | Ship's cargo/fleet hangar capacity |
| `rate_per_m3_per_jump` | float8 | ISK per m³ per gate jump (gate methods) |
| `collateral_rate` | float8 | Fraction of cargo value |
| `collateral_price_basis` | text | `buy`, `sell`, or `split` |
| `fuel_type_id` | bigint | Isotope type ID (JF only) |
| `fuel_per_ly` | float8 | Base fuel consumption per LY (JF only) |
| `fuel_conservation_level` | int | Jump Fuel Conservation skill 0-5 (JF only) |
| `route_preference` | text | `shortest`, `secure`, `insecure` (gate methods only) |
| `is_default` | boolean | Default profile for this transport method |
| `created_at` | timestamptz | |

When creating a transport job, the user selects which profile to use. The profile determines the cost model, cargo capacity (for trip splitting), and the character assigned to the haul.

### `transport_trigger_config` (new table — per-user, per-trigger fulfillment preferences)

| Column | Type | Description |
|--------|------|-------------|
| `user_id` | bigint | PK (with trigger_type) |
| `trigger_type` | text | PK — `plan_generation`, `stockpile_deficit`, `manual` |
| `default_fulfillment` | text | Default: `self_haul`, `courier_contract`, or `contact_haul` |
| `allowed_fulfillments` | text[] | Which fulfillment types are available (e.g., `{self_haul, courier_contract}`) |
| `default_profile_id` | bigint | FK → transport_profiles (nullable, default for self-haul) |
| `default_method` | text | Default transport method for this trigger |
| `courier_rate_per_m3` | float8 | Flat rate per m³ for courier/contact fulfillment |
| `courier_collateral_rate` | float8 | Collateral fraction for courier/contact fulfillment |

### `industry_job_queue` (modified)

Transport jobs also create a corresponding queue entry with `activity = 'transport'`. The queue entry links back to the transport job via a new nullable column:

| Column | Type | Description |
|--------|------|-------------|
| `transport_job_id` | bigint | FK → transport_jobs (nullable) |

This keeps transport visible in the unified job queue alongside manufacturing/reaction jobs.

---

## Cost Calculation

Cost calculation depends on the **fulfillment type**, not just the transport method.

### Self-Haul (detailed, profile-based)

Uses the assigned transport profile's ship stats and skills for precise costing.

**Gate methods (Freighter / DST / Blockade Runner):**
```
route = ESI GET /route/{origin_system}/{dest_system}/?flag={preference}
jumps = len(route) - 1
trips = ceil(total_volume_m3 / profile.cargo_m3)
cost  = ((volume_m3 × profile.rate_per_m3_per_jump × jumps) + (collateral × profile.collateral_rate)) × trips
```

**Jump Freighter:**
```
# Uses a pre-designated JF route (user-configured with cyno waypoints)
# Each leg's distance calculated from SDE system coordinates:
#   leg_distance_ly = sqrt((x2-x1)² + (y2-y1)² + (z2-z1)²) / 9.461e+15
# Fuel per leg:
#   fuel_units = ceil(profile.fuel_per_ly × leg_distance_ly × (1 - profile.fuel_conservation_level × 0.10))
# Total:
total_fuel = sum(fuel_units for each leg in route)
fuel_cost  = total_fuel × isotope_price (based on profile.collateral_price_basis)
trips      = ceil(total_volume_m3 / profile.cargo_m3)
cost       = (fuel_cost + (collateral × profile.collateral_rate)) × trips
```

JF routes are pre-configured by the user with specific cyno waypoints. When creating a self-haul JF transport job, the user selects from their saved routes.

### Courier Contract / Contact Haul (flat rate)

When contracting out to another player, the cost is based on a flat rate — you don't know their ship or skills. The user configures these rates per trigger (via `transport_trigger_config`) or overrides per job.

```
cost = (volume_m3 × flat_rate_per_m3) + (collateral × collateral_rate)
```

Or for jump freighter courier contracts:
```
cost = flat_reward  (user sets the contract reward directly)
```

The flat rate represents what the user expects to pay for the courier service.

### Manual

User enters a flat cost. No automatic calculation.

### Cargo Capacity & Trip Splitting

Transport jobs enforce cargo capacity limits. Each transport profile specifies the ship's cargo capacity (`cargo_m3`). When a transport job's total volume exceeds the selected profile's capacity:

```
trips = ceil(total_volume_m3 / cargo_capacity_m3)
total_cost = single_trip_cost × trips
```

Courier contracts created from transport jobs are split accordingly — each contract stays within the cargo limit.

### Collateral Valuation

Cargo value for collateral is calculated using the user's `collateral_price_basis` setting on the transport profile:
- **buy**: Jita buy price (highest buy order)
- **sell**: Jita sell price (lowest sell order)
- **split**: Average of buy and sell

Same price basis options as the reactions calculator.

---

## Integration Points

Each integration point is configurable for which fulfillment types are allowed and which is the default. This is stored per-user in `transport_trigger_config` so that plan-generated transport jobs can default to "courier contract" while manual jobs default to "self-haul", etc.

### Production Plans → Transport Jobs

When `GenerateJobs` creates a plan run and detects that a child step's station differs from its parent step's station, auto-create a transport job:
- **Origin**: child step's output station
- **Destination**: parent step's source station
- **Items**: the materials the child step produces for the parent
- **Volume/value**: calculated from SDE item data + quantities
- **Fulfillment**: uses `plan_generation` trigger config defaults

The transport job links to the plan run and appears between the child's manufacturing job and the parent's manufacturing job in the queue.

### Stockpile Deficits → Transport Jobs

When a user has items at station A but a stockpile deficit at station B:
- User selects items to "transport to stockpile"
- System creates a transport job with the deficit items
- **Fulfillment**: uses `stockpile_deficit` trigger config defaults
- Cost estimate helps the user decide whether to self-haul or buy locally

### Manual Creation

User creates a transport job directly, specifying:
- Origin/destination stations
- Items and quantities (or just volume + value for estimation)
- Transport method and fulfillment type (overrides `manual` trigger defaults)

### Job Queue Integration

Each transport job creates a queue entry with `activity = 'transport'`. The queue displays:
- Origin → Destination (instead of blueprint name)
- Volume, estimated cost, transport method
- Status progression: planned → in_transit → delivered

---

## Fulfillment

**Self-haul**: User undocks in EVE, moves items, marks job as delivered.

**Courier contract** (future):
- Create an EVE courier contract with origin, destination, volume, collateral, reward
- Extend contract sync to match courier contracts (currently only matches `item_exchange`)
- Extend `EsiContract` model with courier fields: `start_location_id`, `end_location_id`, `volume`, `collateral`, `reward`
- Auto-complete transport job when courier contract status = `finished`

**Contact marketplace hauling** (future):
- Contacts with `hauling` permission can see your pending transport jobs
- They accept a job, creating a courier contract in-game
- Builds on existing contact permission infrastructure

---

## ESI Client Additions

**New method**: `GetRoute(ctx, origin, destination, flag) ([]int32, error)`
- Endpoint: `GET /route/{origin}/{destination}/?flag={shortest|secure|insecure}`
- Returns array of system IDs representing the gate route
- Public endpoint, no auth required

**Extend `EsiContract`** (for courier contract matching, future phase):
```go
type EsiContract struct {
    // ... existing fields ...
    StartLocationID int64   `json:"start_location_id"`
    EndLocationID   int64   `json:"end_location_id"`
    Volume          float64 `json:"volume"`
    Collateral      float64 `json:"collateral"`
    Reward          float64 `json:"reward"`
}
```

---

## Phases

**Phase 1**: Core transport jobs
- `transport_jobs` + `transport_job_items` + `transport_profiles` tables
- `jf_routes` + `jf_route_waypoints` tables for user-defined JF routes
- `transport_trigger_config` table for per-trigger fulfillment preferences
- ESI route endpoint integration (freighter gate routes)
- LY distance calculation from SDE coordinates (JF routes)
- Manual transport job creation (CRUD)
- Freighter and jump freighter cost calculation with trip splitting
- Job queue integration (transport entries appear in queue)
- Transport profiles management (create/edit/delete ship profiles with rates, fuel config, cargo capacity)
- JF route management page (create/edit/delete routes with waypoints)

**Phase 2**: Plan generation integration
- Auto-create transport jobs when plan run detects cross-station transfers
- Transport jobs linked to plan runs, visible in run detail
- Cost rolls up into plan run total cost estimate

**Phase 3**: Stockpile integration
- "Transport to stockpile" action on items at other stations
- Transport job auto-populated with deficit items and quantities
- Stockpile deficit page shows estimated transport cost

**Phase 4**: Courier contract sync
- Extend EsiContract with courier fields
- Match courier contracts to transport jobs (by locations + volume)
- Auto-complete transport jobs when courier delivered

**Phase 5**: Contact marketplace hauling
- `hauling` permission type for contacts
- Contacts can browse/accept pending transport jobs
- Courier contract creation flow

---

## Key Decisions

1. **Standalone table + queue entry**: Transport jobs have their own table for logistics detail but also create a job queue entry so they appear in the unified work view.
2. **Fulfillment-dependent costing**: Self-haul uses detailed profile-based calculation (fuel, per-m³-per-jump). Courier/contact uses flat rates since you don't know the hauler's ship or skills.
3. **Multi-profile support**: Users can have multiple transport profiles (one per ship/alt), not a single global setting. A Rhea alt and a Charon main each get their own profile with independent fuel, cargo, and rate configs.
4. **Pre-designated JF routes**: Jump freighter routes are user-configured with specific cyno waypoints, not auto-calculated. This reflects real gameplay where JF pilots have established cyno chains.
5. **Fuel calculated at query time**: JF route waypoints store only distances (from SDE coordinates). Fuel units are derived from the assigned profile's `fuel_per_ly` and `fuel_conservation_level`, so changing skills or ships automatically updates cost estimates.
6. **System coordinates for LY distance**: `solar_systems` table already has x/y/z coordinates from SDE. No new data source needed.
7. **Soft link to plan steps**: `plan_step_id` has no FK (same pattern as job queue) because steps can be deleted between runs.
8. **Route caching**: ESI gate route responses are deterministic for the same origin/destination/flag — can be cached aggressively.
9. **User-selectable price basis**: Collateral valuation uses buy, sell, or split price — user choice per profile, same pattern as reactions calculator.
10. **Cargo capacity enforcement**: Transport jobs calculate trip count based on ship cargo capacity. Courier contracts are split to stay within limits.
11. **Configurable trigger fulfillment**: Each integration point (plan generation, stockpile deficit, manual) independently configures which fulfillment types are allowed and which is the default.
