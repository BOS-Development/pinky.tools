# Plan Preview Endpoint

## Status

Implemented.

## Overview

`POST /v1/industry/plans/{id}/preview` simulates production job assignment at every parallelism level (1 character, 2 characters, … N characters) and returns estimated wall-clock durations. Nothing is written to the database.

## Request

```json
{ "quantity": 10 }
```

- `quantity` must be positive (same semantics as the generate endpoint)

## Response

```json
{
  "options": [
    {
      "parallelism": 1,
      "estimatedDurationSec": 86400,
      "estimatedDurationLabel": "1d 0h",
      "characters": [
        {
          "characterId": 201,
          "name": "Alpha",
          "jobCount": 2,
          "durationSec": 86400,
          "mfgSlotsUsed": 0,
          "mfgSlotsMax": 11,
          "reactSlotsUsed": 0,
          "reactSlotsMax": 0
        }
      ]
    },
    {
      "parallelism": 2,
      "estimatedDurationSec": 43200,
      "estimatedDurationLabel": "12h 0m",
      "characters": [ ... ]
    }
  ],
  "eligibleCharacters": 2,
  "totalJobs": 2
}
```

## Key Decisions

### Shared walkAndMergeSteps helper

The tree walk + merge logic was extracted from `GenerateJobs` into `walkAndMergeSteps` (a method on `ProductionPlans`). Both `GenerateJobs` and `PreviewPlan` call this helper. The helper returns a `walkResult` struct containing:

- `mergedJobs` — depth-sorted `[]*pendingJob`
- `stepProduction` — per-step production data (used by transport generation)
- `stepDepths` — per-step depth (used by transport ordering)
- `skipped` — steps that could not be walked

### pendingJob extended with TE recalculation fields

`pendingJob` was moved to file scope and gained fields for per-character TE recalculation in the preview simulation:

```go
baseBlueprintTime int    // base seconds from SDE blueprint
activity          string // "manufacturing" or "reaction"
structure         string
rig               string
security          string
blueprintTE       int
```

These allow `simulateAssignment` to recalculate each job's duration using a specific character's actual skills rather than the step's default skill levels.

### simulateAssignment algorithm

1. Take the first `parallelism` characters from the sorted capacity list.
2. Snapshot initial available slot counts (MfgSlotsAvailable / ReactSlotsAvailable) per character.
3. Clone working copies of slot availability from the snapshots.
4. Track the current depth level being processed.
5. For each merged job (deepest depth first):
   - **Depth transition**: When the job's depth differs from the current depth, reset working slot counts back to the initial snapshots. This models EVE's sequential depth behavior — children must finish before parents start, so slots used at a deeper level are free again at shallower levels.
   - Find characters with a free slot for the activity.
   - If no character has a free slot: the job is still included in results with `characterID = 0` (unassigned) and counted toward `unassignedCount`. During generation, these jobs are created in the queue with `character_id = NULL`.
   - Split runs evenly across available characters (`ceil(runs / charCount)`, last gets remainder).
   - Recalculate duration per character using their actual skills:
     - Manufacturing: `ComputeManufacturingTE` → `ComputeSecsPerRun`
     - Reaction: `ComputeTEFactor` → `ComputeSecsPerRun`
   - Consume one slot per character used.

### estimateWallClock algorithm

Depth-aware LPT (Longest Processing Time) scheduling:

1. Collect unique depth levels, sorted descending (leaves first).
2. For each depth level:
   - Group assigned jobs by character.
   - Per character: distribute their jobs across their slots using LPT lane assignment.
   - Character's contribution = max lane time.
   - Depth time = max across all characters at this depth.
3. Total wall-clock = sum of all depth times (depths are sequential — deeper items must finish before their parents start).

### Slot usage is not considered for preview

`BuildCharacterCapacities` is called with `nil` for the slot-usage map. The preview assumes characters start with fresh slots (no currently-running ESI jobs factored in). This is intentional — it shows theoretical throughput given fully available characters.

## File Paths

| File | Change |
|------|--------|
| `internal/models/models.go` | Added `PlanPreviewResult`, `PlanPreviewOption`, `PreviewCharacterInfo`, `FormatDurationLabel` |
| `internal/controllers/productionPlans.go` | Added `ProductionPlansCharacterSkillsRepository` interface; `skillsRepo` field; `walkAndMergeSteps` helper; file-scope `pendingJob` (extended), `mergeKey`, `walkResult`, `assignedJob`; `PreviewPlan` handler; `simulateAssignment`; `estimateWallClock`; updated `NewProductionPlans` constructor and route registration |
| `internal/controllers/productionPlans_test.go` | Added `MockProductionPlansCharacterSkillsRepository`; updated `productionPlanMocks` and `setupProductionPlansController`; added `Test_ProductionPlans_PreviewPlan_Success`, `_NoCharacters`, `_InvalidQuantity` |
| `cmd/industry-tool/cmd/root.go` | Passed `characterSkillsRepository` as final arg to `NewProductionPlans` |
