# SDE (Static Data Export) — Reference Guide

The SDE is CCP's official dump of all EVE Online static game data. It replaces the old FuzzWorks community mirror and is the single source of truth for types, blueprints, dogma, universe geography, and more.

**Source**: https://developers.eveonline.com/static-data/

---

## How It Works

### Data Flow

```
CCP SDE ZIP (sde.zip, ~84MB YAML files)
    → SdeClient.ParseSDE() parses all YAML
    → SdeData struct (in-memory)
    → Updater upserts to PostgreSQL tables
```

### Background Runners (wired in `cmd/industry-tool/cmd/root.go`)

| Runner | Interval | Source | What It Updates |
|--------|----------|--------|-----------------|
| `SdeRunner` | 24h | CCP SDE ZIP | All `sde_*` tables + `asset_item_types`, `regions`, `constellations`, `solar_systems`, `stations` |
| `CcpPricesRunner` | 1h | ESI `GET /markets/prices/` | `market_prices.adjusted_price` |
| `CostIndicesRunner` | 1h | ESI `GET /industry/systems/` | `industry_cost_indices` |

All runners fire immediately on startup, then repeat. The SDE runner compares CCP's checksum with `sde_metadata["checksum"]` and skips if unchanged.

### Key Files

| File | Purpose |
|------|---------|
| `internal/client/sdeClient.go` | Download, checksum, YAML parsing → `SdeData` struct |
| `internal/updaters/sde.go` | Orchestrates full SDE refresh (checksum → download → parse → upsert) |
| `internal/updaters/ccpPrices.go` | Fetches CCP adjusted prices from ESI |
| `internal/updaters/industryCostIndices.go` | Fetches per-system cost indices from ESI |
| `internal/repositories/sdeData.go` | Bulk upsert methods for all `sde_*` tables |
| `internal/repositories/industryCostIndices.go` | CRUD for `industry_cost_indices` |
| `internal/runners/sde.go` | 24h background runner |
| `internal/runners/ccpPrices.go` | 1h background runner |
| `internal/runners/industryCostIndices.go` | 1h background runner |

---

## Table Reference

### Enriched Existing Tables

These tables existed before the SDE but now receive richer data from it.

**`asset_item_types`** — All EVE item types (~51K rows)
| Column | Type | Source | Notes |
|--------|------|--------|-------|
| `type_id` | BIGINT PK | SDE `typeIDs.yaml` | EVE type ID |
| `type_name` | TEXT | SDE | English name |
| `volume` | DOUBLE PRECISION | SDE | Assembled volume |
| `icon_id` | BIGINT | SDE | Icon reference |
| `group_id` | BIGINT | SDE | → `sde_groups.group_id` |
| `packaged_volume` | DOUBLE PRECISION | SDE | Repackaged volume |
| `mass` | DOUBLE PRECISION | SDE | |
| `capacity` | DOUBLE PRECISION | SDE | Cargo capacity |
| `portion_size` | INT | SDE | Manufacturing batch size |
| `published` | BOOLEAN | SDE | Visible on market |
| `market_group_id` | BIGINT | SDE | → `sde_market_groups.market_group_id` |
| `graphic_id` | BIGINT | SDE | → `sde_graphics.graphic_id` |
| `race_id` | BIGINT | SDE | → `sde_races.race_id` |
| `description` | TEXT | SDE | Item description |

**`regions`**, **`constellations`**, **`solar_systems`** — Universe geography. Populated from SDE YAML with full names.

**`stations`** — NPC stations (~5K rows). **Important**: `npcStations.yaml` does NOT contain station names. The upsert preserves existing names when the SDE provides an empty one (`CASE WHEN EXCLUDED.name = '' THEN stations.name ELSE EXCLUDED.name END`). Station names come from seed data or ESI.

**`market_prices`** — Has `adjusted_price` column updated hourly from ESI. Used for industry job cost calculations.

---

### Core Reference Tables

**`sde_categories`** — Top-level item classification (Ship, Module, Charge, Blueprint, Reaction, etc.)
| Column | Type |
|--------|------|
| `category_id` | BIGINT PK |
| `name` | TEXT |
| `published` | BOOLEAN |
| `icon_id` | BIGINT |

**`sde_groups`** — Mid-level classification within categories
| Column | Type | Notes |
|--------|------|-------|
| `group_id` | BIGINT PK | |
| `name` | TEXT | e.g., "Frigate", "Composite", "Intermediate Materials" |
| `category_id` | BIGINT | → `sde_categories` |
| `published` | BOOLEAN | |
| `icon_id` | BIGINT | |

**Hierarchy**: `sde_categories` → `sde_groups` → `asset_item_types` (via `group_id`)

**`sde_market_groups`** — Market browser tree (self-referencing)
| Column | Type | Notes |
|--------|------|-------|
| `market_group_id` | BIGINT PK | |
| `parent_group_id` | BIGINT | Self-referencing FK (NULL = root) |
| `name` | TEXT | |
| `description` | TEXT | |
| `icon_id` | BIGINT | |
| `has_types` | BOOLEAN | Leaf node with actual items |

**`sde_meta_groups`** — Tech level (Tech I, Tech II, Storyline, Faction, Officer, etc.)

**`sde_icons`**, **`sde_graphics`** — Visual asset metadata.

**`sde_metadata`** — Key-value store for tracking SDE state.
| Column | Type | Notes |
|--------|------|-------|
| `key` | TEXT PK | e.g., `"checksum"` |
| `value` | TEXT | Current SDE build hash |
| `updated_at` | TIMESTAMP | Last update time |

---

### Blueprint & Industry Tables

These are the most important tables for manufacturing and reaction calculators.

**`sde_blueprints`** — One row per blueprint type
| Column | Type |
|--------|------|
| `blueprint_type_id` | BIGINT PK |
| `max_production_limit` | INT |

**`sde_blueprint_activities`** — Activities a blueprint can perform
| Column | Type | Notes |
|--------|------|-------|
| `blueprint_type_id` | BIGINT | PK part 1 |
| `activity` | TEXT | PK part 2. Values: `manufacturing`, `reaction`, `invention`, `copying`, `researching_material_efficiency`, `researching_time_efficiency` |
| `time` | INT | Base time in seconds (e.g., 10800 = 3h for reactions) |

**`sde_blueprint_materials`** — Input materials for each activity
| Column | Type | Notes |
|--------|------|-------|
| `blueprint_type_id` | BIGINT | PK part 1 |
| `activity` | TEXT | PK part 2 |
| `type_id` | BIGINT | PK part 3 — the material type |
| `quantity` | INT | Base quantity before ME |

**`sde_blueprint_products`** — Output products for each activity
| Column | Type | Notes |
|--------|------|-------|
| `blueprint_type_id` | BIGINT | PK part 1 |
| `activity` | TEXT | PK part 2 |
| `type_id` | BIGINT | PK part 3 — the product type |
| `quantity` | INT | Output quantity per run |
| `probability` | DOUBLE PRECISION | For invention (NULL for manufacturing/reaction) |

**`sde_blueprint_skills`** — Required skills for each activity
| Column | Type | Notes |
|--------|------|-------|
| `blueprint_type_id` | BIGINT | PK part 1 |
| `activity` | TEXT | PK part 2 |
| `type_id` | BIGINT | PK part 3 — the skill type |
| `level` | INT | Required skill level |

**`industry_cost_indices`** — Per-system cost indices (refreshed hourly from ESI)
| Column | Type | Notes |
|--------|------|-------|
| `system_id` | BIGINT | PK part 1 |
| `activity` | TEXT | PK part 2. Values: `manufacturing`, `reaction`, `copying`, `invention`, etc. |
| `cost_index` | DOUBLE PRECISION | e.g., 0.0638 |
| `updated_at` | TIMESTAMP | |

---

### Dogma Tables (Item Attributes & Effects)

Dogma defines the numeric properties and effects of every item in EVE.

**`sde_dogma_attribute_categories`** — Attribute grouping (e.g., "Fitting", "Shield", "Armor")
- PK: `category_id` BIGINT
- `name`, `description` TEXT

**`sde_dogma_attributes`** — Attribute definitions (~2.8K rows)
- PK: `attribute_id` BIGINT
- `name`, `description`, `display_name` TEXT
- `default_value` DOUBLE PRECISION
- `category_id` BIGINT, `high_is_good` BOOLEAN, `stackable` BOOLEAN, `published` BOOLEAN, `unit_id` BIGINT

**`sde_dogma_effects`** — Effect definitions
- PK: `effect_id` BIGINT
- `name`, `description`, `display_name` TEXT, `category_id` BIGINT

**`sde_type_dogma_attributes`** — Per-type attribute values (the big join table)
- PK: (`type_id`, `attribute_id`)
- `value` DOUBLE PRECISION

**`sde_type_dogma_effects`** — Per-type effects
- PK: (`type_id`, `effect_id`)
- `is_default` BOOLEAN

---

### NPC & Lore Tables

**`sde_factions`** — EVE factions (Caldari State, Minmatar Republic, etc.)
- PK: `faction_id` — `name`, `description`, `corporation_id`, `icon_id`

**`sde_npc_corporations`** — NPC corps (Caldari Navy, etc.)
- PK: `corporation_id` — `name`, `faction_id`, `icon_id`

**`sde_npc_corporation_divisions`** — Corp division names
- PK: (`corporation_id`, `division_id`) — `name`

**`sde_agents`** — NPC agents
- PK: `agent_id` — `name`, `corporation_id`, `division_id`, `level`

**`sde_agents_in_space`** — Agent locations
- PK: `agent_id` — `solar_system_id`

**`sde_races`** — Playable races
- PK: `race_id` — `name`, `description`, `icon_id`

**`sde_bloodlines`** — Character bloodlines
- PK: `bloodline_id` — `name`, `race_id`, `description`, `icon_id`

**`sde_ancestries`** — Character ancestries
- PK: `ancestry_id` — `name`, `bloodline_id`, `description`, `icon_id`

---

### Planetary Interaction & POS Tables

**`sde_planet_schematics`** — PI production recipes
- PK: `schematic_id` — `name`, `cycle_time` (seconds)

**`sde_planet_schematic_types`** — PI inputs/outputs
- PK: (`schematic_id`, `type_id`) — `quantity`, `is_input` BOOLEAN

**`sde_control_tower_resources`** — POS fuel requirements
- PK: (`control_tower_type_id`, `resource_type_id`) — `purpose`, `quantity`, `min_security`, `faction_id`

---

### Misc Tables

**`sde_skins`** — Ship skins (skin_id PK, type_id, material_id)
**`sde_skin_licenses`** — Skin license items (license_type_id PK, skin_id, duration)
**`sde_skin_materials`** — Skin material definitions (skin_material_id PK, name)
**`sde_certificates`** — Certificate definitions (certificate_id PK, name, description, group_id)
**`sde_landmarks`** — Space landmarks (landmark_id PK, name, description)
**`sde_station_operations`** — Station operation types (operation_id PK, name, description)
**`sde_station_services`** — Station service types (service_id PK, name, description)
**`sde_contraband_types`** — Contraband rules (faction_id + type_id PK, standing_loss, fine_by_value)
**`sde_research_agents`** — Research agent skills (agent_id + type_id PK)
**`sde_character_attributes`** — Character attribute definitions (attribute_id PK, name, description, icon_id)
**`sde_corporation_activities`** — Corp activity types (activity_id PK, name)
**`sde_tournament_rule_sets`** — Tournament rules (rule_set_id PK, data JSONB)

---

## Common Query Patterns

### Find all reactions (for a reactions calculator)

```sql
-- All reaction blueprints with their products
SELECT
    ba.blueprint_type_id,
    ba.time,
    bp.type_id AS product_type_id,
    bp.quantity AS product_quantity,
    ait.type_name AS product_name,
    COALESCE(ait.packaged_volume, ait.volume, 0) AS product_volume,
    g.name AS group_name
FROM sde_blueprint_activities ba
JOIN sde_blueprint_products bp
    ON bp.blueprint_type_id = ba.blueprint_type_id
    AND bp.activity = ba.activity
JOIN asset_item_types ait ON ait.type_id = bp.type_id
JOIN sde_groups g ON g.group_id = ait.group_id
WHERE ba.activity = 'reaction'
ORDER BY g.name, ait.type_name;
```

### Get input materials for a reaction/blueprint

```sql
SELECT
    bm.type_id,
    ait.type_name,
    bm.quantity
FROM sde_blueprint_materials bm
JOIN asset_item_types ait ON ait.type_id = bm.type_id
WHERE bm.blueprint_type_id = ?
  AND bm.activity = 'reaction';  -- or 'manufacturing'
```

### Get manufacturing blueprint for an item

```sql
SELECT
    bp.blueprint_type_id,
    ba.time,
    bp.quantity AS output_quantity
FROM sde_blueprint_products bp
JOIN sde_blueprint_activities ba
    ON ba.blueprint_type_id = bp.blueprint_type_id
    AND ba.activity = bp.activity
WHERE bp.type_id = ?           -- the item you want to build
  AND bp.activity = 'manufacturing';
```

### Find item type with group and category

```sql
SELECT
    ait.type_id, ait.type_name, ait.volume, ait.published,
    g.name AS group_name, g.group_id,
    c.name AS category_name, c.category_id
FROM asset_item_types ait
JOIN sde_groups g ON g.group_id = ait.group_id
JOIN sde_categories c ON c.category_id = g.category_id
WHERE ait.type_name ILIKE '%Rifter%';
```

### Get dogma attributes for an item

```sql
SELECT
    da.name AS attribute_name,
    da.display_name,
    tda.value,
    da.unit_id
FROM sde_type_dogma_attributes tda
JOIN sde_dogma_attributes da ON da.attribute_id = tda.attribute_id
WHERE tda.type_id = ?
ORDER BY da.name;
```

### Jita prices (from market orders)

```sql
SELECT
    type_id,
    MIN(price) FILTER (WHERE NOT is_buy_order) AS sell_min,
    MAX(price) FILTER (WHERE is_buy_order) AS buy_max
FROM market_orders
WHERE location_id = 60003760  -- Jita 4-4
GROUP BY type_id;
```

### CCP adjusted prices (for job cost calculations)

```sql
SELECT type_id, adjusted_price FROM market_prices WHERE adjusted_price IS NOT NULL;
```

### System cost index for reactions

```sql
SELECT cost_index
FROM industry_cost_indices
WHERE system_id = ? AND activity = 'reaction';
```

### Systems with a specific activity index (for a system picker)

```sql
SELECT i.system_id, ss.name, ss.security, i.cost_index
FROM industry_cost_indices i
JOIN solar_systems ss ON ss.solar_system_id = i.system_id
WHERE i.activity = 'reaction'
ORDER BY ss.name;
```

### Identify if a material is an intermediate (reaction product)

```sql
-- Materials that are themselves produced by reactions
SELECT DISTINCT bp.type_id
FROM sde_blueprint_products bp
JOIN sde_blueprint_activities ba
    ON ba.blueprint_type_id = bp.blueprint_type_id
    AND ba.activity = bp.activity
WHERE ba.activity = 'reaction';
```

### Market group tree (for a market browser)

```sql
-- Root groups
SELECT * FROM sde_market_groups WHERE parent_group_id IS NULL ORDER BY name;

-- Children of a group
SELECT * FROM sde_market_groups WHERE parent_group_id = ? ORDER BY name;

-- Items in a leaf group
SELECT * FROM asset_item_types WHERE market_group_id = ? AND published = true ORDER BY type_name;
```

---

## Gotchas & Design Decisions

1. **Station names**: `npcStations.yaml` has NO station name field. The station upsert SQL uses `CASE WHEN EXCLUDED.name = '' THEN stations.name ELSE EXCLUDED.name END` to preserve existing names. Station names come from E2E seed data or ESI universe endpoints.

2. **Mixed YAML formats**: Some SDE YAML files have fields that can be either a plain string (`"Damage"`) or a localized map (`{en: "Damage", de: "Schaden"}`). The `localizedString` custom YAML unmarshaler in `sdeClient.go` handles both, always extracting the English value.

3. **Blueprint activities**: Reactions and manufacturing use the SAME blueprint tables — differentiated by the `activity` column. A "reaction blueprint" is a blueprint where `sde_blueprint_activities.activity = 'reaction'`.

4. **Truncate vs upsert**: Large tables (blueprints, dogma, type_dogma_*) use TRUNCATE + bulk INSERT in a transaction for full refresh. Smaller reference tables use ON CONFLICT upsert.

5. **Enriched vs new tables**: `asset_item_types` is referenced in ~35 query locations across 10+ repositories. Rather than creating a separate `sde_types` table and rewriting all queries, the SDE enriches the existing table with additional columns (group_id, mass, packaged_volume, etc.). All new columns are nullable for backward compatibility.

6. **SDE data volumes** (approximate):
   - Types: ~51K
   - Blueprints: ~5K
   - Blueprint materials: ~40K
   - Dogma attributes: ~2.8K
   - Type dogma attributes: ~800K
   - Stations: ~5K
   - Solar systems: ~8K
   - Regions: ~100
