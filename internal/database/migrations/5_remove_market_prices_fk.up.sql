BEGIN;

-- Drop foreign key constraint to allow storing market prices for all items,
-- not just those in asset_item_types table
ALTER TABLE market_prices DROP CONSTRAINT IF EXISTS market_prices_type_id_fkey;

COMMIT;
