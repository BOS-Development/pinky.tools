create or replace view plan_run_job_counts as
select
    plan_run_id,
    count(*) as total,
    count(*) filter (where status = 'planned') as planned,
    count(*) filter (where status = 'active') as active,
    count(*) filter (where status = 'completed') as completed,
    count(*) filter (where status = 'cancelled') as cancelled
from industry_job_queue
where plan_run_id is not null
group by plan_run_id;
