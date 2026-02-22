# Industry Job Manager — Phase 1

## Overview

Track and manage EVE Online industry jobs (manufacturing, reactions, invention) across all characters. Phase 1 provides character skills syncing, ESI industry job tracking, a manufacturing cost calculator, and a job queue with automatic matching to ESI jobs.

## Status

- **Phase 1**: In progress
- **Phase 2** (planned): Stockpile integration — mark markers as "production sourced", auto-generate queue entries from stockpile deficits
- **Phase 3** (planned): Multi-alt parallel planning with skill-based character assignment
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
| blueprint_type_id | bigint | Blueprint type |
| activity | text | manufacturing/reaction/invention/copying |
| runs | int | Number of runs |
| me_level | int | Blueprint ME (0-10) |
| te_level | int | Blueprint TE (0-20) |
| status | text | planned/active/completed/cancelled |
| esi_job_id | bigint | Linked ESI job (when active) |
| estimated_cost | float8 | Calculated total cost |
| estimated_duration | int | Calculated duration (seconds) |

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
