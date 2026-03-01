-- Migration: hauling_run_pnl_unique
-- Add unique constraint on (run_id, type_id) to enable ON CONFLICT upserts.

alter table hauling_run_pnl
	add constraint hauling_run_pnl_run_type_unique unique (run_id, type_id);
