-- Migration: add_sort_order_to_job_queue
-- Adds sort_order for depth-based ordering, and location fields for job execution context

alter table industry_job_queue add column sort_order int not null default 0;
alter table industry_job_queue add column station_name text not null default '';
alter table industry_job_queue add column input_location text not null default '';
alter table industry_job_queue add column output_location text not null default '';
