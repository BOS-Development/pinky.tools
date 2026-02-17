BEGIN;

ALTER TABLE market_prices
    DROP COLUMN IF EXISTS adjusted_price;

ALTER TABLE asset_item_types
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS race_id,
    DROP COLUMN IF EXISTS graphic_id,
    DROP COLUMN IF EXISTS market_group_id,
    DROP COLUMN IF EXISTS published,
    DROP COLUMN IF EXISTS portion_size,
    DROP COLUMN IF EXISTS capacity,
    DROP COLUMN IF EXISTS mass,
    DROP COLUMN IF EXISTS packaged_volume,
    DROP COLUMN IF EXISTS group_id;

COMMIT;
