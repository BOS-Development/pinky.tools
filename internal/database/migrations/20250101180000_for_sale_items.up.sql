CREATE TABLE for_sale_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    owner_type VARCHAR(20) NOT NULL,
    owner_id BIGINT NOT NULL,
    location_id BIGINT NOT NULL,
    container_id BIGINT,
    division_number INT,
    quantity_available BIGINT NOT NULL,
    price_per_unit BIGINT NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT for_sale_positive_quantity CHECK (quantity_available > 0),
    CONSTRAINT for_sale_positive_price CHECK (price_per_unit >= 0)
);

CREATE UNIQUE INDEX idx_for_sale_unique ON for_sale_items(
    user_id, type_id, owner_type, owner_id, location_id,
    COALESCE(container_id, 0), COALESCE(division_number, 0)
) WHERE is_active = true;

CREATE INDEX idx_for_sale_user ON for_sale_items(user_id);
CREATE INDEX idx_for_sale_active ON for_sale_items(is_active);
CREATE INDEX idx_for_sale_type ON for_sale_items(type_id);
