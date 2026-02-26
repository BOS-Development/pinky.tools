# Asset Aggregation

## Overview
Inventory items of the same type are aggregated (stacked) within each scope — hangar, container, delivery, asset safety, or corp division. Multiple ESI item stacks of the same `type_id` from the same owner at the same location are collapsed into a single row with summed quantities, volumes, and values.

## Status
- **Phase 1**: Backend SQL aggregation — COMPLETE

## Key Decisions
- Aggregation is done at the SQL level via `GROUP BY type_id` (not in Go or frontend)
- No model changes required — the `Asset` struct already has no `ItemID` field
- No frontend changes required — the UI renders whatever asset rows it receives
- Stockpile markers still work correctly since they're keyed by `(type_id, owner_id, location_id, container_id, division_number)` which matches the GROUP BY grouping
- `InjectOrphanStockpileRows()` requires no changes since it keys on `{typeID, ownerID, locationID, containerID, divisionNumber}`

## Implementation

### Modified Queries in `internal/repositories/assets.go`

Four SQL queries were updated with `GROUP BY` + `SUM()`:

1. **`hangaredItemsQuery`** — character hangar/deliveries/asset safety items
2. **`itemsInContainersQuery`** — items inside character containers
3. **`corpHangaredItemsQuery`** — corporation division items
4. **`corpItemsInContainersQuery`** — items inside corporation containers

Each query changed:
- `quantity` → `SUM(quantity)`
- `volume * quantity` → `SUM(volume * quantity)`
- `stockpile_delta` and `deficit_value` recalculated using `SUM(quantity)`
- Added `GROUP BY` on all non-aggregated columns
- Replaced `ORDER BY item_id` with `ORDER BY type_name` where applicable

### Files Modified
- `internal/repositories/assets.go` — 4 SQL queries
- `internal/repositories/assets_test.go` — 4 new test cases

### Tests
- `Test_AssetsShouldAggregateMultipleStacksInHangar`
- `Test_AssetsShouldAggregateMultipleStacksInContainers`
- `Test_AssetsShouldAggregateCorpDivisionItems`
- `Test_AggregatedAssetsWithStockpileMarkers`
