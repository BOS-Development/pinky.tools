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

COMMIT;
