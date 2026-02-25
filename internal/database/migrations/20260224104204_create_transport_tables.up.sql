-- Migration: create_transport_tables
-- Created: Tue Feb 24 10:42:04 AM PST 2026

-- Transport profiles: per-ship/character transport configurations
create table transport_profiles (
	id bigserial primary key,
	user_id bigint not null references users(id),
	name text not null,
	transport_method text not null,
	character_id bigint,
	cargo_m3 double precision not null,
	rate_per_m3_per_jump double precision not null default 0,
	collateral_rate double precision not null default 0,
	collateral_price_basis text not null default 'sell',
	fuel_type_id bigint,
	fuel_per_ly double precision,
	fuel_conservation_level int not null default 0,
	route_preference text not null default 'shortest',
	is_default boolean not null default false,
	created_at timestamptz not null default now()
);

create index idx_transport_profiles_user_id on transport_profiles(user_id);

-- JF routes: user-defined jump freighter routes
create table jf_routes (
	id bigserial primary key,
	user_id bigint not null references users(id),
	name text not null,
	origin_system_id bigint not null,
	destination_system_id bigint not null,
	total_distance_ly double precision not null default 0,
	created_at timestamptz not null default now()
);

create index idx_jf_routes_user_id on jf_routes(user_id);

-- JF route waypoints: ordered cyno stops per route
create table jf_route_waypoints (
	id bigserial primary key,
	route_id bigint not null references jf_routes(id) on delete cascade,
	sequence int not null,
	system_id bigint not null,
	distance_ly double precision not null default 0
);

create index idx_jf_route_waypoints_route_id on jf_route_waypoints(route_id);

-- Transport jobs: transport job instances
create table transport_jobs (
	id bigserial primary key,
	user_id bigint not null references users(id),
	origin_station_id bigint not null,
	destination_station_id bigint not null,
	origin_system_id bigint not null,
	destination_system_id bigint not null,
	transport_method text not null,
	route_preference text not null default 'shortest',
	status text not null default 'planned',
	total_volume_m3 double precision not null default 0,
	total_collateral double precision not null default 0,
	estimated_cost double precision not null default 0,
	jumps int not null default 0,
	distance_ly double precision,
	jf_route_id bigint references jf_routes(id),
	fulfillment_type text not null default 'self_haul',
	transport_profile_id bigint references transport_profiles(id),
	plan_run_id bigint references production_plan_runs(id),
	plan_step_id bigint,
	queue_entry_id bigint,
	notes text,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index idx_transport_jobs_user_id on transport_jobs(user_id);

-- Transport job items: items in a transport job
create table transport_job_items (
	id bigserial primary key,
	transport_job_id bigint not null references transport_jobs(id) on delete cascade,
	type_id bigint not null,
	quantity int not null,
	volume_m3 double precision not null default 0,
	estimated_value double precision not null default 0
);

create index idx_transport_job_items_job_id on transport_job_items(transport_job_id);

-- Transport trigger config: per-user, per-trigger fulfillment preferences
create table transport_trigger_config (
	user_id bigint not null references users(id),
	trigger_type text not null,
	default_fulfillment text not null default 'self_haul',
	allowed_fulfillments text[] not null default '{self_haul}',
	default_profile_id bigint references transport_profiles(id),
	default_method text,
	courier_rate_per_m3 double precision not null default 0,
	courier_collateral_rate double precision not null default 0,
	primary key (user_id, trigger_type)
);

-- Add transport_job_id to industry_job_queue
alter table industry_job_queue add column transport_job_id bigint references transport_jobs(id);
