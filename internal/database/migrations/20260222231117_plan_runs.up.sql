create table production_plan_runs (
	id		bigserial	primary key,
	plan_id		bigint		not null references production_plans(id) on delete cascade,
	user_id		bigint		not null references users(id),
	quantity	int		not null,
	created_at	timestamptz	not null default now()
);

create index idx_plan_runs_plan_id on production_plan_runs(plan_id);
create index idx_plan_runs_user_id on production_plan_runs(user_id);

alter table industry_job_queue add column plan_run_id bigint references production_plan_runs(id) on delete set null;
alter table industry_job_queue add column plan_step_id bigint;

create index idx_job_queue_plan_run_id on industry_job_queue(plan_run_id);
