-- E2E Test Seed Data
-- Only bootstraps data that cannot be created through the app.
-- Characters and corporations are now created via E2E API routes (/api/e2e/add-character, /api/e2e/add-corporation).
-- Game data (assets, market prices, divisions) comes from the mock ESI during tests.

BEGIN;

-- ===========================================
-- Static Universe Data (reference data)
-- ===========================================

INSERT INTO regions (region_id, name) VALUES
  (10000002, 'The Forge'),
  (10000043, 'Domain')
ON CONFLICT (region_id) DO NOTHING;

INSERT INTO constellations (constellation_id, name, region_id) VALUES
  (20000020, 'Kimotoro', 10000002),
  (20000322, 'Throne Worlds', 10000043)
ON CONFLICT (constellation_id) DO NOTHING;

INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES
  (30000142, 'Jita', 20000020, 0.9),
  (30002187, 'Amarr', 20000322, 1.0)
ON CONFLICT (solar_system_id) DO NOTHING;

INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES
  (60003760, 'Jita IV - Moon 4 - Caldari Navy Assembly Plant', 30000142, 1000035, true),
  (60008494, 'Amarr VIII (Oris) - Emperor Family Academy', 30002187, 1000066, true)
ON CONFLICT (station_id) DO NOTHING;

INSERT INTO asset_item_types (type_id, type_name, volume, icon_id) VALUES
  (34, 'Tritanium', 0.01, 34),
  (35, 'Pyerite', 0.01, 35),
  (36, 'Mexallon', 0.01, 36),
  (37, 'Isogen', 0.01, 37),
  (38, 'Nocxium', 0.01, 38),
  (587, 'Rifter', 16500, 587),
  (11399, 'Raven Navy Issue', 486000, 11399),
  (9999001, 'Medium Standard Container', 65, NULL),
  (27, 'Office', 0, NULL)
ON CONFLICT (type_id) DO NOTHING;

-- ===========================================
-- Users (created via NextAuth CredentialsProvider login)
-- ===========================================

INSERT INTO users (id, name) VALUES
  (1001, 'Alice Stargazer'),
  (1002, 'Bob Miner'),
  (1003, 'Charlie Trader'),
  (1004, 'Diana Scout');

-- ===========================================
-- SDE Data for Industry Tests
-- ===========================================

-- Categories
INSERT INTO sde_categories (category_id, name, published) VALUES
  (4,  'Material',  true),
  (6,  'Ship',      true),
  (9,  'Blueprint', true),
  (16, 'Skill',     true)
ON CONFLICT (category_id) DO NOTHING;

-- Groups
INSERT INTO sde_groups (group_id, name, category_id, published) VALUES
  (18,  'Mineral',           4,  true),
  (25,  'Frigate',           6,  true),
  (105, 'Frigate Blueprint', 9,  true),
  (270, 'Science',           16, true)
ON CONFLICT (group_id) DO NOTHING;

-- Enrich existing item types with group_id so SDE queries can join sde_groups
UPDATE asset_item_types SET group_id = 18 WHERE type_id IN (34, 35, 36, 37, 38);
UPDATE asset_item_types SET group_id = 25 WHERE type_id = 587;

-- Blueprint item type (needed by SearchBlueprints which joins asset_item_types on blueprint_type_id)
INSERT INTO asset_item_types (type_id, type_name, volume, group_id) VALUES
  (787, 'Rifter Blueprint', 0.01, 105)
ON CONFLICT (type_id) DO NOTHING;

-- Skill item types (referenced by sde_blueprint_skills)
INSERT INTO asset_item_types (type_id, type_name, volume, group_id) VALUES
  (3380,  'Industry',                      0.01, 270),
  (3388,  'Advanced Industry',             0.01, 270),
  (3387,  'Mass Production',               0.01, 270),
  (24625, 'Advanced Mass Production',      0.01, 270),
  (45746, 'Reactions',                     0.01, 270),
  (45748, 'Mass Reactions',                0.01, 270),
  (45749, 'Advanced Mass Reactions',       0.01, 270),
  (3402,  'Science',                       0.01, 270),
  (3406,  'Laboratory Operation',          0.01, 270),
  (24624, 'Advanced Laboratory Operation', 0.01, 270)
ON CONFLICT (type_id) DO NOTHING;

-- Blueprint definition
INSERT INTO sde_blueprints (blueprint_type_id, max_production_limit) VALUES
  (787, 300)
ON CONFLICT (blueprint_type_id) DO NOTHING;

-- Manufacturing activity (1 hour)
INSERT INTO sde_blueprint_activities (blueprint_type_id, activity, time) VALUES
  (787, 'manufacturing', 3600)
ON CONFLICT (blueprint_type_id, activity) DO NOTHING;

-- Product: 1x Rifter per run
INSERT INTO sde_blueprint_products (blueprint_type_id, activity, type_id, quantity) VALUES
  (787, 'manufacturing', 587, 1)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Materials: Tritanium 2500, Pyerite 1000, Mexallon 250, Isogen 50
INSERT INTO sde_blueprint_materials (blueprint_type_id, activity, type_id, quantity) VALUES
  (787, 'manufacturing', 34, 2500),
  (787, 'manufacturing', 35, 1000),
  (787, 'manufacturing', 36, 250),
  (787, 'manufacturing', 37, 50)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Skill requirement: Industry Level 1
INSERT INTO sde_blueprint_skills (blueprint_type_id, activity, type_id, level) VALUES
  (787, 'manufacturing', 3380, 1)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Industry cost indices (Jita system)
INSERT INTO industry_cost_indices (system_id, activity, cost_index) VALUES
  (30000142, 'manufacturing', 0.0638),
  (30000142, 'reaction',      0.0200)
ON CONFLICT (system_id, activity) DO NOTHING;

-- Market prices for manufacturing calculator (EIV calculation)
-- market_prices columns: type_id, region_id, buy_price, sell_price, adjusted_price
INSERT INTO market_prices (type_id, region_id, buy_price, sell_price, adjusted_price) VALUES
  (34,  10000002, 5.50,       6.00,        5.75),
  (35,  10000002, 10.00,      11.50,       10.75),
  (36,  10000002, 70.00,      75.00,       72.50),
  (37,  10000002, 50.00,      55.00,       52.50),
  (38,  10000002, 800.00,     900.00,      850.00),
  (587, 10000002, 500000.00,  600000.00,   550000.00)
ON CONFLICT (type_id) DO UPDATE SET
  buy_price      = EXCLUDED.buy_price,
  sell_price     = EXCLUDED.sell_price,
  adjusted_price = EXCLUDED.adjusted_price;

-- ===========================================
-- SDE Data for Reactions Calculator Tests
-- ===========================================

-- Advanced Material group (category 4 = Material)
INSERT INTO sde_groups (group_id, name, category_id, published) VALUES
  (428, 'Advanced Material', 4, true)
ON CONFLICT (group_id) DO NOTHING;

-- Crystalline Carbonide (reaction output product)
-- This must have a group_id pointing to a non-filtered group so it appears in ReactionPicker
-- (filtered groups: 'Intermediate Materials', 'Unrefined Mineral')
INSERT INTO asset_item_types (type_id, type_name, volume, group_id) VALUES
  (16634, 'Crystalline Carbonide', 0.01, 428)
ON CONFLICT (type_id) DO NOTHING;

-- Crystalline Carbonide Reaction Formula (blueprint-equivalent for reactions)
-- blueprint_type_id: 28209
INSERT INTO asset_item_types (type_id, type_name, volume, group_id) VALUES
  (28209, 'Crystalline Carbonide Reaction Formula', 0.01, 105)
ON CONFLICT (type_id) DO NOTHING;

INSERT INTO sde_blueprints (blueprint_type_id, max_production_limit) VALUES
  (28209, 0)
ON CONFLICT (blueprint_type_id) DO NOTHING;

-- Reaction activity: 1 hour (3600s) — matches GetAllReactions WHERE time >= 3600
INSERT INTO sde_blueprint_activities (blueprint_type_id, activity, time) VALUES
  (28209, 'reaction', 3600)
ON CONFLICT (blueprint_type_id, activity) DO NOTHING;

-- Output: 100x Crystalline Carbonide per run
INSERT INTO sde_blueprint_products (blueprint_type_id, activity, type_id, quantity) VALUES
  (28209, 'reaction', 16634, 100)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Inputs: Nocxium 40 + Isogen 80 (both already in asset_item_types)
INSERT INTO sde_blueprint_materials (blueprint_type_id, activity, type_id, quantity) VALUES
  (28209, 'reaction', 38, 40),
  (28209, 'reaction', 37, 80)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Skill requirement: Reactions Level 1
INSERT INTO sde_blueprint_skills (blueprint_type_id, activity, type_id, level) VALUES
  (28209, 'reaction', 45746, 1)
ON CONFLICT (blueprint_type_id, activity, type_id) DO NOTHING;

-- Market price for Crystalline Carbonide output
INSERT INTO market_prices (type_id, region_id, buy_price, sell_price, adjusted_price) VALUES
  (16634, 10000002, 3500.00, 4000.00, 3750.00)
ON CONFLICT (type_id) DO UPDATE SET
  buy_price      = EXCLUDED.buy_price,
  sell_price     = EXCLUDED.sell_price,
  adjusted_price = EXCLUDED.adjusted_price;

-- ===========================================
-- Character Skills for Job Slot & Industry Tests
-- ===========================================

INSERT INTO character_skills (character_id, user_id, skill_id, trained_level, active_level, skillpoints, updated_at) VALUES
  -- Alice Alpha (2001001, user 1001): Full industry skill set
  (2001001, 1001, 3380,  5, 5, 256000, NOW()),   -- Industry
  (2001001, 1001, 3388,  5, 5, 256000, NOW()),   -- Advanced Industry
  (2001001, 1001, 3387,  5, 5, 256000, NOW()),   -- Mass Production
  (2001001, 1001, 24625, 4, 4, 135765, NOW()),   -- Advanced Mass Production
  (2001001, 1001, 45746, 4, 4, 135765, NOW()),   -- Reactions
  (2001001, 1001, 45748, 3, 3, 40000,  NOW()),   -- Mass Reactions
  (2001001, 1001, 45749, 2, 2, 11314,  NOW()),   -- Advanced Mass Reactions
  (2001001, 1001, 3402,  4, 4, 135765, NOW()),   -- Science
  (2001001, 1001, 3406,  3, 3, 40000,  NOW()),   -- Laboratory Operation
  (2001001, 1001, 24624, 2, 2, 11314,  NOW()),   -- Advanced Laboratory Operation
  -- Bob Bravo (2002001, user 1002): Basic manufacturing
  (2002001, 1002, 3380,  4, 4, 135765, NOW()),   -- Industry
  (2002001, 1002, 3387,  3, 3, 40000,  NOW())    -- Mass Production
ON CONFLICT (character_id, skill_id) DO NOTHING;

COMMIT;
