BEGIN;

CREATE TABLE sde_dogma_attribute_categories (
    category_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT
);

CREATE TABLE sde_dogma_attributes (
    attribute_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT,
    default_value DOUBLE PRECISION,
    display_name TEXT,
    category_id BIGINT,
    high_is_good BOOLEAN,
    stackable BOOLEAN,
    published BOOLEAN,
    unit_id BIGINT
);

CREATE TABLE sde_dogma_effects (
    effect_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT,
    description TEXT,
    display_name TEXT,
    category_id BIGINT
);

CREATE TABLE sde_type_dogma_attributes (
    type_id BIGINT NOT NULL,
    attribute_id BIGINT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (type_id, attribute_id)
);

CREATE TABLE sde_type_dogma_effects (
    type_id BIGINT NOT NULL,
    effect_id BIGINT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (type_id, effect_id)
);

COMMIT;
