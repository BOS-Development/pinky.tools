-- Migration: auto_production
-- Created: Wed Feb 25 08:53:55 PM PST 2026

DROP INDEX IF EXISTS idx_stockpile_markers_auto_production;
ALTER TABLE stockpile_markers DROP COLUMN IF EXISTS auto_production_enabled;
ALTER TABLE stockpile_markers DROP COLUMN IF EXISTS auto_production_parallelism;
ALTER TABLE stockpile_markers DROP COLUMN IF EXISTS plan_id;
