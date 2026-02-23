-- Migration: create_user_stations
-- Created: Sun Feb 22 05:53:30 PM PST 2026

create table user_stations (
	id bigserial primary key,
	user_id bigint not null references users(id),
	station_id bigint not null references stations(station_id),
	structure text not null default 'raitaru',
	facility_tax numeric(5, 2) not null default 1.0,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);
create index idx_user_stations_user_id on user_stations(user_id);

create table user_station_rigs (
	id bigserial primary key,
	user_station_id bigint not null references user_stations(id) on delete cascade,
	rig_name text not null,
	category text not null,
	tier text not null
);
create index idx_user_station_rigs_station_id on user_station_rigs(user_station_id);

create table user_station_services (
	id bigserial primary key,
	user_station_id bigint not null references user_stations(id) on delete cascade,
	service_name text not null,
	activity text not null
);
create index idx_user_station_services_station_id on user_station_services(user_station_id);

alter table production_plan_steps add column user_station_id bigint references user_stations(id) on delete set null;
