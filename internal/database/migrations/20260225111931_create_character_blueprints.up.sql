-- Migration: create_character_blueprints
-- Created: Wed Feb 25 11:19:31 AM PST 2026

CREATE TABLE character_blueprints (
    item_id              BIGINT       NOT NULL PRIMARY KEY,
    user_id              BIGINT       NOT NULL,
    owner_id             BIGINT       NOT NULL,
    owner_type           TEXT         NOT NULL,
    type_id              BIGINT       NOT NULL,
    location_id          BIGINT       NOT NULL,
    location_flag        TEXT         NOT NULL DEFAULT '',
    quantity             INT          NOT NULL DEFAULT 0,
    material_efficiency  INT          NOT NULL DEFAULT 0,
    time_efficiency      INT          NOT NULL DEFAULT 0,
    runs                 INT          NOT NULL DEFAULT -1,
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_character_blueprints_user_type ON character_blueprints (user_id, type_id);
CREATE INDEX idx_character_blueprints_owner ON character_blueprints (owner_id, owner_type);
