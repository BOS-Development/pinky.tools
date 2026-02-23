-- Migration: add_step_output_location
-- Created: Sun Feb 22 10:02:29 PM PST 2026

alter table production_plan_steps drop column output_owner_type;
alter table production_plan_steps drop column output_owner_id;
alter table production_plan_steps drop column output_division_number;
alter table production_plan_steps drop column output_container_id;
