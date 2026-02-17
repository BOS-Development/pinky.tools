BEGIN;

CREATE TABLE sde_skins (
    skin_id BIGINT PRIMARY KEY NOT NULL,
    type_id BIGINT,
    material_id BIGINT
);

CREATE TABLE sde_skin_licenses (
    license_type_id BIGINT PRIMARY KEY NOT NULL,
    skin_id BIGINT,
    duration INT
);

CREATE TABLE sde_skin_materials (
    skin_material_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT
);

CREATE TABLE sde_certificates (
    certificate_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT,
    group_id BIGINT
);

CREATE TABLE sde_landmarks (
    landmark_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT
);

CREATE TABLE sde_station_operations (
    operation_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT
);

CREATE TABLE sde_station_services (
    service_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT
);

CREATE TABLE sde_contraband_types (
    faction_id BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    standing_loss DOUBLE PRECISION,
    fine_by_value DOUBLE PRECISION,
    PRIMARY KEY (faction_id, type_id)
);

CREATE TABLE sde_research_agents (
    agent_id BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    PRIMARY KEY (agent_id, type_id)
);

CREATE TABLE sde_character_attributes (
    attribute_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT,
    icon_id BIGINT
);

CREATE TABLE sde_corporation_activities (
    activity_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT
);

CREATE TABLE sde_tournament_rule_sets (
    rule_set_id BIGINT PRIMARY KEY NOT NULL,
    data JSONB
);

COMMIT;
