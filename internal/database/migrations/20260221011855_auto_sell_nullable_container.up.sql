BEGIN;

-- Make container_id nullable so auto-sell can target corp hangar divisions directly
ALTER TABLE auto_sell_containers ALTER COLUMN container_id DROP NOT NULL;

-- Rebuild unique index to use COALESCE for nullable container_id (matches auto_buy_configs pattern)
DROP INDEX idx_auto_sell_unique_container;
CREATE UNIQUE INDEX idx_auto_sell_unique_container ON auto_sell_containers(
    user_id, owner_type, owner_id, location_id,
    coalesce(container_id, 0), coalesce(division_number, 0)
) WHERE is_active = true;

COMMIT;
