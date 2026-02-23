# Preferred Stations

## Overview

Save station configurations linked to real player-owned structures. Users paste in-game structure fitting scans to auto-detect structure type, rigs, and services. Production plan steps reference preferred stations by FK, so structure/rig/security/tax/name are derived at query time and changes propagate automatically.

## Status

- **Phase 1**: Complete — station CRUD, scan parser, plan step integration with activity-aware rig matching
- **Phase 2**: Complete — plan-level default stations (manufacturing + reaction), auto-assigned to new steps
- **Phase 3**: Complete — independent input and output hangar configuration per step with transfer indicators

---

## How It Works

### Data Flow

```
Frontend                    Backend Controller              Database
────────                    ──────────────────              ────────
/stations page           →  GET /v1/user-stations         →  user_stations + rigs + services
                            POST /v1/user-stations        →  (create with rigs + services)
                            PUT /v1/user-stations/{id}    →  (update station + replace rigs/services)
                            DELETE /v1/user-stations/{id}  →  (cascade delete rigs + services)
                            POST /v1/user-stations/parse-scan → (pure parser, no DB)

Plan Step Edit              Backend (plan step update)      Database
──────────                  ──────────────────────          ────────
Select preferred station →  PUT .../steps/{id}             →  production_plan_steps.user_station_id
                            GET .../plans/{id}             →  JOIN user_stations → derive fields
```

### Scan Parser

Parses EVE Online structure fitting scan text to extract structure type, rigs, and services.

**Sections parsed**: Rig Slots, Service Slots (High/Medium/Low Power Slots are ignored)

**Rig categories**:
| Keyword | Category |
|---------|----------|
| Ship Manufacturing | ship |
| Structure and Component | component |
| Equipment Manufacturing | equipment |
| Ammunition Manufacturing | ammo |
| Drone and Fighter | drone |
| Biochemical/Composite/Hybrid/Polymer Reactor | reaction |
| Thukker Component | component |

**Structure detection from rig size prefix**:
| Prefix | Manufacturing | Reaction |
|--------|--------------|----------|
| XL-Set | Sotiyo | — |
| L-Set | Azbel | Tatara |
| M-Set | Raitaru | Athanor |

**Service → activity mapping**:
| Service | Activity |
|---------|----------|
| Manufacturing Plant, Capital/Supercapital Shipyard | manufacturing |
| Biochemical/Composite/Hybrid/Polymer Reactor | reaction |

### Rig Category Matching

When a plan step references a station, the correct rig is determined by the step's activity and SDE product category:

**Manufacturing steps**: SDE category → rig category (6=ship, 7=equipment, 8=ammo, 18/87=drone, else=component)

**Reaction steps**: Always use rig category "reaction"

---

## Schema

### user_stations
| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| user_id | bigint | FK → users |
| station_id | bigint | FK → stations (real structure) |
| structure | text | Calculator value: raitaru, azbel, sotiyo, tatara, athanor |
| facility_tax | numeric(5,2) | Tax percentage |

### user_station_rigs
| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| user_station_id | bigint | FK → user_stations (cascade) |
| rig_name | text | Full rig name from scan |
| category | text | ship, component, equipment, ammo, drone, reaction |
| tier | text | t1 or t2 |

### user_station_services
| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | Primary key |
| user_station_id | bigint | FK → user_stations (cascade) |
| service_name | text | Full service name from scan |
| activity | text | manufacturing or reaction |

### production_plan_steps (added columns)
| Column | Type | Description |
|--------|------|-------------|
| user_station_id | bigint | Nullable FK → user_stations (on delete set null) |
| source_location_id | bigint | Input station ID (derived from user_station_id) |
| source_owner_type | text | Input owner: "character" or "corporation" |
| source_owner_id | bigint | Input owner ID |
| source_division_number | int | Input corp hangar division (1-7) |
| source_container_id | bigint | Input container item ID (optional) |
| output_owner_type | text | Output owner: "character" or "corporation" |
| output_owner_id | bigint | Output owner ID |
| output_division_number | int | Output corp hangar division (1-7) |
| output_container_id | bigint | Output container item ID (optional) |

### production_plans (added columns)
| Column | Type | Description |
|--------|------|-------------|
| default_manufacturing_station_id | bigint | Nullable FK → user_stations (on delete set null) |
| default_reaction_station_id | bigint | Nullable FK → user_stations (on delete set null) |

When creating a plan, users can select default stations. All steps created in that plan auto-inherit the appropriate station based on their activity (manufacturing or reaction).

---

## Key Decisions

- **Individual rig storage**: Rigs stored as separate rows with category and tier, enabling per-item rig matching
- **Reference, not copy**: Plan steps reference stations by FK. Station config changes propagate to all steps using that station
- **Activity-aware filtering**: EditStepDialog only shows stations whose services match the step's activity
- **Enriched query**: Station name, system, security derived via JOINs (not duplicated)
- **rigCategory enrichment**: Computed in SQL via SDE category CASE expression when fetching plan steps
- **Plan-level default stations**: Each plan stores default manufacturing + reaction station IDs. Backend auto-assigns `user_station_id` on step creation based on activity. Different plans can use different stations.
- **Single input location per step**: All materials for a production step come from the same hangar/container, matching EVE Online's behavior where a job pulls all materials from one location. Enforced by one set of source fields per step.
- **Containers nested within hangars**: Selection flow is Owner → Division (corp only) → Container (optional). A container is always within the context of its parent hangar.
- **Independent output per step**: Each step independently configures its output location (owner/division/container). No auto-linking — different steps can output to different hangars/stations.
- **Transfer indicator**: When a child step's station differs from the parent step's station, a warning "Transfer" badge appears in the tree, indicating items must be moved between stations in-game.
- **Backend enrichment for display names**: Source and output owner/division/container names resolved via LEFT JOINs in `GetByID`, avoiding extra frontend API calls.

---

## Input/Output Hangar Configuration

Each production plan step can specify WHERE materials come from (input) and WHERE completed items go (output). Both use the same location hierarchy:

**Owner** (character or corporation) → **Division** (corp hangar division 1-7, only for corporations) → **Container** (optional, a named container within the hangar)

Both input and output are always at the step's own preferred station — in EVE Online, a job's materials and output are at the same station where the job runs.

### How It Works

1. User selects a preferred station for a step
2. "Input Location" and "Output Location" sections appear in EditStepDialog
3. Backend returns available characters, corporations (with division names), and containers at that station via `GET /v1/industry/plans/hangars?user_station_id=X`
4. User selects owner, division (if corp), and optionally a container for both input and output
5. Input fields saved: `source_owner_type`, `source_owner_id`, `source_division_number`, `source_container_id`, `source_location_id`
6. Output fields saved: `output_owner_type`, `output_owner_id`, `output_division_number`, `output_container_id`
7. Display names resolved via LEFT JOINs when fetching the plan

### Transfer Indicator

When a child step runs at a different station than its parent step, a warning "Transfer" badge appears in the step tree. This indicates items must be moved between stations in-game (e.g., reaction products from a Tatara need to be hauled to an Azbel for manufacturing).

---

## File Structure

| File | Purpose |
|------|---------|
| `migrations/20260222175330_create_user_stations.up.sql` | Tables + FK on plan steps |
| `internal/models/models.go` | UserStation, UserStationRig, ScanResult types |
| `internal/parser/scan.go` | Structure scan parser |
| `internal/parser/scan_test.go` | Parser tests (11 cases) |
| `internal/repositories/userStations.go` | CRUD with rigs + JOINs |
| `internal/repositories/userStations_test.go` | Integration tests (7 cases) |
| `internal/controllers/userStations.go` | HTTP handlers |
| `internal/controllers/userStations_test.go` | Controller tests (7 cases) |
| `internal/repositories/productionPlans.go` | rigCategory enrichment + station override + default station columns + container query + source name enrichment |
| `internal/controllers/productionPlans.go` | Auto-assign station to steps from plan defaults + GetHangars endpoint |
| `migrations/20260222185246_add_plan_default_stations.up.sql` | Default station columns on plans |
| `frontend/packages/components/stations/StationsList.tsx` | Station list table |
| `frontend/packages/components/stations/StationDialog.tsx` | Add/edit dialog with scan |
| `frontend/packages/components/industry/ProductionPlanEditor.tsx` | Station dropdown in EditStepDialog |
| `frontend/packages/components/industry/ProductionPlansList.tsx` | Default station dropdowns in CreatePlanDialog |
| `frontend/packages/pages/stations.tsx` | Page wrapper |
| `frontend/pages/api/stations/user-stations.ts` | API proxy (GET + POST) |
| `frontend/pages/api/stations/user-stations/[id].ts` | API proxy (PUT + DELETE) |
| `frontend/pages/api/stations/parse-scan.ts` | API proxy (POST) |
| `frontend/pages/api/industry/plans/hangars.ts` | API proxy for hangars endpoint |
| `migrations/20260222220229_add_step_output_location.up.sql` | Output location columns on plan steps |
| `frontend/packages/client/data/models.ts` | HangarsResponse type + enriched source/output fields on ProductionPlanStep |
