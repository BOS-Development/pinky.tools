# Industry Job Manager — Phase 1

## Overview

Track and manage EVE Online industry jobs (manufacturing, reactions, invention) across all characters. Phase 1 provides character skills syncing, ESI industry job tracking, a manufacturing cost calculator, and a job queue with automatic matching to ESI jobs.

## Status

- **Phase 1**: Complete — character skills syncing, ESI job tracking, manufacturing calculator, job queue with auto-matching
- **Phase 2**: Complete — production plans with full production chain tree, step management, job generation
- **Phase 3**: Complete — multi-alt parallel planning with skill-based character assignment, preview endpoint, manual reassignment
- **Phase 4** (planned): Blueprint auto-detection via `GET /characters/{id}/blueprints/`

---

## How It Works

### Data Flow

```
ESI API                     Background Runners              Database
────────                    ──────────────────              ────────
/characters/{id}/skills  →  CharacterSkillsRunner (6h)   →  character_skills
/characters/{id}/industry/jobs → IndustryJobsRunner (10m) → esi_industry_jobs
                                     ↓
                              Queue Matching
                              (planned → active → completed)
                                     ↓
                              industry_job_queue

Frontend                    Backend Controller              Database
────────                    ──────────────────              ────────
/industry page           →  GET /v1/industry/jobs         →  esi_industry_jobs
                            GET /v1/industry/queue         →  industry_job_queue
                            POST /v1/industry/queue        →  (create planned job)
                            POST /v1/industry/calculate    →  (manufacturing calc)
                            GET /v1/industry/blueprints    →  sde_blueprint_*
                            GET /v1/industry/systems       →  industry_cost_indices
```

### Background Runners

| Runner | Interval | Default | Env Var | Description |
|--------|----------|---------|---------|-------------|
| Character Skills | 6h | 21600s | `SKILLS_UPDATE_INTERVAL_SEC` | Sync character skills from ESI |
| Industry Jobs | 10m | 600s | `INDUSTRY_JOBS_UPDATE_INTERVAL_SEC` | Sync ESI jobs + queue matching |

### Queue Matching Logic

The industry jobs runner matches planned queue entries to active ESI jobs:
1. Get `planned` queue entries for user
2. Get `active` ESI jobs for user
3. Match by `(blueprint_type_id, activity, runs)` where `esi.start_date > queue.created_at`
4. Link match: set `queue.esi_job_id`, status → `active`
5. For already-linked entries: if ESI job `status = 'delivered'`, mark queue → `completed`

Activity ID mapping: manufacturing=1, TE research=3, ME research=4, copying=5, invention=8, reaction=9

---

## Key Decisions

- **ME/TE are user-provided** — ESI industry jobs endpoint doesn't return blueprint ME/TE. Phase 1 uses manual input; blueprint auto-detection in Phase 4.
- **ESI auth pattern** — Token refresh in updater, pass `token string` to ESI client methods (same as PI pattern).
- **Manufacturing calculator** — Reuses exported helpers from reactions calculator (`ComputeBatchQty`, `GetPrice`, rig/security multipliers).
- **Structure TE** — Engineering complexes (Raitaru/Azbel/Sotiyo) provide 1% TE bonus for manufacturing (vs 25% for Tatara reactions).

### Manufacturing Formulas

```
ME Factor = (1 - blueprint_me/100) * (1 - rig_me * security_mult)
TE Factor = (1 - blueprint_te/100) * (1 - industry*0.04) * (1 - adv_industry*0.03) * (1 - structure_te) * (1 - rig_te * sec_mult)
Job Cost  = EIV * (cost_index + SCC_surcharge_0.04 + facility_tax/100)
```

---

## Database Schema

### character_skills

| Column | Type | Description |
|--------|------|-------------|
| character_id | bigint | PK (with skill_id) |
| user_id | bigint | Owning user |
| skill_id | bigint | PK (with character_id) |
| trained_level | int | 0-5 |
| active_level | int | 0-5 |
| skillpoints | bigint | SP in skill |
| updated_at | timestamptz | Last sync |

### esi_industry_jobs

| Column | Type | Description |
|--------|------|-------------|
| job_id | bigint | PK (ESI job ID) |
| installer_id | bigint | Character who installed |
| user_id | bigint | Owning user |
| activity_id | int | 1=mfg, 3=TE, 4=ME, 5=copy, 8=inv, 9=react |
| blueprint_type_id | bigint | Blueprint type |
| runs | int | Number of runs |
| cost | float8 | Job install cost |
| status | text | active/paused/ready/delivered/cancelled |
| duration | int | Seconds |
| start_date | timestamptz | Job start |
| end_date | timestamptz | Expected completion |

### industry_job_queue

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | PK |
| user_id | bigint | Owning user |
| character_id | bigint | Assigned character (nullable) |
| blueprint_type_id | bigint | Blueprint type |
| activity | text | manufacturing/reaction/invention/copying |
| runs | int | Number of runs |
| me_level | int | Blueprint ME (0-10) |
| te_level | int | Blueprint TE (0-20) |
| system_id | bigint | Solar system (nullable) |
| facility_tax | float8 | Facility tax (default 0) |
| status | text | planned/active/completed/cancelled |
| esi_job_id | bigint | Linked ESI job (when active) |
| product_type_id | bigint | Product type ID (nullable) |
| estimated_cost | float8 | Calculated total cost |
| estimated_duration | int | Calculated duration (seconds) |
| notes | text | Optional notes |
| plan_run_id | bigint | FK → production_plan_runs (ON DELETE SET NULL) |
| plan_step_id | bigint | Soft reference to production_plan_steps (no FK) |
| transport_job_id | bigint | FK → transport_jobs — links queue entry to a transport job |
| sort_order | int | Depth-based ordering (default 0, higher = manufactured first) |
| station_name | text | Where the job runs (default '') |
| input_location | text | Source location context, e.g. "Corp > Division 1 > Container" |
| output_location | text | Output location context |
| created_at | timestamptz | |
| updated_at | timestamptz | |

---

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/industry/jobs` | User | Active ESI jobs (enriched with names) |
| GET | `/v1/industry/jobs/all` | User | All ESI jobs including completed |
| GET | `/v1/industry/queue` | User | Job queue entries |
| POST | `/v1/industry/queue` | User | Create planned job (calculates cost) |
| PUT | `/v1/industry/queue/{id}` | User | Update planned job |
| DELETE | `/v1/industry/queue/{id}` | User | Cancel planned/active job |
| POST | `/v1/industry/calculate` | Backend | Calculate manufacturing cost |
| GET | `/v1/industry/blueprints` | Backend | Search blueprints by name |
| GET | `/v1/industry/systems` | Backend | Systems with manufacturing cost indices |

---

## File Structure

### Backend

| File | Purpose |
|------|---------|
| `internal/repositories/characterSkills.go` | Character skills CRUD |
| `internal/repositories/industryJobs.go` | ESI industry jobs CRUD |
| `internal/repositories/jobQueue.go` | Job queue CRUD |
| `internal/repositories/sdeData.go` | Manufacturing blueprint/material/system queries (additions) |
| `internal/calculator/manufacturing.go` | Manufacturing ME/TE/cost calculations |
| `internal/updaters/characterSkills.go` | Skills sync from ESI |
| `internal/updaters/industryJobs.go` | Jobs sync + queue matching |
| `internal/runners/characterSkills.go` | Skills background runner |
| `internal/runners/industryJobs.go` | Jobs background runner |
| `internal/controllers/industry.go` | HTTP handlers |

### Frontend

| File | Purpose |
|------|---------|
| `pages/industry.tsx` | Page router entry |
| `packages/pages/industry.tsx` | Main page with 3 tabs |
| `packages/components/industry/ActiveJobs.tsx` | ESI jobs table |
| `packages/components/industry/JobQueue.tsx` | Queue table with cancel action |
| `packages/components/industry/AddJob.tsx` | Form: blueprint search, calculate, add to queue |
| `pages/api/industry/jobs.ts` | Proxy GET → jobs/queue endpoints |
| `pages/api/industry/queue.ts` | Proxy GET/POST → queue endpoint |
| `pages/api/industry/queue/[id].ts` | Proxy PUT/DELETE → queue entry |
| `pages/api/industry/calculate.ts` | Proxy POST → calculate endpoint |
| `pages/api/industry/blueprints.ts` | Proxy GET → blueprint search |
| `pages/api/industry/systems.ts` | Proxy GET → systems endpoint |

### Migrations

| File | Purpose |
|------|---------|
| `20260222014855_create_character_skills.up.sql` | Create character_skills table |
| `20260222014858_create_esi_industry_jobs.up.sql` | Create esi_industry_jobs table |
| `20260222014858_create_industry_job_queue.up.sql` | Create industry_job_queue table |
| `20260224104204_create_transport_tables.up.sql` | Add `transport_job_id` FK to industry_job_queue (+ transport tables) |
| `20260224205134_add_plan_transport_settings.up.sql` | Add transport columns to production_plans |
| `20260224222923_add_sort_order_to_job_queue.up.sql` | Add `sort_order`, `station_name`, `input_location`, `output_location` to job queue |

---

# Phase 2: Production Plans

## Overview

Production plans define reusable, hierarchical production chains for items. Each plan has a tree of production steps — the root step produces the final product, and child steps produce materials that would otherwise need to be purchased. Plans are dynamic: run counts are calculated at generation time based on the quantity needed.

## How It Works

### Tree Structure

```
Production Plan: "Muninn Production"
└── Muninn (manufacturing, ME10/TE20)
    ├── Rupture (manufacturing, ME10/TE20) ← produced
    │   ├── Tritanium ← buy
    │   ├── Pyerite ← buy
    │   └── ...
    ├── Plasma Thruster (manufacturing) ← produced
    │   └── ...
    └── Morphite ← buy (no child step = purchased)
```

- **Root step** (`parent_step_id IS NULL`): the final product
- **Child steps**: materials the user chose to produce rather than buy
- Each material with a blueprint can be toggled between "buy" and "produce"
- Toggling to "produce" creates a child step; toggling back deletes it (and cascading children)

### Job Generation Algorithm

Given a plan and a target quantity for the root product:

1. Calculate root runs: `runs = ceil(quantity / product_quantity_per_run)`
2. Get root materials with batch quantities (applying ME): `batch_qty = ComputeBatchQty(runs, base_qty, me_factor)`
3. For each material that has a child step (is produced):
   - `child_runs = ceil(batch_qty / child_product_quantity_per_run)`
   - Track step depth (root = 0, children = parent + 1)
   - Recurse into child step
4. Calculate cost for manufacturing steps using existing `CalculateManufacturingJob`
5. Skip steps missing required data with reason
6. **Job merging**: identical jobs (same blueprint, activity, ME, TE, facility tax) are merged — runs, costs, and durations are summed; the deepest depth is preserved for ordering
7. **Sort order**: each job gets `sort_order = depth * 2` (even numbers); transport jobs interleave at odd values (`depth * 2 - 1`)
8. **Location context**: each job captures `station_name`, `input_location`, and `output_location` from the step's station and hangar config
9. Jobs are ordered deepest-first (`ORDER BY sort_order DESC`) so leaf materials are manufactured before parents
10. **Transport auto-generation**: if the plan has `transport_fulfillment` set and steps span multiple stations, transport jobs are created between stations (see [transportation.md](../transportation.md) for details)

## Database Schema

### production_plans

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | PK |
| user_id | bigint | FK → users |
| product_type_id | bigint | Final product type |
| name | text | User-friendly name (defaults to product name) |
| notes | text | Optional |
| default_manufacturing_station_id | bigint | FK → user_stations — default station for mfg steps |
| default_reaction_station_id | bigint | FK → user_stations — default station for reaction steps |
| transport_fulfillment | text | self_haul / courier_contract / contact_haul (NULL = disabled) |
| transport_method | text | freighter / jump_freighter / dst / blockade_runner |
| transport_profile_id | bigint | FK → transport_profiles |
| courier_rate_per_m3 | numeric(12,2) | Courier ISK per m³ (default 0) |
| courier_collateral_rate | numeric(6,4) | Courier collateral % (default 0) |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### production_plan_steps

| Column | Type | Description |
|--------|------|-------------|
| id | bigserial | PK |
| plan_id | bigint | FK → production_plans ON DELETE CASCADE |
| parent_step_id | bigint | FK → self ON DELETE CASCADE (NULL for root) |
| product_type_id | bigint | What this step produces |
| blueprint_type_id | bigint | Blueprint to use |
| activity | text | manufacturing / reaction |
| me_level | int | Blueprint ME (default 10) |
| te_level | int | Blueprint TE (default 20) |
| industry_skill | int | Industry skill level (default 5) |
| adv_industry_skill | int | Adv. Industry skill (default 5) |
| structure | text | Station type (default raitaru) |
| rig | text | Rig type (default t2) |
| security | text | Security status (default high) |
| facility_tax | numeric(5,2) | Facility tax % (default 1.0) |
| user_station_id | bigint | FK → user_stations ON DELETE SET NULL |
| station_name | text | Station/structure name (nullable) |
| source_location_id | bigint | Station for inputs (nullable) |
| source_container_id | bigint | Input container/hangar (nullable) |
| source_division_number | int | Input corp division 1-7 (nullable) |
| source_owner_type | text | character / corporation (nullable) |
| source_owner_id | bigint | Input owner ID (nullable) |
| output_owner_type | text | character / corporation (nullable) |
| output_owner_id | bigint | Output owner ID (nullable) |
| output_division_number | int | Output corp division 1-7 (nullable) |
| output_container_id | bigint | Output container/hangar (nullable) |

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/industry/plans` | User | List user's plans |
| POST | `/v1/industry/plans` | User | Create plan (auto-creates root step) |
| GET | `/v1/industry/plans/{id}` | User | Get plan with full step tree |
| PUT | `/v1/industry/plans/{id}` | User | Update plan name/notes/transport settings |
| DELETE | `/v1/industry/plans/{id}` | User | Delete plan and all steps |
| POST | `/v1/industry/plans/{id}/steps` | User | Add step (toggle material to "produce") |
| PUT | `/v1/industry/plans/{id}/steps/{stepId}` | User | Update step params |
| DELETE | `/v1/industry/plans/{id}/steps/{stepId}` | User | Remove step (toggle back to "buy") |
| GET | `/v1/industry/plans/{id}/steps/{stepId}/materials` | User | Get materials with producibility info |
| POST | `/v1/industry/plans/{id}/generate` | User | Generate queue entries from plan + quantity |

## Key Decisions

- **Auto root step**: Creating a plan automatically looks up the blueprint and creates the root step, reducing friction
- **Blueprint lookup**: `GetBlueprintByProduct` prefers manufacturing over reaction via `ORDER BY CASE`
- **ME/TE defaults**: Steps default to ME10/TE20 with max skills, matching common T2 production setups
- **Cascade deletes**: Deleting a step cascades to all child steps; deleting a plan cascades to all steps
- **Material producibility**: `GetStepMaterials` joins against `sde_blueprint_products` to check if each material can be produced, and against `production_plan_steps` to check if it already has a step

## File Structure

### Backend

| File | Purpose |
|------|---------|
| `internal/repositories/productionPlans.go` | Plans + steps CRUD, GetStepMaterials |
| `internal/repositories/productionPlans_test.go` | Integration tests |
| `internal/repositories/sdeData.go` | GetBlueprintByProduct (addition) |
| `internal/controllers/productionPlans.go` | HTTP handlers + job generation |
| `internal/controllers/productionPlans_test.go` | Unit tests with mocks |

### Frontend

| File | Purpose |
|------|---------|
| `pages/production-plans.tsx` | Page router entry |
| `packages/pages/production-plans.tsx` | Page wrapper |
| `packages/components/industry/ProductionPlansList.tsx` | Plans list + create dialog |
| `packages/components/industry/ProductionPlanEditor.tsx` | Tree editor + generate dialog |
| `packages/components/industry/__tests__/ProductionPlansList.test.tsx` | List component tests |
| `packages/components/industry/__tests__/ProductionPlanEditor.test.tsx` | Editor component tests |
| `pages/api/industry/plans/index.ts` | Proxy GET/POST → plans |
| `pages/api/industry/plans/[id].ts` | Proxy GET/PUT/DELETE → plan |
| `pages/api/industry/plans/[id]/steps/index.ts` | Proxy POST → add step |
| `pages/api/industry/plans/[id]/steps/[stepId].ts` | Proxy PUT/DELETE → step |
| `pages/api/industry/plans/[id]/steps/[stepId]/materials.ts` | Proxy GET → materials |
| `pages/api/industry/plans/[id]/generate.ts` | Proxy POST → generate jobs |

### Migrations

| File | Purpose |
|------|---------|
| `20260222151815_create_production_plans.up.sql` | Create production_plans + production_plan_steps tables |

---

# Phase 3: Multi-Alt Parallel Planning

## Overview

Character assignment enables distributing production plan jobs across multiple characters (alts) for parallel manufacturing. The system auto-detects eligible characters from synced skills, provides a preview of estimated completion times at different parallelism levels, and assigns characters during job generation.

## How It Works

### Character Eligibility (Auto-Detected)

Characters are eligible based on their synced skills:
- **Manufacturing eligible**: Industry skill (3380) >= 1
- **Reaction eligible**: Reactions skill (45746) >= 1

No manual configuration needed — eligibility is derived from `character_skills` at query time.

### Slot Calculation

Manufacturing slots: `1 + Mass Production level + Adv Mass Production level` (max 11)
Reaction slots: `1 + Mass Reactions level + Adv Mass Reactions level` (requires Reactions >= 1)

Both `planned` and `active` queue entries count against a character's slot limit.

### Preview Flow

1. User enters quantity in the generate dialog
2. Clicks "Preview" → `POST /v1/industry/plans/{id}/preview`
3. Backend simulates assignment at every parallelism level (1..N eligible characters)
4. Returns estimated wall-clock time per option using depth-aware LPT scheduling
5. User selects desired parallelism level from the comparison table

### Job Generation with Assignment

When `parallelism >= 1` is provided to the generate endpoint:
1. Discover eligible characters, build skill/slot capacities
2. Walk the step tree and merge jobs (shared logic with preview via `walkAndMergeSteps`)
3. Process merged jobs deepest-first; slots recycle at each depth transition (see below)
4. Split merged job runs evenly across characters with available slots
5. Recalculate TE/duration per split using actual character skills:
   - Manufacturing: `ComputeManufacturingTE(te, charIndustry, charAdvIndustry, ...)`
   - Reaction: `ComputeTEFactor(charReactions, structure, rig, security)` (bug fix: reactions no longer use manufacturing TE formula)
6. Create queue entries with `CharacterID` set
7. If slots exhausted at a depth level, unassigned jobs are still created in the queue with `character_id = NULL`

When `parallelism = 0` (default): no character assignment (backward compatible).

### Depth-Aware Slot Recycling

Merged jobs are processed deepest-first (leaves before parents). In EVE, children must finish before parents can start, so depth levels are sequential. `simulateAssignment` resets available slot counts when it transitions to a new (shallower) depth level. This means a character with 2 manufacturing slots can handle 2 jobs at depth 4, then reuse those same 2 slots for 2 more jobs at depth 2.

Without this, a plan with many depth levels would quickly exhaust slots and leave most jobs unassigned even though the character would have free slots by the time shallower jobs are ready to start.

### Manual Reassignment

After generation, users can reassign any `planned` queue entry to a different character via the Job Queue UI. Clicking the character chip opens a menu of eligible characters with slot info.

## Key Decisions

- **No new tables**: Eligibility auto-detected from existing `character_skills`; `industry_job_queue.character_id` already existed
- **Reaction TE fix**: Reactions now correctly use `ComputeTEFactor` (Reactions skill only, 4% per level) instead of `ComputeManufacturingTE` (Industry + Adv Industry)
- **Preview shares tree walk**: `walkAndMergeSteps` is extracted as a shared helper used by both GenerateJobs and PreviewPlan
- **LPT scheduling for estimates**: Wall-clock estimates use Longest Processing Time heuristic per depth level for realistic time prediction
- **Depth-aware slot recycling**: Slots reset at each depth transition — children finish before parents start, freeing slots for reuse
- **Unassigned jobs still queued**: When slots are exhausted at a depth, jobs are created with `character_id = NULL` rather than silently dropped
- **Parallelism per-run**: Users choose parallelism at generation time (per-run), not as a plan setting

## EVE Skill IDs

| Skill | ID | Effect |
|-------|------|--------|
| Industry | 3380 | 4% mfg time reduction per level |
| Advanced Industry | 3388 | 3% mfg time reduction per level |
| Mass Production | 3387 | +1 mfg slot per level |
| Adv Mass Production | 24625 | +1 mfg slot per level |
| Reactions | 45746 | 4% reaction time reduction per level |
| Mass Reactions | 45748 | +1 reaction slot per level |
| Adv Mass Reactions | 45749 | +1 reaction slot per level |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/industry/plans/{id}/preview` | Preview estimated duration at all parallelism levels |
| GET | `/v1/industry/character-slots` | Slot summary for all eligible characters |
| PUT | `/v1/industry/queue/{id}/character` | Reassign a queue entry to a different character |

### Generate endpoint change

`POST /v1/industry/plans/{id}/generate` now accepts optional `parallelism` in the request body. Response includes `characterAssignments` and `unassignedCount`.

## File Structure

### Backend

| File | Purpose |
|------|---------|
| `internal/calculator/slots.go` | Skill ID constants, slot calculation, `BuildCharacterCapacities` |
| `internal/calculator/slots_test.go` | Unit tests for slot calculator |
| `internal/controllers/productionPlans.go` | `PreviewPlan`, `walkAndMergeSteps`, `simulateAssignment`, `estimateWallClock` |
| `internal/controllers/industry.go` | `GetCharacterSlots`, `ReassignQueueCharacter` handlers |
| `internal/repositories/jobQueue.go` | `GetSlotUsage`, `ReassignCharacter` |

### Frontend

| File | Purpose |
|------|---------|
| `packages/components/industry/ProductionPlanEditor.tsx` | Preview table in generate dialog |
| `packages/components/industry/JobQueue.tsx` | Clickable character chip with reassignment menu |
| `pages/api/industry/plans/[id]/preview.ts` | POST proxy → preview |
| `pages/api/industry/character-slots.ts` | GET proxy → character slots |
| `pages/api/industry/queue/[id]/character.ts` | PUT proxy → reassign character |
