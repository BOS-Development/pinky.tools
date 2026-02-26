# Batch Configure Production Plan Steps

## Overview

The Batch Configure tab in the Production Plan Editor groups production steps that produce the same item (same `productTypeId` + `activity`) and allows configuring all instances at once. This simplifies setup for production chains where the same material (e.g., fuel blocks) appears in multiple places in the tree.

## Status

- **Phase**: Implemented
- **Branch**: `feature/production-plans`

## How It Works

### Grouping Logic

Steps are grouped by `productTypeId + activity`. For each group, the UI shows:
- Product icon, name, and activity type
- Count of steps in the group
- Summary values: if all steps share the same value, it's shown directly; if they differ, a "Mixed" chip is displayed

### Batch Edit

Clicking "Edit" on a group opens a dialog identical to the single-step EditStepDialog, with the same fields:
- ME/TE levels, skills, structure, rig, security, facility tax
- Preferred station selection (auto-populates structure/rig/security/tax)
- Input location (owner, division, container)
- Output location (owner, division, container)

On save, all steps in the group are updated in a single batch API call.

### Material Toggle (Buy / Produce)

Each group row can be expanded to reveal the materials required by that step. For each material:

- **Status detection**: The component checks all steps in the group for child steps producing that material, yielding one of three statuses:
  - `all` — every step in the group has a child step producing this material → shown as "Produce" chip
  - `none` — no step in the group has a child step for this material → shown as "Buy" chip
  - `mixed` — some steps produce it, others don't → shown as "Mixed" chip
- **Toggle action**: Clicking the toggle button switches between buy and produce for **all** steps in the group simultaneously:
  - If status is `all` or `mixed`: deletes all child steps for that material across the group (switches to Buy)
  - If status is `none`: creates child steps for all steps that don't have one (switches to Produce)
- Only materials with a blueprint (`hasBlueprint: true`) show a toggle button
- Materials are fetched lazily on first expand from the first step's materials endpoint
- Status is recomputed when `plan.steps` changes (e.g., after a toggle action)

## API

### `PUT /v1/industry/plans/{id}/steps/batch`

Batch updates multiple steps with the same parameters.

**Request body:**
```json
{
  "step_ids": [10, 20, 30],
  "me_level": 10,
  "te_level": 20,
  "industry_skill": 5,
  "adv_industry_skill": 5,
  "structure": "raitaru",
  "rig": "t2",
  "security": "high",
  "facility_tax": 1.0,
  "user_station_id": 5,
  "source_owner_type": "corporation",
  "source_owner_id": 2001,
  "source_division_number": 2,
  "source_container_id": null,
  "source_location_id": 60003760,
  "output_owner_type": "corporation",
  "output_owner_id": 2001,
  "output_division_number": 3,
  "output_container_id": null
}
```

**Response:**
```json
{
  "status": "updated",
  "rows_affected": 3
}
```

## File Structure

| File | Purpose |
|------|---------|
| `internal/repositories/productionPlans.go` | `BatchUpdateSteps` method — transaction-based batch UPDATE with `WHERE s.id = ANY($1)` |
| `internal/controllers/productionPlans.go` | `BatchUpdateSteps` handler, request validation, route at `/v1/industry/plans/{id}/steps/batch` |
| `frontend/pages/api/industry/plans/[id]/steps/batch.ts` | Next.js API proxy |
| `frontend/packages/components/industry/ProductionPlanEditor.tsx` | Tab UI (Step Tree / Batch Configure) |
| `frontend/packages/components/industry/BatchConfigureTab.tsx` | Grouping table + BatchEditStepDialog |

## Key Decisions

1. **Route ordering**: The `/steps/batch` route is registered before `/steps/{stepId}` to prevent Gorilla mux from matching "batch" as a step ID.
2. **Transaction**: Batch update uses a database transaction to ensure all-or-nothing semantics.
3. **Ownership verification**: The batch update query JOINs through `production_plans` to verify the requesting user owns the plan, same as the single-step update.
4. **No partial updates**: All fields are set for all steps — there's no "only update ME/TE" mode. The dialog pre-populates from the first step in the group.
