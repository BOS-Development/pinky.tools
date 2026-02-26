# Production Plan Runs

## Overview

Plan runs track each execution of a production plan. When a user generates jobs from a plan, a run record is created and all generated job queue entries are linked to that run and their originating plan step. This provides progress tracking at the plan execution level.

## Status

- **Phase 1**: Complete — run creation, per-plan listing, run detail with jobs, delete
- **Phase 2**: Complete — cross-plan runs list page, cancel plan run (cancels planned jobs only)

## How It Works

### Data Flow

1. User calls `POST /v1/industry/plans/{id}/generate` with `{ "quantity": N }`
2. A `production_plan_run` is created with the plan ID, user ID, and requested quantity
3. The plan step tree is walked depth-first, tracking dependency depth
4. Identical jobs are merged (same blueprint, activity, ME, TE, tax) with runs/costs summed
5. Each `industry_job_queue` entry is created with `plan_run_id`, `plan_step_id`, `sort_order`, station/location context
6. If the plan has transport settings and steps span multiple stations, transport jobs are auto-generated and linked via `transport_job_id`
7. The response includes the run object, created jobs, skipped items, and any transport jobs

### Run Status Derivation

Run status is not stored — it is derived at query time from the statuses of its linked jobs:

| Condition | Status |
|-----------|--------|
| No jobs linked | `pending` |
| All jobs completed or cancelled | `completed` |
| Any job active or completed (but not all done) | `in_progress` |
| All jobs still planned | `pending` |

### Job Summary

Each run includes a `jobSummary` with counts by status:
- `total` — total jobs in this run
- `planned` — waiting to be started
- `active` — matched to an ESI job
- `completed` — ESI job delivered
- `cancelled` — manually cancelled

## Database Schema

### `production_plan_runs` (new table)

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial | Primary key |
| `plan_id` | bigint | FK → `production_plans` ON DELETE CASCADE |
| `user_id` | bigint | FK → `users` |
| `quantity` | int | Requested production quantity |
| `created_at` | timestamptz | When this run was generated |

### `industry_job_queue` (added columns)

| Column | Type | Description |
|--------|------|-------------|
| `plan_run_id` | bigint | FK → `production_plan_runs` ON DELETE SET NULL |
| `plan_step_id` | bigint | Soft reference to `production_plan_steps` (no FK) |

`plan_step_id` has no foreign key constraint because plan steps can be deleted and recreated between runs.

## API

### `GET /v1/industry/plans/{id}/runs`

Lists all runs for a production plan with derived status and job summary.

**Response:**
```json
[
  {
    "id": 1,
    "planId": 5,
    "userId": 100,
    "quantity": 10,
    "createdAt": "2026-02-22T23:00:00Z",
    "planName": "Rifter Plan",
    "productName": "Rifter",
    "status": "in_progress",
    "jobSummary": {
      "total": 5,
      "planned": 2,
      "active": 2,
      "completed": 1,
      "cancelled": 0
    }
  }
]
```

### `GET /v1/industry/plans/{id}/runs/{runId}`

Returns a single run with its full job list.

**Response:** Same as above, plus a `jobs` array with full `IndustryJobQueueEntry` objects.

### `DELETE /v1/industry/plans/{id}/runs/{runId}`

Deletes a run. Jobs survive but lose their `plan_run_id` link (ON DELETE SET NULL).

**Response:**
```json
{ "status": "deleted" }
```

### `GET /v1/industry/plans/runs`

Lists all runs across all plans for the user, with derived status and job summary. Same response format as per-plan listing.

### `POST /v1/industry/plans/runs/{runId}/cancel`

Cancels all `planned` jobs in a run. Active and completed jobs are not affected.

**Response:**
```json
{ "status": "cancelled", "jobsCancelled": 3 }
```

### `POST /v1/industry/plans/{id}/generate` (modified)

Now returns a `run` object in the response alongside `created`, `skipped`, and `transportJobs`:

```json
{
  "run": { "id": 1, "planId": 5, "userId": 100, "quantity": 10, ... },
  "created": [...],
  "skipped": [...],
  "transportJobs": [...]
}
```

`transportJobs` is populated when the plan has `transport_fulfillment` set and steps span multiple stations. Each transport job includes items, route info, and cost estimates. See [transportation.md](../transportation.md) for the full transport job schema.

## File Structure

| File | Purpose |
|------|---------|
| `internal/database/migrations/*_plan_runs.up.sql` | Migration: new table + ALTER job queue |
| `internal/models/models.go` | `ProductionPlanRun`, `PlanRunJobSummary` structs |
| `internal/repositories/planRuns.go` | Create, GetByPlan, GetByUser, GetByID, Delete, CancelPlannedJobs |
| `internal/repositories/jobQueue.go` | Added `plan_run_id`/`plan_step_id` to all queries |
| `internal/controllers/productionPlans.go` | New handlers + modified GenerateJobs |
| `cmd/industry-tool/cmd/root.go` | Wire up PlanRuns repository |
| `frontend/pages/api/industry/plans/[id]/runs/index.ts` | GET proxy |
| `frontend/pages/api/industry/plans/[id]/runs/[runId].ts` | GET + DELETE proxy |
| `frontend/pages/api/industry/plans/runs.ts` | GET all runs proxy |
| `frontend/pages/api/industry/plans/runs/[runId]/cancel.ts` | POST cancel proxy |
| `frontend/packages/components/industry/PlanRunsList.tsx` | Plan runs list page component |
| `frontend/packages/pages/plan-runs.tsx` | Page wrapper |
| `frontend/pages/plan-runs.tsx` | Route |

## Key Decisions

1. **No stored status**: Run status is derived from job statuses at query time using SQL CASE expressions. This avoids stale data and complex synchronization.
2. **Soft reference for `plan_step_id`**: No FK constraint to `production_plan_steps` because steps can be deleted/rebuilt between runs of the same plan.
3. **ON DELETE SET NULL for `plan_run_id`**: If a run is deleted, jobs survive but lose their link. This is safer than cascading deletes of job entries.
4. **LATERAL join for job counts**: Uses PostgreSQL LATERAL subquery for efficient per-run aggregation.
5. **Cancel only planned jobs**: The cancel action targets only `planned` status jobs. Active jobs are already submitted in-game and cannot be recalled.
6. **Bulk cancel via single UPDATE**: `CancelPlannedJobs` uses a single UPDATE statement rather than iterating individual jobs, for atomic bulk cancellation.
