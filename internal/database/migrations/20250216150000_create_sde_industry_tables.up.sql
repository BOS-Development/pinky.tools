BEGIN;

CREATE TABLE sde_planet_schematics (
    schematic_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    cycle_time INT NOT NULL
);

CREATE TABLE sde_planet_schematic_types (
    schematic_id BIGINT NOT NULL,
    type_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    is_input BOOLEAN NOT NULL,
    PRIMARY KEY (schematic_id, type_id)
);

CREATE TABLE sde_control_tower_resources (
    control_tower_type_id BIGINT NOT NULL,
    resource_type_id BIGINT NOT NULL,
    purpose INT,
    quantity INT NOT NULL,
    min_security DOUBLE PRECISION,
    faction_id BIGINT,
    PRIMARY KEY (control_tower_type_id, resource_type_id)
);

CREATE TABLE industry_cost_indices (
    system_id BIGINT NOT NULL,
    activity TEXT NOT NULL,
    cost_index DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (system_id, activity)
);

COMMIT;
