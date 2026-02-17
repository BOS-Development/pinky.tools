BEGIN;

CREATE TABLE sde_blueprints (
    blueprint_type_id BIGINT PRIMARY KEY NOT NULL,
    max_production_limit INT
);

CREATE TABLE sde_blueprint_activities (
    blueprint_type_id BIGINT NOT NULL,
    activity TEXT NOT NULL,
    time INT NOT NULL,
    PRIMARY KEY (blueprint_type_id, activity)
);

CREATE TABLE sde_blueprint_materials (
    blueprint_type_id BIGINT NOT NULL,
    activity TEXT NOT NULL,
    type_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (blueprint_type_id, activity, type_id)
);

CREATE TABLE sde_blueprint_products (
    blueprint_type_id BIGINT NOT NULL,
    activity TEXT NOT NULL,
    type_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    probability DOUBLE PRECISION,
    PRIMARY KEY (blueprint_type_id, activity, type_id)
);

CREATE TABLE sde_blueprint_skills (
    blueprint_type_id BIGINT NOT NULL,
    activity TEXT NOT NULL,
    type_id BIGINT NOT NULL,
    level INT NOT NULL,
    PRIMARY KEY (blueprint_type_id, activity, type_id)
);

COMMIT;
