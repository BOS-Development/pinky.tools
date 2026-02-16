BEGIN;

CREATE TABLE stockpile_markers (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    owner_type VARCHAR(20) NOT NULL,
    owner_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    container_id BIGINT,
    division_number INT,
    desired_quantity BIGINT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_stockpile_unique ON stockpile_markers(
    user_id, type_id, owner_type, owner_id, location_id,
    COALESCE(container_id, 0), COALESCE(division_number, 0)
);

CREATE INDEX idx_stockpile_user ON stockpile_markers(user_id);
CREATE INDEX idx_stockpile_type ON stockpile_markers(type_id);

COMMIT;
