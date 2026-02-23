alter table production_plan_steps drop column station_name;
alter table production_plan_steps add column system_id bigint;
