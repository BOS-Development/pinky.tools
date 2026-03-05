-- Migration: hauling_runs_add_station_columns
-- Created: Thu Mar  5 02:32:56 AM PST 2026

alter table hauling_runs
	add column from_station_id bigint,
	add column to_station_id bigint;
