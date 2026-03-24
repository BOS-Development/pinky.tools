begin;

create table arbiter_manufacturing_profiles (
	id          bigint      primary key generated always as identity,
	user_id     bigint      not null references users(id) on delete cascade,
	name        varchar(100) not null,

	reaction_structure  text,
	reaction_rig        text,
	reaction_system_id  bigint      references solar_systems(solar_system_id) on delete set null,
	reaction_facility_tax double precision,

	invention_structure  text,
	invention_rig        text,
	invention_system_id  bigint      references solar_systems(solar_system_id) on delete set null,
	invention_facility_tax double precision,

	component_structure  text,
	component_rig        text,
	component_system_id  bigint      references solar_systems(solar_system_id) on delete set null,
	component_facility_tax double precision,

	final_structure  text,
	final_rig        text,
	final_system_id  bigint      references solar_systems(solar_system_id) on delete set null,
	final_facility_tax double precision,

	created_at  timestamptz not null default now(),
	updated_at  timestamptz not null default now()
);

create unique index idx_arbiter_manufacturing_profiles_user_name
	on arbiter_manufacturing_profiles (user_id, name);

create index idx_arbiter_manufacturing_profiles_user_id
	on arbiter_manufacturing_profiles (user_id);

commit;
