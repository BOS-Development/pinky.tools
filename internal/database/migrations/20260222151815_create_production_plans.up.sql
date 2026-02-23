create table production_plans (
	id bigserial primary key,
	user_id bigint not null references users(id),
	product_type_id bigint not null,
	name text not null default '',
	notes text,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index idx_production_plans_user_id on production_plans(user_id);

create table production_plan_steps (
	id bigserial primary key,
	plan_id bigint not null references production_plans(id) on delete cascade,
	parent_step_id bigint references production_plan_steps(id) on delete cascade,
	product_type_id bigint not null,
	blueprint_type_id bigint not null,
	activity text not null,
	me_level int not null default 10,
	te_level int not null default 20,
	industry_skill int not null default 5,
	adv_industry_skill int not null default 5,
	structure text not null default 'raitaru',
	rig text not null default 't2',
	security text not null default 'high',
	facility_tax numeric(5, 2) not null default 1.0,
	system_id bigint,
	source_location_id bigint,
	source_container_id bigint,
	source_division_number int,
	source_owner_type text,
	source_owner_id bigint
);

create index idx_production_plan_steps_plan_id on production_plan_steps(plan_id);
create index idx_production_plan_steps_parent on production_plan_steps(parent_step_id);
