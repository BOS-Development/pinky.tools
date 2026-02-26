# Database Schema Reference

**Last updated:** 2026-02-25
**Tables:** 68 | **Views:** 1 | **Migrations:** 87 (43 up/down pairs + 1 schema-only migration)

---

## Table of Contents

1. [Core](#1-core)
2. [Assets](#2-assets)
3. [Geography](#3-geography)
4. [SDE (Static Data Export)](#4-sde-static-data-export)
5. [Market & Pricing](#5-market--pricing)
6. [Social & Contacts](#6-social--contacts)
7. [Commerce](#7-commerce)
8. [Industry & Production](#8-industry--production)
9. [Planetary Industry](#9-planetary-industry)
10. [Transportation](#10-transportation)
11. [Notifications](#11-notifications)
12. [Views](#12-views)
13. [Relationship Map](#13-relationship-map)
14. [Index Reference](#14-index-reference)
15. [Schema Conventions](#15-schema-conventions)
16. [Migration History](#16-migration-history)

---

## 1. Core

### `users`

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | NOT NULL |
| assets_last_updated_at | TIMESTAMPTZ | nullable |

### `characters`

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGINT | PK (composite) |
| user_id | BIGINT | PK (composite), FK → users(id) |
| name | VARCHAR(500) | NOT NULL |
| esi_token | VARCHAR(5000) | NOT NULL |
| esi_refresh_token | VARCHAR(5000) | NOT NULL |
| esi_token_expires_on | TIMESTAMP | NOT NULL |
| esi_scopes | TEXT | NOT NULL DEFAULT '' |
| corporation_id | BIGINT | nullable |

Primary key: `(id, user_id)`

### `player_corporations`

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGINT | PK (composite) |
| user_id | BIGINT | PK (composite), FK → users(id) |
| name | VARCHAR(500) | NOT NULL |
| esi_token | VARCHAR(5000) | NOT NULL |
| esi_refresh_token | VARCHAR(5000) | NOT NULL |
| esi_token_expires_on | TIMESTAMP | NOT NULL |
| esi_scopes | TEXT | NOT NULL DEFAULT '' |
| alliance_id | BIGINT | nullable |
| alliance_name | VARCHAR(500) | nullable |

Primary key: `(id, user_id)`

### `corporations`

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | NOT NULL |

NPC or player corporations used for station ownership references. Not linked to `player_corporations`.

### `corporation_divisions`

Replaces the initial `corporation_hanger_divisions` table (dropped in migration `20250101110000`).

| Column | Type | Constraints |
|--------|------|-------------|
| corporation_id | BIGINT | PK (composite), FK → player_corporations(id, user_id) |
| user_id | BIGINT | PK (composite) |
| division_number | INT | PK (composite) |
| division_type | VARCHAR(20) | PK (composite) |
| name | VARCHAR(500) | NOT NULL |

Primary key: `(corporation_id, user_id, division_number, division_type)`

---

## 2. Assets

### `character_assets`

| Column | Type | Constraints |
|--------|------|-------------|
| character_id | BIGINT | PK (composite), FK → characters(id, user_id) |
| user_id | BIGINT | PK (composite) |
| item_id | BIGINT | PK (composite) |
| update_key | VARCHAR(100) | NOT NULL |
| is_blueprint_copy | BOOLEAN | NOT NULL |
| is_singleton | BOOLEAN | NOT NULL |
| location_id | BIGINT | NOT NULL |
| location_type | VARCHAR(15) | NOT NULL |
| quantity | BIGINT | NOT NULL |
| type_id | BIGINT | NOT NULL |
| location_flag | VARCHAR(50) | NOT NULL |

Primary key: `(character_id, user_id, item_id)`

### `character_asset_location_names`

Custom names for player-named containers or ships.

| Column | Type | Constraints |
|--------|------|-------------|
| character_id | BIGINT | PK (composite), FK → characters(id, user_id) |
| user_id | BIGINT | PK (composite) |
| item_id | BIGINT | PK (composite) |
| name | VARCHAR(500) | NOT NULL |

Primary key: `(character_id, user_id, item_id)`

### `corporation_assets`

| Column | Type | Constraints |
|--------|------|-------------|
| corporation_id | BIGINT | PK (composite), FK → player_corporations(id, user_id) |
| user_id | BIGINT | PK (composite) |
| item_id | BIGINT | PK (composite) |
| is_blueprint_copy | BOOLEAN | NOT NULL |
| is_singleton | BOOLEAN | NOT NULL |
| location_id | BIGINT | NOT NULL |
| location_type | VARCHAR(15) | NOT NULL |
| quantity | BIGINT | NOT NULL |
| type_id | BIGINT | NOT NULL |
| location_flag | VARCHAR(50) | NOT NULL |
| update_key | TIMESTAMP | NOT NULL |

Primary key: `(corporation_id, user_id, item_id)`

### `corporation_asset_location_names`

| Column | Type | Constraints |
|--------|------|-------------|
| corporation_id | BIGINT | PK (composite), FK → player_corporations(id, user_id) |
| user_id | BIGINT | PK (composite) |
| item_id | BIGINT | PK (composite) |
| name | VARCHAR(500) | NOT NULL |

Primary key: `(corporation_id, user_id, item_id)`

### `asset_item_types`

Item type definitions, enriched by SDE import.

| Column | Type | Constraints |
|--------|------|-------------|
| type_id | BIGINT | PRIMARY KEY |
| type_name | VARCHAR(500) | NOT NULL |
| volume | DOUBLE PRECISION | NOT NULL |
| icon_id | BIGINT | nullable |
| group_id | BIGINT | nullable |
| packaged_volume | DOUBLE PRECISION | nullable |
| mass | DOUBLE PRECISION | nullable |
| capacity | DOUBLE PRECISION | nullable |
| portion_size | INT | nullable |
| published | BOOLEAN | nullable |
| market_group_id | BIGINT | nullable |
| graphic_id | BIGINT | nullable |
| race_id | BIGINT | nullable |
| description | TEXT | nullable |

### `character_blueprints`

Blueprint copies and originals held by characters or corporations. Supports auto-detection of ME/TE for production plans.

| Column | Type | Constraints |
|--------|------|-------------|
| item_id | BIGINT | PRIMARY KEY |
| user_id | BIGINT | NOT NULL |
| owner_id | BIGINT | NOT NULL |
| owner_type | TEXT | NOT NULL |
| type_id | BIGINT | NOT NULL |
| location_id | BIGINT | NOT NULL |
| location_flag | TEXT | NOT NULL DEFAULT '' |
| quantity | INT | NOT NULL DEFAULT 0 |
| material_efficiency | INT | NOT NULL DEFAULT 0 |
| time_efficiency | INT | NOT NULL DEFAULT 0 |
| runs | INT | NOT NULL DEFAULT -1 (-1 = original) |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(user_id, type_id)`, `(owner_id, owner_type)`

---

## 3. Geography

### `regions`

| Column | Type | Constraints |
|--------|------|-------------|
| region_id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | nullable |

### `constellations`

| Column | Type | Constraints |
|--------|------|-------------|
| constellation_id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | NOT NULL |
| region_id | BIGINT | NOT NULL, FK → regions(region_id) |

### `solar_systems`

| Column | Type | Constraints |
|--------|------|-------------|
| solar_system_id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | NOT NULL |
| constellation_id | BIGINT | NOT NULL, FK → constellations(constellation_id) |
| security | DOUBLE PRECISION | NOT NULL |
| x | DOUBLE PRECISION | nullable (3D coordinates for distance calc) |
| y | DOUBLE PRECISION | nullable |
| z | DOUBLE PRECISION | nullable |

### `stations`

| Column | Type | Constraints |
|--------|------|-------------|
| station_id | BIGINT | PRIMARY KEY |
| name | VARCHAR(500) | NOT NULL |
| solar_system_id | BIGINT | NOT NULL (FK dropped in migration 20260220120858) |
| corporation_id | BIGINT | NOT NULL |
| is_npc_station | BOOLEAN | NOT NULL |
| last_updated_at | TIMESTAMP | nullable |

Note: The FK constraint to `solar_systems` was dropped to allow storing stations whose solar system is not yet imported.

---

## 4. SDE (Static Data Export)

SDE tables are populated by the SDE import pipeline. Full column details are in `docs/features/sde-import.md`. The tables below use a compact format.

### Core SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_categories` | category_id | Item categories (ships, modules, etc.) |
| `sde_groups` | group_id | Item groups within categories |
| `sde_meta_groups` | meta_group_id | Meta levels (T1, T2, faction, etc.) |
| `sde_market_groups` | market_group_id | Market browser hierarchy |
| `sde_icons` | icon_id | Icon definitions |
| `sde_graphics` | graphic_id | 3D graphic references |
| `sde_metadata` | key (TEXT) | SDE version/import tracking (key-value store) |

### Blueprint SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_blueprints` | blueprint_type_id | Blueprint headers (max_production_limit) |
| `sde_blueprint_activities` | (blueprint_type_id, activity) | Activity timing per blueprint |
| `sde_blueprint_materials` | (blueprint_type_id, activity, type_id) | Input materials per activity |
| `sde_blueprint_products` | (blueprint_type_id, activity, type_id) | Output products per activity |
| `sde_blueprint_skills` | (blueprint_type_id, activity, type_id) | Required skills per activity |

### Dogma SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_dogma_attribute_categories` | category_id | Dogma attribute groupings |
| `sde_dogma_attributes` | attribute_id | Item dogma attributes |
| `sde_dogma_effects` | effect_id | Item dogma effects |
| `sde_type_dogma_attributes` | (type_id, attribute_id) | Per-type attribute values |
| `sde_type_dogma_effects` | (type_id, effect_id) | Per-type effect assignments |

### NPC SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_factions` | faction_id | EVE factions |
| `sde_npc_corporations` | corporation_id | NPC corporations |
| `sde_npc_corporation_divisions` | (corporation_id, division_id) | NPC corp divisions |
| `sde_agents` | agent_id | NPC agents |
| `sde_agents_in_space` | agent_id | In-space agent locations |
| `sde_races` | race_id | Character races |
| `sde_bloodlines` | bloodline_id | Character bloodlines |
| `sde_ancestries` | ancestry_id | Character ancestries |

### Industry SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_planet_schematics` | schematic_id | PI schematic definitions |
| `sde_planet_schematic_types` | (schematic_id, type_id) | PI schematic inputs/outputs |
| `sde_control_tower_resources` | (control_tower_type_id, resource_type_id) | POS fuel requirements |
| `industry_cost_indices` | (system_id, activity) | Per-system industry cost indices |

### Misc SDE Tables

| Table | Primary Key | Description |
|-------|-------------|-------------|
| `sde_skins` | skin_id | Ship skin definitions |
| `sde_skin_licenses` | license_type_id | Skin license items |
| `sde_skin_materials` | skin_material_id | Skin material references |
| `sde_certificates` | certificate_id | Character certificate definitions |
| `sde_landmarks` | landmark_id | In-space landmarks |
| `sde_station_operations` | operation_id | Station operation types |
| `sde_station_services` | service_id | Station service types |
| `sde_contraband_types` | (faction_id, type_id) | Contraband rules per faction |
| `sde_research_agents` | (agent_id, type_id) | Research agent skill mappings |
| `sde_character_attributes` | attribute_id | Character sheet attributes |
| `sde_corporation_activities` | activity_id | Corporation activity types |
| `sde_tournament_rule_sets` | rule_set_id | Tournament ruleset data (JSONB) |

---

## 5. Market & Pricing

### `market_prices`

Jita market snapshots. FK to `asset_item_types` was dropped (migration `20250101140000`) to allow prices for all items.

| Column | Type | Constraints |
|--------|------|-------------|
| type_id | BIGINT | PRIMARY KEY |
| region_id | BIGINT | NOT NULL |
| buy_price | DOUBLE PRECISION | nullable |
| sell_price | DOUBLE PRECISION | nullable |
| daily_volume | BIGINT | nullable |
| adjusted_price | DOUBLE PRECISION | nullable |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Indexes: `(region_id)`, `(updated_at)`

### `stockpile_markers`

Desired quantity targets per item per location.

| Column | Type | Constraints |
|--------|------|-------------|
| id | SERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| type_id | BIGINT | NOT NULL, FK → asset_item_types(type_id) |
| owner_type | VARCHAR(20) | NOT NULL |
| owner_id | BIGINT | NOT NULL |
| location_id | BIGINT | NOT NULL |
| container_id | BIGINT | nullable |
| division_number | INT | nullable |
| desired_quantity | BIGINT | NOT NULL |
| notes | TEXT | nullable |
| price_source | VARCHAR(20) | nullable |
| price_percentage | NUMERIC(5,2) | nullable |
| plan_id | BIGINT | nullable, FK → production_plans(id) ON DELETE SET NULL |
| auto_production_parallelism | INT | nullable DEFAULT 0 |
| auto_production_enabled | BOOLEAN | NOT NULL DEFAULT false |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Indexes: `idx_stockpile_unique` (unique partial, COALESCE nulls), `(user_id)`, `(type_id)`, `(user_id) WHERE auto_production_enabled`

### `industry_cost_indices`

Note: Although imported via SDE pipeline, this table drives industry cost calculations at runtime.

| Column | Type | Constraints |
|--------|------|-------------|
| system_id | BIGINT | PK (composite) |
| activity | TEXT | PK (composite) |
| cost_index | DOUBLE PRECISION | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

---

## 6. Social & Contacts

### `contacts`

Peer-to-peer contact relationships between users.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| requester_user_id | BIGINT | NOT NULL, FK → users(id) |
| recipient_user_id | BIGINT | NOT NULL, FK → users(id) |
| status | VARCHAR(20) | NOT NULL |
| contact_rule_id | BIGINT | nullable, FK → contact_rules(id) ON DELETE CASCADE |
| requested_at | TIMESTAMP | NOT NULL DEFAULT now() |
| responded_at | TIMESTAMP | nullable |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Constraints: `contacts_unique_pair UNIQUE (requester_user_id, recipient_user_id)`, `contacts_no_self CHECK (requester != recipient)`

Indexes: `(requester_user_id)`, `(recipient_user_id)`, `(status)`, `(contact_rule_id) WHERE NOT NULL`

### `contact_permissions`

Per-service access grants between connected users.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| contact_id | BIGINT | NOT NULL, FK → contacts(id) ON DELETE CASCADE |
| granting_user_id | BIGINT | NOT NULL, FK → users(id) |
| receiving_user_id | BIGINT | NOT NULL, FK → users(id) |
| service_type | VARCHAR(50) | NOT NULL |
| can_access | BOOLEAN | NOT NULL DEFAULT false |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Constraint: `permission_unique_grant UNIQUE (contact_id, granting_user_id, receiving_user_id, service_type)`

Indexes: `(contact_id)`, `(receiving_user_id, service_type, can_access)`

### `contact_rules`

Auto-create contact rules based on corporation or alliance membership.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| rule_type | VARCHAR(20) | NOT NULL — CHECK IN ('corporation', 'alliance', 'everyone') |
| entity_id | BIGINT | nullable |
| entity_name | VARCHAR(500) | nullable |
| permissions | JSONB | NOT NULL DEFAULT '["for_sale_browse"]' |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Indexes: `idx_contact_rules_unique (user_id, rule_type, COALESCE(entity_id,0)) WHERE is_active`, `(rule_type, entity_id) WHERE is_active`

---

## 7. Commerce

### `for_sale_items`

Marketplace listings by users offering items to contacts.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| type_id | BIGINT | NOT NULL, FK → asset_item_types(type_id) |
| owner_type | VARCHAR(20) | NOT NULL |
| owner_id | BIGINT | NOT NULL |
| location_id | BIGINT | NOT NULL |
| container_id | BIGINT | nullable |
| division_number | INT | nullable |
| quantity_available | BIGINT | NOT NULL |
| price_per_unit | NUMERIC(20,2) | NOT NULL |
| auto_sell_container_id | BIGINT | nullable, FK → auto_sell_containers(id) |
| notes | TEXT | nullable |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Constraints: `for_sale_positive_quantity CHECK (quantity_available > 0)`, `for_sale_positive_price CHECK (price_per_unit >= 0)`

Indexes: `idx_for_sale_unique (user_id, type_id, ..., COALESCE nulls) WHERE is_active`, `(user_id)`, `(is_active)`, `(type_id)`, `(auto_sell_container_id) WHERE NOT NULL`

### `purchase_transactions`

Records of completed or pending purchases. `for_sale_item_id` has no FK to preserve history.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| for_sale_item_id | BIGINT | NOT NULL (no FK) |
| buyer_user_id | BIGINT | NOT NULL, FK → users(id) |
| seller_user_id | BIGINT | NOT NULL, FK → users(id) |
| type_id | BIGINT | NOT NULL, FK → asset_item_types(type_id) |
| quantity_purchased | BIGINT | NOT NULL |
| price_per_unit | NUMERIC(20,2) | NOT NULL |
| total_price | NUMERIC(20,2) | NOT NULL |
| status | VARCHAR(20) | NOT NULL DEFAULT 'completed' |
| buy_order_id | BIGINT | nullable, FK → buy_orders(id) |
| is_auto_fulfilled | BOOLEAN | NOT NULL DEFAULT false |
| contract_key | VARCHAR(50) | nullable |
| transaction_notes | TEXT | nullable |
| purchased_at | TIMESTAMP | NOT NULL DEFAULT now() |
| contract_created_at | TIMESTAMP | nullable |
| completed_at | TIMESTAMP | nullable |

Constraints: `purchase_positive_quantity CHECK (quantity_purchased > 0)`, `purchase_different_users CHECK (buyer != seller)`

Indexes: `(buyer_user_id, purchased_at DESC)`, `(seller_user_id, purchased_at DESC)`, `(for_sale_item_id)`, `(contract_key) WHERE NOT NULL`, `idx_auto_fulfill_unique (buy_order_id, for_sale_item_id) WHERE is_auto_fulfilled AND status IN ('pending','contract_created')`

### `buy_orders`

Demand signals from buyers, optionally auto-generated from `auto_buy_configs`.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| buyer_user_id | BIGINT | NOT NULL, FK → users(id) |
| type_id | BIGINT | NOT NULL, FK → asset_item_types(type_id) |
| location_id | BIGINT | NOT NULL |
| quantity_desired | BIGINT | NOT NULL |
| max_price_per_unit | NUMERIC(20,2) | NOT NULL |
| min_price_per_unit | NUMERIC(20,2) | NOT NULL DEFAULT 0 |
| auto_buy_config_id | BIGINT | nullable, FK → auto_buy_configs(id) |
| notes | TEXT | nullable |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Constraints: `buy_order_positive_quantity`, `buy_order_positive_price`

Indexes: `(buyer_user_id)`, `(type_id)`, `(is_active)`, `(buyer_user_id, is_active)`, `idx_buy_orders_auto_buy_unique (buyer_user_id, type_id, location_id, auto_buy_config_id) WHERE auto_buy_config_id NOT NULL AND is_active`

### `auto_sell_containers`

Containers configured for automatic for-sale listing based on price source.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| owner_type | VARCHAR(20) | NOT NULL |
| owner_id | BIGINT | NOT NULL |
| location_id | BIGINT | NOT NULL |
| container_id | BIGINT | nullable (nullable since 20260221011855) |
| division_number | INT | nullable |
| price_percentage | NUMERIC(5,2) | NOT NULL DEFAULT 90.00 |
| price_source | VARCHAR(20) | NOT NULL DEFAULT 'jita_buy' |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Constraint: `auto_sell_valid_percentage CHECK (price_percentage > 0 AND price_percentage <= 200)`

Indexes: `idx_auto_sell_unique_container (..., COALESCE nulls) WHERE is_active`, `(user_id)`

### `auto_buy_configs`

Configurations for auto-generated buy orders scoped to a specific location/container.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| owner_type | VARCHAR(20) | NOT NULL |
| owner_id | BIGINT | NOT NULL |
| location_id | BIGINT | NOT NULL |
| container_id | BIGINT | nullable |
| division_number | INT | nullable |
| max_price_percentage | NUMERIC(5,2) | NOT NULL DEFAULT 100.00 |
| min_price_percentage | NUMERIC(5,2) | NOT NULL DEFAULT 0.00 |
| price_source | VARCHAR(20) | NOT NULL DEFAULT 'jita_sell' |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Indexes: `idx_auto_buy_unique_active (..., COALESCE nulls) WHERE is_active`, `(user_id)`

---

## 8. Industry & Production

### `character_skills`

ESI-synced skill levels per character.

| Column | Type | Constraints |
|--------|------|-------------|
| character_id | BIGINT | PK (composite) |
| user_id | BIGINT | PK (composite) |
| skill_id | BIGINT | PK (composite) |
| trained_level | INT | NOT NULL DEFAULT 0 |
| active_level | INT | NOT NULL DEFAULT 0 |
| skillpoints | BIGINT | NOT NULL DEFAULT 0 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Primary key: `(character_id, skill_id)` | Index: `(user_id)`

### `esi_industry_jobs`

Live industry jobs fetched from ESI.

| Column | Type | Constraints |
|--------|------|-------------|
| job_id | BIGINT | PRIMARY KEY |
| installer_id | BIGINT | NOT NULL |
| user_id | BIGINT | NOT NULL |
| facility_id | BIGINT | NOT NULL |
| station_id | BIGINT | NOT NULL |
| activity_id | INT | NOT NULL |
| blueprint_id | BIGINT | NOT NULL |
| blueprint_type_id | BIGINT | NOT NULL |
| blueprint_location_id | BIGINT | NOT NULL |
| output_location_id | BIGINT | NOT NULL |
| runs | INT | NOT NULL |
| cost | FLOAT8 | nullable |
| licensed_runs | INT | nullable |
| probability | FLOAT8 | nullable |
| product_type_id | BIGINT | nullable |
| status | TEXT | NOT NULL |
| duration | INT | NOT NULL |
| start_date | TIMESTAMPTZ | NOT NULL |
| end_date | TIMESTAMPTZ | NOT NULL |
| pause_date | TIMESTAMPTZ | nullable |
| completed_date | TIMESTAMPTZ | nullable |
| completed_character_id | BIGINT | nullable |
| successful_runs | INT | nullable |
| solar_system_id | BIGINT | nullable |
| source | TEXT | NOT NULL DEFAULT 'character' |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(user_id)`, `(status)`

### `industry_job_queue`

Planned and dispatched industry jobs, linked to production plan runs.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL |
| character_id | BIGINT | nullable |
| blueprint_type_id | BIGINT | NOT NULL |
| activity | TEXT | NOT NULL |
| runs | INT | NOT NULL |
| me_level | INT | NOT NULL DEFAULT 0 |
| te_level | INT | NOT NULL DEFAULT 0 |
| facility_tax | FLOAT8 | NOT NULL DEFAULT 0 |
| status | TEXT | NOT NULL DEFAULT 'planned' |
| esi_job_id | BIGINT | nullable |
| product_type_id | BIGINT | nullable |
| estimated_cost | FLOAT8 | nullable |
| estimated_duration | INT | nullable |
| notes | TEXT | nullable |
| plan_run_id | BIGINT | nullable, FK → production_plan_runs(id) ON DELETE SET NULL |
| plan_step_id | BIGINT | nullable |
| transport_job_id | BIGINT | nullable, FK → transport_jobs(id) |
| sort_order | INT | NOT NULL DEFAULT 0 |
| station_name | TEXT | NOT NULL DEFAULT '' |
| input_location | TEXT | NOT NULL DEFAULT '' |
| output_location | TEXT | NOT NULL DEFAULT '' |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(user_id)`, `(status)`, `(plan_run_id)`

### `production_plans`

Multi-step production plans linking a product to a hierarchy of manufacturing/reaction steps.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| product_type_id | BIGINT | NOT NULL |
| name | TEXT | NOT NULL DEFAULT '' |
| notes | TEXT | nullable |
| default_manufacturing_station_id | BIGINT | nullable, FK → user_stations(id) ON DELETE SET NULL |
| default_reaction_station_id | BIGINT | nullable, FK → user_stations(id) ON DELETE SET NULL |
| transport_fulfillment | TEXT | nullable |
| transport_method | TEXT | nullable |
| transport_profile_id | BIGINT | nullable, FK → transport_profiles(id) ON DELETE SET NULL |
| courier_rate_per_m3 | NUMERIC(12,2) | NOT NULL DEFAULT 0 |
| courier_collateral_rate | NUMERIC(6,4) | NOT NULL DEFAULT 0 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `production_plan_steps`

Individual manufacturing or reaction steps within a plan. Supports recursive parent/child relationships for sub-component production.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| plan_id | BIGINT | NOT NULL, FK → production_plans(id) ON DELETE CASCADE |
| parent_step_id | BIGINT | nullable, FK → production_plan_steps(id) ON DELETE CASCADE |
| product_type_id | BIGINT | NOT NULL |
| blueprint_type_id | BIGINT | NOT NULL |
| activity | TEXT | NOT NULL |
| me_level | INT | NOT NULL DEFAULT 10 |
| te_level | INT | NOT NULL DEFAULT 20 |
| industry_skill | INT | NOT NULL DEFAULT 5 |
| adv_industry_skill | INT | NOT NULL DEFAULT 5 |
| structure | TEXT | NOT NULL DEFAULT 'raitaru' |
| rig | TEXT | NOT NULL DEFAULT 't2' |
| security | TEXT | NOT NULL DEFAULT 'high' |
| facility_tax | NUMERIC(5,2) | NOT NULL DEFAULT 1.0 |
| station_name | TEXT | nullable |
| user_station_id | BIGINT | nullable, FK → user_stations(id) ON DELETE SET NULL |
| source_location_id | BIGINT | nullable |
| source_container_id | BIGINT | nullable |
| source_division_number | INT | nullable |
| source_owner_type | TEXT | nullable |
| source_owner_id | BIGINT | nullable |
| output_owner_type | TEXT | nullable |
| output_owner_id | BIGINT | nullable |
| output_division_number | INT | nullable |
| output_container_id | BIGINT | nullable |

Indexes: `(plan_id)`, `(parent_step_id)`

### `production_plan_runs`

Records each time a plan was triggered to generate jobs, with the quantity multiplier.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| plan_id | BIGINT | NOT NULL, FK → production_plans(id) ON DELETE CASCADE |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| quantity | INT | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(plan_id)`, `(user_id)`

### `user_stations`

User-configured structures (Raitaru, Azbel, etc.) with their tax and type metadata.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| station_id | BIGINT | NOT NULL, FK → stations(station_id) |
| structure | TEXT | NOT NULL DEFAULT 'raitaru' |
| facility_tax | NUMERIC(5,2) | NOT NULL DEFAULT 1.0 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `user_station_rigs`

Rig bonuses installed on a user-configured station.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_station_id | BIGINT | NOT NULL, FK → user_stations(id) ON DELETE CASCADE |
| rig_name | TEXT | NOT NULL |
| category | TEXT | NOT NULL |
| tier | TEXT | NOT NULL |

Index: `(user_station_id)`

### `user_station_services`

Services offered by a user-configured station, scoped by activity type.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_station_id | BIGINT | NOT NULL, FK → user_stations(id) ON DELETE CASCADE |
| service_name | TEXT | NOT NULL |
| activity | TEXT | NOT NULL |

Index: `(user_station_id)`

---

## 9. Planetary Industry

### `pi_launchpad_labels`

User-defined label for a PI launchpad pin.

| Column | Type | Constraints |
|--------|------|-------------|
| user_id | BIGINT | PK (composite), FK → users(id) |
| character_id | BIGINT | PK (composite) |
| planet_id | BIGINT | PK (composite) |
| pin_id | BIGINT | PK (composite) |
| label | VARCHAR(100) | NOT NULL |

Primary key: `(user_id, character_id, planet_id, pin_id)`

### `pi_planets`

ESI-synced planet data per character.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| character_id | BIGINT | NOT NULL |
| user_id | BIGINT | NOT NULL |
| planet_id | BIGINT | NOT NULL |
| planet_type | VARCHAR(20) | NOT NULL |
| solar_system_id | BIGINT | NOT NULL |
| upgrade_level | INT | NOT NULL DEFAULT 0 |
| num_pins | INT | NOT NULL DEFAULT 0 |
| last_update | TIMESTAMP | NOT NULL |
| last_stall_notified_at | TIMESTAMP | nullable |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Unique: `(character_id, planet_id)` | Index: `(user_id)`

### `pi_pins`

Individual pins (structures) on a planet.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| character_id | BIGINT | NOT NULL |
| planet_id | BIGINT | NOT NULL |
| pin_id | BIGINT | NOT NULL |
| type_id | BIGINT | NOT NULL |
| schematic_id | INT | nullable |
| latitude | DOUBLE PRECISION | nullable |
| longitude | DOUBLE PRECISION | nullable |
| install_time | TIMESTAMP | nullable |
| expiry_time | TIMESTAMP | nullable |
| last_cycle_start | TIMESTAMP | nullable |
| extractor_cycle_time | INT | nullable |
| extractor_head_radius | DOUBLE PRECISION | nullable |
| extractor_product_type_id | BIGINT | nullable |
| extractor_qty_per_cycle | INT | nullable |
| extractor_num_heads | INT | nullable |
| pin_category | VARCHAR(20) | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Unique: `(character_id, planet_id, pin_id)` | Index: `(character_id, planet_id)`

### `pi_pin_contents`

Contents of a pi pin at last snapshot.

| Column | Type | Constraints |
|--------|------|-------------|
| character_id | BIGINT | PK (composite) |
| planet_id | BIGINT | PK (composite) |
| pin_id | BIGINT | PK (composite) |
| type_id | BIGINT | PK (composite) |
| amount | BIGINT | NOT NULL |

Primary key: `(character_id, planet_id, pin_id, type_id)`

### `pi_routes`

Material routing between pins on a planet.

| Column | Type | Constraints |
|--------|------|-------------|
| character_id | BIGINT | PK (composite) |
| planet_id | BIGINT | PK (composite) |
| route_id | BIGINT | PK (composite) |
| source_pin_id | BIGINT | NOT NULL |
| destination_pin_id | BIGINT | NOT NULL |
| content_type_id | BIGINT | NOT NULL |
| quantity | BIGINT | NOT NULL |

Primary key: `(character_id, planet_id, route_id)`

### `pi_tax_config`

Per-user per-planet (or default) PI tax rate configuration.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| planet_id | BIGINT | nullable (NULL = default for all planets) |
| tax_rate | DOUBLE PRECISION | NOT NULL DEFAULT 10.0 |

Unique: `(user_id, planet_id)`

---

## 10. Transportation

### `transport_profiles`

Per-user ship/freight transport configurations with cost parameters.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| name | TEXT | NOT NULL |
| transport_method | TEXT | NOT NULL |
| character_id | BIGINT | nullable |
| cargo_m3 | DOUBLE PRECISION | NOT NULL |
| rate_per_m3_per_jump | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| collateral_rate | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| collateral_price_basis | TEXT | NOT NULL DEFAULT 'sell' |
| fuel_type_id | BIGINT | nullable |
| fuel_per_ly | DOUBLE PRECISION | nullable |
| fuel_conservation_level | INT | NOT NULL DEFAULT 0 |
| route_preference | TEXT | NOT NULL DEFAULT 'shortest' |
| is_default | BOOLEAN | NOT NULL DEFAULT false |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `jf_routes`

User-defined jump freighter routes between two systems.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| name | TEXT | NOT NULL |
| origin_system_id | BIGINT | NOT NULL |
| destination_system_id | BIGINT | NOT NULL |
| total_distance_ly | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `jf_route_waypoints`

Ordered cyno stops within a JF route.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| route_id | BIGINT | NOT NULL, FK → jf_routes(id) ON DELETE CASCADE |
| sequence | INT | NOT NULL |
| system_id | BIGINT | NOT NULL |
| distance_ly | DOUBLE PRECISION | NOT NULL DEFAULT 0 |

Index: `(route_id)`

### `transport_jobs`

Individual transport job instances linking origin/destination with items.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| origin_station_id | BIGINT | NOT NULL |
| destination_station_id | BIGINT | NOT NULL |
| origin_system_id | BIGINT | NOT NULL |
| destination_system_id | BIGINT | NOT NULL |
| transport_method | TEXT | NOT NULL |
| route_preference | TEXT | NOT NULL DEFAULT 'shortest' |
| status | TEXT | NOT NULL DEFAULT 'planned' |
| total_volume_m3 | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| total_collateral | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| estimated_cost | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| jumps | INT | NOT NULL DEFAULT 0 |
| distance_ly | DOUBLE PRECISION | nullable |
| jf_route_id | BIGINT | nullable, FK → jf_routes(id) |
| fulfillment_type | TEXT | NOT NULL DEFAULT 'self_haul' |
| transport_profile_id | BIGINT | nullable, FK → transport_profiles(id) |
| plan_run_id | BIGINT | nullable, FK → production_plan_runs(id) |
| plan_step_id | BIGINT | nullable |
| queue_entry_id | BIGINT | nullable |
| notes | TEXT | nullable |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `transport_job_items`

Items included in a transport job.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| transport_job_id | BIGINT | NOT NULL, FK → transport_jobs(id) ON DELETE CASCADE |
| type_id | BIGINT | NOT NULL |
| quantity | INT | NOT NULL |
| volume_m3 | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| estimated_value | DOUBLE PRECISION | NOT NULL DEFAULT 0 |

Index: `(transport_job_id)`

### `transport_trigger_config`

Per-user per-trigger-type transport fulfillment preferences.

| Column | Type | Constraints |
|--------|------|-------------|
| user_id | BIGINT | PK (composite), FK → users(id) |
| trigger_type | TEXT | PK (composite) |
| default_fulfillment | TEXT | NOT NULL DEFAULT 'self_haul' |
| allowed_fulfillments | TEXT[] | NOT NULL DEFAULT '{self_haul}' |
| default_profile_id | BIGINT | nullable, FK → transport_profiles(id) |
| default_method | TEXT | nullable |
| courier_rate_per_m3 | DOUBLE PRECISION | NOT NULL DEFAULT 0 |
| courier_collateral_rate | DOUBLE PRECISION | NOT NULL DEFAULT 0 |

Primary key: `(user_id, trigger_type)`

---

## 11. Notifications

### `discord_links`

OAuth link between a user account and their Discord identity.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) UNIQUE |
| discord_user_id | VARCHAR(50) | NOT NULL |
| discord_username | VARCHAR(100) | NOT NULL |
| access_token | TEXT | NOT NULL |
| refresh_token | TEXT | NOT NULL |
| token_expires_at | TIMESTAMP | NOT NULL |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

### `discord_notification_targets`

Discord channels/DMs configured to receive notifications.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| user_id | BIGINT | NOT NULL, FK → users(id) |
| target_type | VARCHAR(20) | NOT NULL |
| channel_id | VARCHAR(50) | nullable |
| guild_name | VARCHAR(100) | NOT NULL DEFAULT '' |
| channel_name | VARCHAR(100) | NOT NULL DEFAULT '' |
| is_active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMP | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMP | NOT NULL DEFAULT now() |

Index: `(user_id)`

### `notification_preferences`

Per-target event enable/disable switches.

| Column | Type | Constraints |
|--------|------|-------------|
| id | BIGSERIAL | PRIMARY KEY |
| target_id | BIGINT | NOT NULL, FK → discord_notification_targets(id) ON DELETE CASCADE |
| event_type | VARCHAR(50) | NOT NULL |
| is_enabled | BOOLEAN | NOT NULL DEFAULT true |

Unique: `(target_id, event_type)`

---

## 12. Views

### `corporation_asset_locations`

Resolves corporation asset locations to their station, division, and geographic hierarchy. Handles three levels of container nesting plus the ESI `OfficeFolder` pattern.

**Columns exposed:**
- `corporation_id`, `user_id`, `item_id`, `type_id`, `location_id`, `location_type`, `location_flag`
- `container_id`, `container_type_id`, `container_location_flag`, `container_location_type`
- `division_number` (extracted from CorpSAG location flags via SUBSTRING)
- `station_id` (resolved through up to 3 levels of nesting)
- `station_name`, `station_corporation_id`, `is_npc_station`
- `solar_system_id`, `solar_system_name`, `security`
- `constellation_id`, `constellation_name`
- `region_id`, `region_name`

**Source tables:** `corporation_assets` (self-joined 3x for container recursion), `stations`, `solar_systems`, `constellations`, `regions`

**File:** Created in `20250101150000`, revised in `20260221003740` to fix the OfficeFolder ESI data model.

---

## 13. Relationship Map

```
users
├── characters (user_id)
│   ├── character_assets (character_id, user_id)
│   │   └── character_asset_location_names (character_id, user_id, item_id)
│   ├── character_skills (character_id, user_id)
│   └── character_blueprints (user_id, owner_id)
├── player_corporations (user_id)
│   ├── corporation_assets (corporation_id, user_id)
│   │   └── corporation_asset_location_names (corporation_id, user_id, item_id)
│   └── corporation_divisions (corporation_id, user_id)
├── stockpile_markers (user_id)
│   └── production_plans (plan_id)
├── contacts (requester_user_id, recipient_user_id)
│   ├── contact_permissions (contact_id)
│   └── contact_rules (contact_rule_id)
├── for_sale_items (user_id)
│   └── auto_sell_containers (auto_sell_container_id)
├── purchase_transactions (buyer_user_id, seller_user_id)
│   └── buy_orders (buy_order_id)
├── buy_orders (buyer_user_id)
│   └── auto_buy_configs (auto_buy_config_id)
├── production_plans (user_id)
│   ├── production_plan_steps (plan_id) [self-referential via parent_step_id]
│   │   └── user_stations (user_station_id)
│   │       ├── user_station_rigs (user_station_id)
│   │       └── user_station_services (user_station_id)
│   └── production_plan_runs (plan_id)
│       ├── industry_job_queue (plan_run_id)
│       │   └── transport_jobs (transport_job_id)
│       └── transport_jobs (plan_run_id)
│           ├── transport_job_items (transport_job_id)
│           ├── transport_profiles (transport_profile_id)
│           └── jf_routes (jf_route_id)
│               └── jf_route_waypoints (route_id)
├── transport_profiles (user_id)
├── transport_trigger_config (user_id)
├── discord_links (user_id)
└── discord_notification_targets (user_id)
    └── notification_preferences (target_id)

pi_planets (character_id, user_id)
├── pi_pins (character_id, planet_id)
│   └── pi_pin_contents (character_id, planet_id, pin_id)
├── pi_routes (character_id, planet_id)
└── pi_launchpad_labels (user_id, character_id, planet_id, pin_id)

pi_tax_config (user_id)

corporations (id) — referenced by stations(corporation_id), sde_npc_corporations
stations (station_id) — referenced by user_stations, esi_industry_jobs

Geography hierarchy:
regions → constellations → solar_systems → stations

SDE hierarchy:
sde_categories → sde_groups → asset_item_types (group_id)
sde_blueprints → sde_blueprint_activities → sde_blueprint_materials
                                           → sde_blueprint_products
                                           → sde_blueprint_skills
```

---

## 14. Index Reference

### Custom Unique Indexes

| Index | Table | Columns | Where Clause |
|-------|-------|---------|--------------|
| `idx_stockpile_unique` | stockpile_markers | user_id, type_id, owner_type, owner_id, location_id, COALESCE(container_id,0), COALESCE(division_number,0) | — |
| `idx_for_sale_unique` | for_sale_items | user_id, type_id, owner_type, owner_id, location_id, COALESCE(container_id,0), COALESCE(division_number,0) | WHERE is_active |
| `idx_auto_sell_unique_container` | auto_sell_containers | user_id, owner_type, owner_id, location_id, COALESCE(container_id,0), COALESCE(division_number,0) | WHERE is_active |
| `idx_auto_buy_unique_active` | auto_buy_configs | user_id, owner_type, owner_id, location_id, COALESCE(container_id,0), COALESCE(division_number,0) | WHERE is_active |
| `idx_buy_orders_auto_buy_unique` | buy_orders | buyer_user_id, type_id, location_id, auto_buy_config_id | WHERE auto_buy_config_id NOT NULL AND is_active |
| `idx_auto_fulfill_unique` | purchase_transactions | buy_order_id, for_sale_item_id | WHERE is_auto_fulfilled AND status IN ('pending','contract_created') |
| `idx_contact_rules_unique` | contact_rules | user_id, rule_type, COALESCE(entity_id,0) | WHERE is_active |
| `contacts_unique_pair` | contacts | requester_user_id, recipient_user_id | — |
| `permission_unique_grant` | contact_permissions | contact_id, granting_user_id, receiving_user_id, service_type | — |

### Custom Non-Unique Indexes

| Index | Table | Columns | Where Clause |
|-------|-------|---------|--------------|
| `idx_stockpile_user` | stockpile_markers | user_id | — |
| `idx_stockpile_type` | stockpile_markers | type_id | — |
| `idx_stockpile_markers_auto_production` | stockpile_markers | user_id | WHERE auto_production_enabled |
| `idx_market_prices_region` | market_prices | region_id | — |
| `idx_market_prices_updated` | market_prices | updated_at | — |
| `idx_contacts_requester` | contacts | requester_user_id | — |
| `idx_contacts_recipient` | contacts | recipient_user_id | — |
| `idx_contacts_status` | contacts | status | — |
| `idx_contacts_rule` | contacts | contact_rule_id | WHERE NOT NULL |
| `idx_permission_contact` | contact_permissions | contact_id | — |
| `idx_permission_receiving` | contact_permissions | receiving_user_id, service_type, can_access | — |
| `idx_for_sale_user` | for_sale_items | user_id | — |
| `idx_for_sale_active` | for_sale_items | is_active | — |
| `idx_for_sale_type` | for_sale_items | type_id | — |
| `idx_for_sale_auto_sell` | for_sale_items | auto_sell_container_id | WHERE NOT NULL |
| `idx_purchase_buyer` | purchase_transactions | buyer_user_id, purchased_at DESC | — |
| `idx_purchase_seller` | purchase_transactions | seller_user_id, purchased_at DESC | — |
| `idx_purchase_item` | purchase_transactions | for_sale_item_id | — |
| `idx_purchase_transactions_contract_key` | purchase_transactions | contract_key | WHERE NOT NULL |
| `idx_buy_orders_buyer` | buy_orders | buyer_user_id | — |
| `idx_buy_orders_type` | buy_orders | type_id | — |
| `idx_buy_orders_active` | buy_orders | is_active | — |
| `idx_buy_orders_buyer_active` | buy_orders | buyer_user_id, is_active | — |
| `idx_auto_sell_user` | auto_sell_containers | user_id | — |
| `idx_auto_buy_user` | auto_buy_configs | user_id | — |
| `idx_contact_rules_entity` | contact_rules | rule_type, entity_id | WHERE is_active |
| `idx_discord_targets_user` | discord_notification_targets | user_id | — |
| `idx_character_skills_user` | character_skills | user_id | — |
| `idx_esi_jobs_user` | esi_industry_jobs | user_id | — |
| `idx_esi_jobs_status` | esi_industry_jobs | status | — |
| `idx_job_queue_user` | industry_job_queue | user_id | — |
| `idx_job_queue_status` | industry_job_queue | status | — |
| `idx_job_queue_plan_run_id` | industry_job_queue | plan_run_id | — |
| `idx_production_plans_user_id` | production_plans | user_id | — |
| `idx_production_plan_steps_plan_id` | production_plan_steps | plan_id | — |
| `idx_production_plan_steps_parent` | production_plan_steps | parent_step_id | — |
| `idx_plan_runs_plan_id` | production_plan_runs | plan_id | — |
| `idx_plan_runs_user_id` | production_plan_runs | user_id | — |
| `idx_user_stations_user_id` | user_stations | user_id | — |
| `idx_user_station_rigs_station_id` | user_station_rigs | user_station_id | — |
| `idx_user_station_services_station_id` | user_station_services | user_station_id | — |
| `idx_pi_planets_user` | pi_planets | user_id | — |
| `idx_pi_pins_planet` | pi_pins | character_id, planet_id | — |
| `idx_transport_profiles_user_id` | transport_profiles | user_id | — |
| `idx_jf_routes_user_id` | jf_routes | user_id | — |
| `idx_jf_route_waypoints_route_id` | jf_route_waypoints | route_id | — |
| `idx_transport_jobs_user_id` | transport_jobs | user_id | — |
| `idx_transport_job_items_job_id` | transport_job_items | transport_job_id | — |
| `idx_character_blueprints_user_type` | character_blueprints | user_id, type_id | — |
| `idx_character_blueprints_owner` | character_blueprints | owner_id, owner_type | — |

---

## 15. Schema Conventions

### Data Types

| Use Case | Type |
|----------|------|
| Entity IDs from EVE ESI | BIGINT |
| Auto-increment PKs | BIGSERIAL |
| Prices (ISK) | NUMERIC(20,2) — was BIGINT before migration 20260217183856 |
| Percentages | NUMERIC(5,2) |
| Timestamps (app-managed) | TIMESTAMP with DEFAULT NOW() |
| Timestamps (ESI-sourced) | TIMESTAMPTZ |
| Short enumerations | VARCHAR(N) with optional CHECK constraint |
| Flags and multi-value strings | TEXT or TEXT[] |
| Config blobs | JSONB |
| Floating point game values | DOUBLE PRECISION or FLOAT8 |

### Naming Conventions

- Tables: `snake_case`, plural nouns
- Columns: `snake_case`
- Primary keys: `id` (BIGSERIAL) for application entities; composite PKs for join/asset tables
- Foreign keys: `{referenced_table_singular}_id` (e.g., `user_id`, `plan_id`)
- Timestamps: `created_at`, `updated_at`, `*_at` suffix
- Boolean columns: `is_{adjective}` (e.g., `is_active`, `is_npc_station`)
- Polymorphic owner columns: `owner_type` + `owner_id` pair (values: `'character'`, `'corporation'`)

### Recurring Patterns

**Soft deletes:** `is_active BOOLEAN NOT NULL DEFAULT true` — rows are deactivated, not deleted. Unique indexes use `WHERE is_active = true` to allow reuse of the same logical slot.

**COALESCE in unique indexes:** Nullable columns (container_id, division_number) use `COALESCE(col, 0)` in unique index expressions since NULL != NULL in SQL.

**Polymorphic owner:** `(owner_type, owner_id)` pair used in stockpile_markers, for_sale_items, auto_sell_containers, auto_buy_configs, character_blueprints. `owner_type` values: `'character'` or `'corporation'`.

**Composite PKs for asset tables:** Character and corporation asset tables use `(entity_id, user_id, item_id)` composite PKs scoped to the owning user.

**Cascade deletes:** Plan steps, plan runs, rigs, services, waypoints, and notification preferences all cascade on parent deletion. Contacts cascade to permissions and rules.

**History preservation:** `purchase_transactions.for_sale_item_id` intentionally has no FK to preserve records after listing deletion.

---

## 16. Migration History

| Migration | Timestamp | Description |
|-----------|-----------|-------------|
| `initial_database` | 20250101100000 | Creates users, characters, player_corporations, corporations, character/corporation assets + location names, corporation_hanger_divisions, asset_item_types, regions, constellations, solar_systems, stations |
| `corporation_divisions` | 20250101110000 | Drops corporation_hanger_divisions; creates corporation_divisions with division_type |
| `stockpile_markers` | 20250101120000 | Creates stockpile_markers with COALESCE unique index |
| `market_prices` | 20250101130000 | Creates market_prices table |
| `remove_market_prices_fk` | 20250101140000 | Drops market_prices FK to asset_item_types |
| `corporation_asset_locations_view` | 20250101150000 | Creates corporation_asset_locations view with container recursion |
| `contacts` | 20250101160000 | Creates contacts table with unique pair and self-check constraints |
| `contact_permissions` | 20250101170000 | Creates contact_permissions table |
| `for_sale_items` | 20250101180000 | Creates for_sale_items with partial unique index |
| `purchase_transactions` | 20250101190000 | Creates purchase_transactions |
| `add_contract_key` | 20250101200000 | Adds contract_key to purchase_transactions |
| `buy_orders` | 20250101210000 | Creates buy_orders table |
| `enrich_existing_tables` | 20250216100000 | Adds group_id, packaged_volume, mass, capacity, etc. to asset_item_types; adds adjusted_price to market_prices |
| `create_sde_core_tables` | 20250216110000 | Creates sde_categories, sde_groups, sde_meta_groups, sde_market_groups, sde_icons, sde_graphics, sde_metadata |
| `create_sde_blueprint_tables` | 20250216120000 | Creates sde_blueprints, sde_blueprint_activities/materials/products/skills |
| `create_sde_dogma_tables` | 20250216130000 | Creates sde_dogma_attribute_categories/attributes/effects, sde_type_dogma_attributes/effects |
| `create_sde_npc_tables` | 20250216140000 | Creates sde_factions, sde_npc_corporations/divisions, sde_agents/agents_in_space, sde_races/bloodlines/ancestries |
| `create_sde_industry_tables` | 20250216150000 | Creates sde_planet_schematics/types, sde_control_tower_resources, industry_cost_indices |
| `create_sde_misc_tables` | 20250216160000 | Creates sde_skins, sde_skin_licenses/materials, sde_certificates, sde_landmarks, sde_station_operations/services, sde_contraband_types, sde_research_agents, sde_character_attributes, sde_corporation_activities, sde_tournament_rule_sets |
| `add_esi_scopes` | 20250217100000 | Adds esi_scopes to characters and player_corporations |
| `decimal_prices` | 20260217183856 | Converts price columns to NUMERIC(20,2) in for_sale_items, purchase_transactions, buy_orders |
| `add_assets_last_updated` | 20260217205017 | Adds assets_last_updated_at to users |
| `auto_sell_containers` | 20260217215240 | Creates auto_sell_containers table |
| `add_auto_sell_to_for_sale` | 20260217215243 | Adds auto_sell_container_id to for_sale_items |
| `contact_rules` | 20260219132354 | Creates contact_rules with JSONB permissions |
| `add_contact_rule_id_to_contacts` | 20260219132355 | Adds contact_rule_id to contacts |
| `add_alliance_to_player_corporations` | 20260219132356 | Adds alliance_id and alliance_name to player_corporations |
| `add_permissions_to_contact_rules` | 20260219144445 | Adds permissions JSONB column to contact_rules |
| `add_price_source_to_auto_sell` | 20260219155015 | Adds price_source to auto_sell_containers |
| `add_location_to_buy_orders` | 20260219163012 | Adds location_id (NOT NULL) to buy_orders; clears existing rows |
| `add_discord_notifications` | 20260219172431 | Creates discord_links, discord_notification_targets, notification_preferences |
| `auto_buy_configs` | 20260219200824 | Creates auto_buy_configs table |
| `add_auto_buy_to_buy_orders` | 20260219200829 | Adds auto_buy_config_id to buy_orders with unique partial index |
| `add_pricing_to_stockpile_markers` | 20260219200835 | Adds price_source and price_percentage to stockpile_markers |
| `add_corporation_id_to_characters` | 20260220001437 | Adds corporation_id to characters |
| `auto_buy_price_range` | 20260220005130 | Renames price_percentage to max_price_percentage; adds min_price_percentage to auto_buy_configs; adds min_price_per_unit to buy_orders |
| `auto_fulfill_purchases` | 20260220005134 | Adds buy_order_id, is_auto_fulfilled to purchase_transactions; creates auto-fulfill unique partial index |
| `add_station_last_updated` | 20260220105221 | Adds last_updated_at to stations |
| `drop_stations_solar_system_fk` | 20260220120858 | Drops FK constraint from stations to solar_systems |
| `create_pi_launchpad_labels` | 20260220131419 | Creates pi_launchpad_labels |
| `create_pi_tables` | 20260221000943 | Creates pi_planets, pi_pins, pi_pin_contents, pi_routes, pi_tax_config |
| `fix_corp_asset_locations_view` | 20260221003740 | Replaces corporation_asset_locations view to correctly handle OfficeFolder ESI pattern |
| `auto_sell_nullable_container` | 20260221011855 | Makes auto_sell_containers.container_id nullable |
| `widen_auto_fulfill_unique_index` | 20260221113833 | Widens auto-fulfill unique index to cover both 'pending' and 'contract_created' statuses; cancels pre-existing duplicates |
| `cancel_stale_auto_fulfill_purchases` | 20260221124555 | Data migration: cancels pending auto-fulfill purchases for inactive listings |
| `cancel_duplicate_auto_fulfill_from_id_cycling` | 20260221165358 | Data migration: cancels duplicate auto-fulfill purchases caused by buy_order ID cycling |
| `add_contract_created_at` | 20260221222727 | Adds contract_created_at to purchase_transactions |
| `create_character_skills` | 20260222014855 | Creates character_skills table |
| `create_esi_industry_jobs` | 20260222014858 | Creates esi_industry_jobs table |
| `create_industry_job_queue` | 20260222014859 | Creates industry_job_queue table |
| `add_source_to_esi_industry_jobs` | 20260222125510 | Adds source column to esi_industry_jobs |
| `add_completed_at` | 20260222143509 | Adds completed_at to purchase_transactions |
| `create_production_plans` | 20260222151815 | Creates production_plans and production_plan_steps tables |
| `rename_system_id_to_station_id` | 20260222170233 | Drops system_id from production_plan_steps; adds station_name |
| `create_user_stations` | 20260222175330 | Creates user_stations, user_station_rigs, user_station_services; adds user_station_id to production_plan_steps |
| `add_plan_default_stations` | 20260222185246 | Adds default_manufacturing_station_id, default_reaction_station_id to production_plans |
| `add_step_output_location` | 20260222220229 | Adds output_owner_type/id/division_number/container_id to production_plan_steps |
| `plan_runs` | 20260222231117 | Creates production_plan_runs; adds plan_run_id, plan_step_id to industry_job_queue |
| `add_solar_system_coordinates` | 20260224104031 | Adds x, y, z coordinates to solar_systems |
| `create_transport_tables` | 20260224104204 | Creates transport_profiles, jf_routes, jf_route_waypoints, transport_jobs, transport_job_items, transport_trigger_config; adds transport_job_id to industry_job_queue |
| `add_plan_transport_settings` | 20260224205134 | Adds transport_fulfillment, transport_method, transport_profile_id, courier_rate_per_m3, courier_collateral_rate to production_plans |
| `add_sort_order_to_job_queue` | 20260224222923 | Adds sort_order, station_name, input_location, output_location to industry_job_queue |
| `create_character_blueprints` | 20260225111931 | Creates character_blueprints table for ME/TE auto-detection |
| `auto_production` | 20260225205355 | Adds plan_id, auto_production_parallelism, auto_production_enabled to stockpile_markers |
