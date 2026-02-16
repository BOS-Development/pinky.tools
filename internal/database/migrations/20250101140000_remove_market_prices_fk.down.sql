BEGIN;

-- Restore foreign key constraint
-- Note: This will fail if there are market prices for type_ids not in asset_item_types
ALTER TABLE market_prices ADD CONSTRAINT market_prices_type_id_fkey
    FOREIGN KEY (type_id) REFERENCES asset_item_types(type_id);

COMMIT;
