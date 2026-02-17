# SDE Import — Full Static Data Export from CCP

## Context

The project currently uses FuzzWorks (a community mirror) to fetch EVE static data (types, regions, systems, stations). This is limited — it only provides basic fields and depends on a third party. Importing the full SDE directly from CCP gives us:
- Richer data (blueprints, reactions, dogma, market groups, PI schematics, etc.)
- Single authoritative source with automatic freshness detection via checksum
- Foundation for the reactions calculator and future tools

This PR imports the **entire** EVE SDE YAML and replaces FuzzWorks as the data source for existing tables.

---

## Approach: FuzzWorks Replacement

The `asset_item_types` table is referenced in ~35 query locations across 10+ repositories. To avoid a massive query rewrite:

1. **Enrich existing tables** — add SDE columns to `asset_item_types` via ALTER TABLE (group_id, packaged_volume, mass, etc.)
2. **Keep existing table names** for regions, constellations, solar_systems, stations
3. **Populate from SDE** instead of FuzzWorks
4. **Delete FuzzWorks client** entirely

All other SDE data (blueprints, dogma, market groups, etc.) goes into new `sde_*` prefixed tables.

---

## Phase 1: Database Migrations

### Migration 1: `enrich_existing_tables`
Add SDE columns to `asset_item_types`:
- `group_id bigint`, `packaged_volume double precision`, `mass double precision`
- `capacity double precision`, `portion_size int`, `published boolean`
- `market_group_id bigint`, `graphic_id bigint`, `race_id bigint`, `description text`

Add `adjusted_price` to `market_prices`.

### Migration 2: `create_sde_core_tables`
**Reference data:**
- `sde_categories` (category_id PK, name, published, icon_id)
- `sde_groups` (group_id PK, name, category_id, published, icon_id)
- `sde_meta_groups` (meta_group_id PK, name)
- `sde_market_groups` (market_group_id PK, parent_group_id, name, description, icon_id, has_types)
- `sde_icons` (icon_id PK, description)
- `sde_graphics` (graphic_id PK, description)
- `sde_metadata` (key PK, value, updated_at)

### Migration 3: `create_sde_blueprint_tables`
- `sde_blueprints` (blueprint_type_id PK, max_production_limit)
- `sde_blueprint_activities` (blueprint_type_id + activity PK, time)
- `sde_blueprint_materials` (blueprint_type_id + activity + type_id PK, quantity)
- `sde_blueprint_products` (blueprint_type_id + activity + type_id PK, quantity, probability)
- `sde_blueprint_skills` (blueprint_type_id + activity + type_id PK, level)

### Migration 4: `create_sde_dogma_tables`
- `sde_dogma_attribute_categories` (category_id PK, name, description)
- `sde_dogma_attributes` (attribute_id PK, name, description, default_value, display_name, category_id, high_is_good, stackable, published, unit_id)
- `sde_dogma_effects` (effect_id PK, name, description, display_name, category_id)
- `sde_type_dogma_attributes` (type_id + attribute_id PK, value)
- `sde_type_dogma_effects` (type_id + effect_id PK, is_default)

### Migration 5: `create_sde_npc_tables`
- `sde_factions` (faction_id PK, name, description, corporation_id, icon_id)
- `sde_npc_corporations` (corporation_id PK, name, faction_id, icon_id)
- `sde_npc_corporation_divisions` (corporation_id + division_id PK, name)
- `sde_agents` (agent_id PK, name, corporation_id, division_id, level)
- `sde_agents_in_space` (agent_id PK, solar_system_id)
- `sde_races` (race_id PK, name, description, icon_id)
- `sde_bloodlines` (bloodline_id PK, name, race_id, description, icon_id)
- `sde_ancestries` (ancestry_id PK, name, bloodline_id, description, icon_id)

### Migration 6: `create_sde_industry_tables`
- `sde_planet_schematics` (schematic_id PK, name, cycle_time)
- `sde_planet_schematic_types` (schematic_id + type_id PK, quantity, is_input)
- `sde_control_tower_resources` (control_tower_type_id + resource_type_id PK, purpose, quantity, min_security, faction_id)
- `industry_cost_indices` (system_id + activity PK, cost_index, updated_at)

### Migration 7: `create_sde_misc_tables`
- `sde_skins` (skin_id PK, type_id, material_id)
- `sde_skin_licenses` (license_type_id PK, skin_id, duration)
- `sde_skin_materials` (skin_material_id PK, name)
- `sde_certificates` (certificate_id PK, name, description, group_id)
- `sde_landmarks` (landmark_id PK, name, description)
- `sde_station_operations` (operation_id PK, name, description)
- `sde_station_services` (service_id PK, name, description)
- `sde_contraband_types` (faction_id + type_id PK, standing_loss, fine_by_value)
- `sde_research_agents` (agent_id + type_id PK)
- `sde_character_attributes` (attribute_id PK, name, description, icon_id)
- `sde_corporation_activities` (activity_id PK, name)
- `sde_tournament_rule_sets` (rule_set_id PK, data jsonb)

---

## Phase 2: Models

**Update `internal/models/models.go`:**
- Enrich `EveInventoryType` with: GroupID, PackagedVolume, Mass, Capacity, PortionSize, Published, MarketGroupID, GraphicID, RaceID, Description
- Add `SdeMetadata` struct
- Add structs for each new table group (SdeCategory, SdeGroup, SdeBlueprint, SdeBlueprintActivity, SdeBlueprintMaterial, SdeBlueprintProduct, SdeBlueprintSkill, SdeDogmaAttribute, SdeDogmaEffect, etc.)

---

## Phase 3: SDE Client

**New file: `internal/client/sdeClient.go`**

```go
type SdeClient struct {
    httpClient HTTPDoer
    baseURL    string // https://developers.eveonline.com/static-data/
}
```

**Methods:**
- `GetChecksum(ctx) (string, error)` — fetch checksum file from CCP, return hash
- `DownloadSDE(ctx) (string, error)` — download ZIP to temp file, return file path
- `ParseSDE(zipPath string) (*SdeData, error)` — open ZIP, parse all YAML files

**SDE ZIP handling:**
1. Download to OS temp directory
2. Open with `archive/zip`
3. For each known filename → find ZIP entry → parse with `gopkg.in/yaml.v3`
4. Unknown files → log warning, skip (forward-compatible)
5. Clean up temp file after parsing

**YAML parse structs** (internal to client package):

typeIDs.yaml → `map[int64]SdeTypeYAML` with fields: name (localized map), description, groupID, volume, packagedVolume, mass, capacity, portionSize, published, marketGroupID, iconID, graphicID, raceID

blueprints.yaml → `map[int64]SdeBlueprintYAML` with activities map containing: manufacturing, reaction, invention, copying, researching_material_efficiency, researching_time_efficiency. Each activity has time, materials[], products[], skills[].

groupIDs.yaml, categoryIDs.yaml, marketGroups.yaml, dogmaAttributes.yaml, dogmaEffects.yaml, etc. → similar map[int64] patterns.

**Return type** `SdeData` holds all parsed data organized by table group.

---

## Phase 4: ESI Client Extensions

**Update `internal/client/esiClient.go`** — add two public ESI endpoints (no auth required):

- `GetCcpMarketPrices(ctx) ([]*CcpMarketPrice, error)` — `GET /latest/markets/prices/`
  - Returns `type_id`, `adjusted_price`, `average_price` per item
- `GetIndustryCostIndices(ctx) ([]*IndustryCostIndexSystem, error)` — `GET /latest/industry/systems/`
  - Returns `solar_system_id` + array of `{activity, cost_index}` pairs

Both use the `httpClient` (not OAuth) since they're public.

---

## Phase 5: Repositories

**New: `internal/repositories/sdeData.go`**
- Bulk upsert methods for each table group (categories, groups, blueprints, dogma, NPCs, PI, misc)
- All use transactions with `defer tx.Rollback()`
- For large datasets (blueprints, dogma): truncate + bulk insert (full refresh is simpler and safer than ON CONFLICT for many-column composite keys)
- `GetMetadata(ctx, key)` / `SetMetadata(ctx, key, value)` — checksum tracking

**New: `internal/repositories/industryCostIndices.go`**
- `UpsertIndices(ctx, indices)` — truncate + insert
- `GetCostIndex(ctx, systemID, activity)` — single lookup
- `GetLastUpdateTime(ctx)` — staleness check

**Update: `internal/repositories/itemType.go`**
- Add new columns to `UpsertItemTypes` query (group_id, packaged_volume, mass, etc.)
- Update `EveInventoryType` model usage

**Update: `internal/repositories/marketPrices.go`**
- `UpsertAdjustedPrices(ctx, map[int64]float64)` — update adjusted_price column

---

## Phase 6: Updaters

**New: `internal/updaters/sde.go`**

```
Update(ctx) flow:
1. GetChecksum from CCP
2. Compare with sde_metadata["checksum"]
3. If unchanged → return (already current)
4. Download ZIP → parse all YAMLs → SdeData
5. Transform SdeData → model structs
6. Populate existing tables: asset_item_types, regions, constellations, solar_systems, stations
7. Populate all sde_* tables
8. Update sde_metadata["checksum"]
9. Clean up temp files
```

**New: `internal/updaters/ccpPrices.go`**
- Staleness check (1h), fetch CCP adjusted prices, upsert to market_prices.adjusted_price

**New: `internal/updaters/industryCostIndices.go`**
- Staleness check (1h), fetch from ESI, flatten to rows, upsert

**Update/Delete: `internal/updaters/static.go`**
- Remove FuzzWorks dependency
- Either: delegate to SDE updater for types/regions/systems/stations
- Or: delete entirely if SDE updater handles everything

---

## Phase 7: Runners

Three new runners, each following the `MarketPricesRunner` pattern with testable `TickerFactory`:

- `internal/runners/sde.go` — SDE refresh, **24h** default interval
- `internal/runners/ccpPrices.go` — CCP adjusted prices, **1h** default interval
- `internal/runners/industryCostIndices.go` — cost indices, **1h** default interval

---

## Phase 8: Wire Up + Remove FuzzWorks

**Update: `cmd/industry-tool/cmd/root.go`**
- Create `SdeClient`, `SdeData` repository, `IndustryCostIndices` repository
- Create SDE updater, CCP prices updater, cost indices updater
- Create 3 new runners via `group.Go()`
- Remove FuzzWorks client creation
- Remove or update static updater references

**Delete:**
- `internal/client/fuzzWorks.go`
- `internal/client/fuzzWorks_test.go`
- `internal/client/fuzzWorks_mock_test.go`

---

## Files Summary

| File | Action |
|------|--------|
| `internal/database/migrations/` (7 pairs) | New — enrich tables + all SDE tables |
| `internal/models/models.go` | Update — enrich EveInventoryType + add ~30 SDE structs |
| `internal/client/sdeClient.go` | **New** — download, checksum, YAML parsing |
| `internal/client/esiClient.go` | Update — add 2 public ESI methods |
| `internal/client/fuzzWorks.go` + tests | **Delete** |
| `internal/repositories/sdeData.go` | **New** — all SDE table CRUD |
| `internal/repositories/industryCostIndices.go` | **New** |
| `internal/repositories/itemType.go` | Update — new columns |
| `internal/repositories/marketPrices.go` | Update — adjusted_price |
| `internal/updaters/sde.go` | **New** — SDE refresh orchestrator |
| `internal/updaters/ccpPrices.go` | **New** — CCP adjusted price refresh |
| `internal/updaters/industryCostIndices.go` | **New** — cost index refresh |
| `internal/updaters/static.go` | Update or delete |
| `internal/runners/sde.go` | **New** — 24h runner |
| `internal/runners/ccpPrices.go` | **New** — 1h runner |
| `internal/runners/industryCostIndices.go` | **New** — 1h runner |
| `cmd/industry-tool/cmd/root.go` | Update — wire new components, remove FuzzWorks |

---

## Verification

1. **SDE client unit tests** — mock HTTP, verify YAML parsing produces correct model structs
2. **Repository integration tests** — testcontainers, verify round-trip upsert/query for SDE tables
3. **Backend tests**: `make test-backend`
4. **Manual verification**:
   - `make dev` → check logs for "SDE update" on startup
   - Verify `asset_item_types` has `group_id` populated
   - Verify `sde_blueprints` has manufacturing + reaction entries
   - Verify `sde_groups`, `sde_categories`, `sde_dogma_attributes` populated
   - Verify existing features (inventory, marketplace, stockpiles) still work
5. **Freshness detection** — run twice, second run should skip with "SDE already up to date"
