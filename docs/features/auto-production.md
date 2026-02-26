# Auto-Production for Stockpile Deficits

## Overview

Auto-Production links stockpile markers to production plans so a background runner automatically generates production plan runs to fill stockpile deficits. When a stockpile item falls below its desired quantity, the runner calculates the net deficit (gross deficit minus pending/active production output) and generates plan runs to cover it.

## Status

- **Phase**: Implementation
- **Branch**: `feature/auto-production`

## Key Decisions

1. **Net deficit dedup prevents over-production** — Before generating runs, the runner subtracts pending/active job output for the linked plan so already-in-progress production is not double-counted.
2. **Runner does NOT run on startup** — Waits for the first tick so asset and price data is populated before any plan runs are generated.
3. **Auto-buy exclusion at SQL level** — Markers with `auto_production_enabled = TRUE` are excluded from `GetStockpileDeficitsForConfig` to prevent double-sourcing the same deficit via both auto-buy and auto-production.
4. **Service extraction pattern** — Shared job generation logic lives in `internal/services/jobGeneration.go` to avoid circular dependencies between the updater and controller packages.
5. **Grouping by (userID, planID)** — Markers sharing the same user and plan are grouped; deficits are summed and `MAX(parallelism)` is used for character slot assignment.
6. **Cooldown guard** — The runner skips a plan if it was triggered within the current runner interval to prevent rapid re-queuing.
7. **ON DELETE SET NULL** — If a production plan is deleted, `plan_id` goes to NULL on all linked markers; no orphan references, auto-production silently disables itself.

## Schema

### `stockpile_markers` (MODIFIED)

Three new columns added via migration `20260225205355_auto_production`:

```sql
ALTER TABLE stockpile_markers ADD COLUMN plan_id BIGINT REFERENCES production_plans(id) ON DELETE SET NULL;
ALTER TABLE stockpile_markers ADD COLUMN auto_production_parallelism INT DEFAULT 0;
ALTER TABLE stockpile_markers ADD COLUMN auto_production_enabled BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_stockpile_markers_auto_production ON stockpile_markers(user_id) WHERE auto_production_enabled = TRUE;
```

| Column | Type | Description |
|--------|------|-------------|
| `plan_id` | BIGINT (nullable) | Foreign key to the linked production plan |
| `auto_production_parallelism` | INT | Max parallel character slots for job assignment (0 = unlimited) |
| `auto_production_enabled` | BOOLEAN | Whether auto-production is active for this marker |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/industry/plans/by-product/{typeId}` | Returns production plans for a given product type (used by frontend plan selector) |
| POST | `/v1/stockpiles` | Existing upsert endpoint — now accepts `planId`, `autoProductionParallelism`, `autoProductionEnabled` fields |

## Runner Logic

1. Query all stockpile markers where `auto_production_enabled = TRUE`
2. Group markers by `(user_id, plan_id)`
3. For each group:
   a. Sum gross deficits across all markers in the group
   b. Query `GetPendingOutputForPlan` — subtract in-progress production output
   c. Skip if net deficit <= 0
   d. Skip if last auto-production for this plan was within the runner interval (cooldown)
   e. Call job generation service with net deficit and `MAX(parallelism)` from the group
4. Runner interval is configurable via `AUTO_PRODUCTION_INTERVAL_SEC` (default 1800 = 30 minutes)

## Safety Mechanisms

1. **Net deficit dedup** — Always subtract pending/active output before generating runs
2. **Cooldown** — Skip if last auto-production for this plan was within the runner interval
3. **No startup run** — Runner waits for first tick (assets and prices must be populated first)
4. **Enabled flag** — Users can disable per-marker without unlinking the plan
5. **ON DELETE SET NULL** — If a plan is deleted, `plan_id` goes to NULL; no orphan references

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTO_PRODUCTION_INTERVAL_SEC` | `1800` | Runner interval in seconds (30 minutes) |

## File Structure

### Backend

- `internal/database/migrations/20260225205355_auto_production.up.sql` — Schema migration (add columns + index)
- `internal/database/migrations/20260225205355_auto_production.down.sql` — Rollback migration
- `internal/models/models.go` — `StockpileMarker` struct with 3 new fields (`PlanID`, `AutoProductionParallelism`, `AutoProductionEnabled`)
- `internal/repositories/stockpileMarkers.go` — `GetAutoProductionMarkers` method
- `internal/repositories/planRuns.go` — `GetPendingOutputForPlan` method
- `internal/repositories/productionPlans.go` — `GetByProductTypeAndUser` method
- `internal/repositories/autoBuyConfigs.go` — `auto_production_enabled` filter in `GetStockpileDeficitsForConfig`
- `internal/services/jobGeneration.go` — Extracted shared job generation logic
- `internal/updaters/autoProduction.go` — Core deficit-to-runs logic
- `internal/runners/autoProduction.go` — Background tick runner
- `internal/controllers/productionPlans.go` — `GetPlansByProduct` handler
- `cmd/industry-tool/cmd/settings.go` — `AutoProductionIntervalSec` setting
- `cmd/industry-tool/cmd/root.go` — Runner and service wiring

### Frontend

- `frontend/packages/client/data/models.ts` — `StockpileMarker` type with 3 new fields (`planId`, `autoProductionParallelism`, `autoProductionEnabled`)
- `frontend/pages/api/industry/plans/by-product/[typeId].ts` — API proxy route for plan selector
- `frontend/packages/components/assets/AddStockpileDialog.tsx` — Auto-production toggle, plan selector dropdown, parallelism input
- `frontend/packages/components/stockpiles/StockpilesList.tsx` — Auto-production status column
