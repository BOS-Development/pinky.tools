CREATE TABLE purchase_transactions (
    id BIGSERIAL PRIMARY KEY,
    for_sale_item_id BIGINT NOT NULL,  -- No FK constraint - preserve history even if listing deleted
    buyer_user_id BIGINT NOT NULL REFERENCES users(id),
    seller_user_id BIGINT NOT NULL REFERENCES users(id),
    type_id BIGINT NOT NULL REFERENCES asset_item_types(type_id),
    quantity_purchased BIGINT NOT NULL,
    price_per_unit BIGINT NOT NULL,
    total_price BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    transaction_notes TEXT,
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT purchase_positive_quantity CHECK (quantity_purchased > 0),
    CONSTRAINT purchase_different_users CHECK (buyer_user_id != seller_user_id)
);

CREATE INDEX idx_purchase_buyer ON purchase_transactions(buyer_user_id, purchased_at DESC);
CREATE INDEX idx_purchase_seller ON purchase_transactions(seller_user_id, purchased_at DESC);
CREATE INDEX idx_purchase_item ON purchase_transactions(for_sale_item_id);
