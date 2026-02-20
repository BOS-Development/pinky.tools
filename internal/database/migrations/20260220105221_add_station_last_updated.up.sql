alter table stations add column last_updated_at timestamp;

-- Backfill existing rows so they don't appear stale
update stations set last_updated_at = now();
