BEGIN;

CREATE TABLE sde_factions (
    faction_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    corporation_id BIGINT,
    icon_id BIGINT
);

CREATE TABLE sde_npc_corporations (
    corporation_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    faction_id BIGINT,
    icon_id BIGINT
);

CREATE TABLE sde_npc_corporation_divisions (
    corporation_id BIGINT NOT NULL,
    division_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY (corporation_id, division_id)
);

CREATE TABLE sde_agents (
    agent_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    corporation_id BIGINT,
    division_id BIGINT,
    level INT
);

CREATE TABLE sde_agents_in_space (
    agent_id BIGINT PRIMARY KEY NOT NULL,
    solar_system_id BIGINT
);

CREATE TABLE sde_races (
    race_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    icon_id BIGINT
);

CREATE TABLE sde_bloodlines (
    bloodline_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    race_id BIGINT,
    description TEXT,
    icon_id BIGINT
);

CREATE TABLE sde_ancestries (
    ancestry_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    bloodline_id BIGINT,
    description TEXT,
    icon_id BIGINT
);

COMMIT;
