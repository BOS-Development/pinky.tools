-- Run this SQL script on ALL databases (dev, staging, production)
-- AFTER renaming migration files to update schema_migrations table

BEGIN;

-- Update version numbers to new timestamp format
UPDATE schema_migrations SET version = 20250101100000 WHERE version = 1;
UPDATE schema_migrations SET version = 20250101110000 WHERE version = 2;
UPDATE schema_migrations SET version = 20250101120000 WHERE version = 3;
UPDATE schema_migrations SET version = 20250101130000 WHERE version = 4;
UPDATE schema_migrations SET version = 20250101140000 WHERE version = 5;
UPDATE schema_migrations SET version = 20250101150000 WHERE version = 6;
UPDATE schema_migrations SET version = 20250101160000 WHERE version = 7;
UPDATE schema_migrations SET version = 20250101170000 WHERE version = 8;
UPDATE schema_migrations SET version = 20250101180000 WHERE version = 9;
UPDATE schema_migrations SET version = 20250101190000 WHERE version = 10;
UPDATE schema_migrations SET version = 20250101200000 WHERE version = 11;
UPDATE schema_migrations SET version = 20250101210000 WHERE version = 12;

-- Verify updates
SELECT version FROM schema_migrations ORDER BY version;

COMMIT;
