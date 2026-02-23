-- Migration: add_step_output_location
-- Created: Sun Feb 22 10:02:29 PM PST 2026

alter table production_plan_steps add column output_owner_type text;
alter table production_plan_steps add column output_owner_id bigint;
alter table production_plan_steps add column output_division_number int;
alter table production_plan_steps add column output_container_id bigint;
