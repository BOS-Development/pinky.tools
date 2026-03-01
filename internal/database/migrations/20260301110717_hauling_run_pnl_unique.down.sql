-- Migration: hauling_run_pnl_unique (rollback)

alter table hauling_run_pnl
	drop constraint if exists hauling_run_pnl_run_type_unique;
