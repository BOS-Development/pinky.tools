drop index if exists idx_sde_blueprint_products_type_id;
drop index if exists idx_job_queue_plan_run_status;
create index if not exists idx_job_queue_plan_run_id on industry_job_queue(plan_run_id);
drop index if exists idx_plan_runs_plan_user;
drop index if exists idx_for_sale_user_active;
drop index if exists idx_purchase_contract_polling;
drop index if exists idx_buy_orders_location;
drop index if exists idx_esi_jobs_user_status;
drop index if exists idx_job_queue_user_status;
