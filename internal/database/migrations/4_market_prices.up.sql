BEGIN;

CREATE TABLE market_prices (
    type_id BIGINT PRIMARY KEY REFERENCES asset_item_types(type_id),
    region_id BIGINT NOT NULL,
    buy_price DOUBLE PRECISION,
    sell_price DOUBLE PRECISION,
    daily_volume BIGINT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_market_prices_region ON market_prices(region_id);
CREATE INDEX idx_market_prices_updated ON market_prices(updated_at);

COMMIT;
