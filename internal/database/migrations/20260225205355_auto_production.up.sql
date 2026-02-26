-- Migration: auto_production
-- Created: Wed Feb 25 08:53:55 PM PST 2026

ALTER TABLE stockpile_markers ADD COLUMN plan_id BIGINT REFERENCES production_plans(id) ON DELETE SET NULL;
ALTER TABLE stockpile_markers ADD COLUMN auto_production_parallelism INT DEFAULT 0;
ALTER TABLE stockpile_markers ADD COLUMN auto_production_enabled BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_stockpile_markers_auto_production ON stockpile_markers(user_id) WHERE auto_production_enabled = TRUE;
