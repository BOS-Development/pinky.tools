BEGIN;

CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(500) NOT NULL
);

CREATE TABLE characters (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL references users(id),
    name VARCHAR(500) NOT NULL,
    esi_token VARCHAR(5000) NOT NULL,
    esi_refresh_token VARCHAR(5000) NOT NULL,
    esi_token_expires_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id, user_id)
);

CREATE TABLE player_corporations (
    id BIGINT NOT NULL,
    user_id BIGINT NOT NULL references users(id),
    name VARCHAR(500) NOT NULL,
    esi_token VARCHAR(5000) NOT NULL,
    esi_refresh_token VARCHAR(5000) NOT NULL,
    esi_token_expires_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id, user_id)
);

CREATE TABLE character_assets (
    character_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_id BIGINT NOT NULL,
    update_key VARCHAR(100) NOT NULL,
    is_blueprint_copy BOOLEAN NOT NULL,
    is_singleton BOOLEAN NOT NULL,
    location_id BIGINT NOT NULL,
    location_type VARCHAR(15) NOT NULL,
    quantity BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    location_flag VARCHAR(50) NOT NULL,
    PRIMARY KEY (character_id, user_id, item_id),
    FOREIGN KEY (character_id, user_id) REFERENCES characters(id, user_id)
);

CREATE TABLE character_asset_location_names (
    character_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_id BIGINT NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (character_id, user_id, item_id),
    FOREIGN KEY (character_id, user_id) REFERENCES characters(id, user_id)
);

CREATE TABLE corporations (
    id BIGINT PRIMARY KEY NOT NULL,
    name VARCHAR(500) NOT NULL
);

CREATE TABLE corporation_assets (
    corporation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_id BIGINT NOT NULL,
    is_blueprint_copy BOOLEAN NOT NULL,
    is_singleton BOOLEAN NOT NULL,
    location_id BIGINT NOT NULL,
    location_type VARCHAR(15) NOT NULL,
    quantity BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    location_flag VARCHAR(50) NOT NULL,
    update_key TIMESTAMP NOT NULL,
    PRIMARY KEY (corporation_id, user_id, item_id),
    FOREIGN KEY (corporation_id, user_id) REFERENCES player_corporations(id, user_id)
);

CREATE TABLE corporation_asset_location_names (
    corporation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_id BIGINT NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (corporation_id, user_id, item_id),
    FOREIGN KEY (corporation_id, user_id) REFERENCES player_corporations(id, user_id)
);

CREATE TABLE corporation_hanger_divisions (
    corporation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    division INT NOT NULL,
    name VARCHAR(500),
    PRIMARY KEY (corporation_id, user_id, division),
    FOREIGN KEY (corporation_id, user_id) REFERENCES player_corporations(id, user_id)
);

CREATE TABLE asset_item_types (
    type_id BIGINT PRIMARY KEY  NOT NULL,
    type_name VARCHAR(500) NOT NULL,
    volume DOUBLE PRECISION NOT NULL,
    icon_id BIGINT
);

CREATE TABLE regions (
    region_id BIGINT PRIMARY KEY NOT NULL,
    name varchar(500)
);

CREATE TABLE constellations (
    constellation_id BIGINT PRIMARY KEY NOT NULL,
    name varchar(500) NOT NULL,
    region_id BIGINT NOT NULL REFERENCES regions(region_id)
);

CREATE TABLE solar_systems (
    solar_system_id BIGINT PRIMARY KEY NOT NULL,
    name varchar(500) NOT NULL,
    constellation_id BIGINT NOT NULL references constellations(constellation_id),
    security DOUBLE PRECISION NOT NULL
);

CREATE TABLE stations (
    station_id BIGINT PRIMARY KEY NOT NULL,
    name varchar(500) NOT NULL,
    solar_system_id BIGINT NOT NULL REFERENCES solar_systems(solar_system_id),
    corporation_id BIGINT NOT NULL,
    is_npc_station BOOLEAN NOT NULL
);

COMMIT;
