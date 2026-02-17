BEGIN;

CREATE TABLE sde_categories (
    category_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    published BOOLEAN NOT NULL DEFAULT false,
    icon_id BIGINT
);

CREATE TABLE sde_groups (
    group_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    category_id BIGINT NOT NULL,
    published BOOLEAN NOT NULL DEFAULT false,
    icon_id BIGINT
);

CREATE TABLE sde_meta_groups (
    meta_group_id BIGINT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE sde_market_groups (
    market_group_id BIGINT PRIMARY KEY NOT NULL,
    parent_group_id BIGINT,
    name TEXT NOT NULL,
    description TEXT,
    icon_id BIGINT,
    has_types BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE sde_icons (
    icon_id BIGINT PRIMARY KEY NOT NULL,
    description TEXT
);

CREATE TABLE sde_graphics (
    graphic_id BIGINT PRIMARY KEY NOT NULL,
    description TEXT
);

CREATE TABLE sde_metadata (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMIT;
