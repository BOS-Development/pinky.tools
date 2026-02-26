create index if not exists idx_job_queue_user_status on industry_job_queue (user_id, status);
create index if not exists idx_esi_jobs_user_status on esi_industry_jobs (user_id, status);
create index if not exists idx_buy_orders_location on buy_orders(location_id);
create index if not exists idx_purchase_contract_polling on purchase_transactions(status) where status = 'contract_created' and contract_key is not null;
create index if not exists idx_for_sale_user_active on for_sale_items(user_id, is_active);
create index if not exists idx_plan_runs_plan_user on production_plan_runs(plan_id, user_id);
drop index if exists idx_job_queue_plan_run_id;
create index if not exists idx_job_queue_plan_run_status on industry_job_queue(plan_run_id, status) where plan_run_id is not null;
create index if not exists idx_sde_blueprint_products_type_id on sde_blueprint_products(type_id, activity);
