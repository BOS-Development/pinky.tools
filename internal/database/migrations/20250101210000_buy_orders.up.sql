CREATE TABLE buy_orders (
    id BIGSERIAL PRIMARY KEY,
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_desired BIGINT NOT NULL,
    max_price_per_unit BIGINT NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT buy_order_positive_quantity CHECK (quantity_desired > 0),
    CONSTRAINT buy_order_positive_price CHECK (max_price_per_unit >= 0)
);

CREATE INDEX idx_buy_orders_buyer ON buy_orders(buyer_user_id);
CREATE INDEX idx_buy_orders_type ON buy_orders(type_id);
CREATE INDEX idx_buy_orders_active ON buy_orders(is_active);
CREATE INDEX idx_buy_orders_buyer_active ON buy_orders(buyer_user_id, is_active);
