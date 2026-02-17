-- E2E Test Seed Data
-- Only bootstraps data that cannot be created through the app without real EVE OAuth.
-- Game data (assets, market prices, divisions) comes from the mock ESI during tests.

BEGIN;

-- ===========================================
-- Static Universe Data (reference data)
-- ===========================================

INSERT INTO regions (region_id, name) VALUES
  (10000002, 'The Forge'),
  (10000043, 'Domain');

INSERT INTO constellations (constellation_id, name, region_id) VALUES
  (20000020, 'Kimotoro', 10000002),
  (20000322, 'Throne Worlds', 10000043);

INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES
  (30000142, 'Jita', 20000020, 0.9),
  (30002187, 'Amarr', 20000322, 1.0);

INSERT INTO stations (station_id, name, solar_system_id, corporation_id, is_npc_station) VALUES
  (60003760, 'Jita IV - Moon 4 - Caldari Navy Assembly Plant', 30000142, 1000035, true),
  (60008494, 'Amarr VIII (Oris) - Emperor Family Academy', 30002187, 1000066, true);

INSERT INTO asset_item_types (type_id, type_name, volume, icon_id) VALUES
  (34, 'Tritanium', 0.01, 34),
  (35, 'Pyerite', 0.01, 35),
  (36, 'Mexallon', 0.01, 36),
  (37, 'Isogen', 0.01, 37),
  (38, 'Nocxium', 0.01, 38),
  (587, 'Rifter', 16500, 587),
  (11399, 'Raven Navy Issue', 486000, 11399),
  (17703, 'Medium Standard Container', 65, 17703),
  (27, 'Office', 0, NULL);

-- ===========================================
-- Users (created via NextAuth CredentialsProvider login, but characters need to pre-exist)
-- ===========================================

INSERT INTO users (id, name) VALUES
  (1001, 'Alice Stargazer'),
  (1002, 'Bob Miner'),
  (1003, 'Charlie Trader'),
  (1004, 'Diana Scout');

-- ===========================================
-- Characters with fake ESI tokens
-- (Adding characters requires EVE OAuth which we mock out entirely)
-- ===========================================

-- Alice Stargazer's characters
INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on) VALUES
  (2001001, 1001, 'Alice Alpha', 'fake-token-alice-alpha', 'fake-refresh-alice-alpha', NOW() + INTERVAL '1 hour'),
  (2001002, 1001, 'Alice Beta', 'fake-token-alice-beta', 'fake-refresh-alice-beta', NOW() + INTERVAL '1 hour');

-- Bob Miner's character
INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on) VALUES
  (2002001, 1002, 'Bob Bravo', 'fake-token-bob-bravo', 'fake-refresh-bob-bravo', NOW() + INTERVAL '1 hour');

-- Charlie Trader's character
INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on) VALUES
  (2003001, 1003, 'Charlie Charlie', 'fake-token-charlie', 'fake-refresh-charlie', NOW() + INTERVAL '1 hour');

-- Diana Scout's character
INSERT INTO characters (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on) VALUES
  (2004001, 1004, 'Diana Delta', 'fake-token-diana', 'fake-refresh-diana', NOW() + INTERVAL '1 hour');

-- ===========================================
-- Corporations with fake ESI tokens
-- ===========================================

INSERT INTO corporations (id, name) VALUES
  (3001001, 'Stargazer Industries');

INSERT INTO player_corporations (id, user_id, name, esi_token, esi_refresh_token, esi_token_expires_on) VALUES
  (3001001, 1001, 'Stargazer Industries', 'fake-token-stargazer-corp', 'fake-refresh-stargazer-corp', NOW() + INTERVAL '1 hour');

COMMIT;
