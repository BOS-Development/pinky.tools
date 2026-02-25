-- Migration: add_sort_order_to_job_queue (rollback)

alter table industry_job_queue drop column sort_order;
alter table industry_job_queue drop column station_name;
alter table industry_job_queue drop column input_location;
alter table industry_job_queue drop column output_location;
