drop index if exists idx_job_queue_plan_run_id;
alter table industry_job_queue drop column if exists plan_step_id;
alter table industry_job_queue drop column if exists plan_run_id;
drop table if exists production_plan_runs;
