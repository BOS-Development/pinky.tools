alter table production_plan_steps drop column system_id;
alter table production_plan_steps add column station_name text;
