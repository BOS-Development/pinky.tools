BEGIN;

-- Delete any rows with NULL container_id before restoring NOT NULL
DELETE FROM auto_sell_containers WHERE container_id IS NULL;

-- Restore original unique index with non-nullable container_id
DROP INDEX idx_auto_sell_unique_container;
CREATE UNIQUE INDEX idx_auto_sell_unique_container ON auto_sell_containers(
    user_id, owner_type, owner_id, location_id, container_id,
    coalesce(division_number, 0)
) WHERE is_active = true;

-- Restore NOT NULL constraint
ALTER TABLE auto_sell_containers ALTER COLUMN container_id SET NOT NULL;

COMMIT;
