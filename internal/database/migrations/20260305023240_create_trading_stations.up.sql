-- Migration: create_trading_stations
-- Created: Thu Mar  5 02:32:40 AM PST 2026

create table trading_stations (
	id bigserial primary key,
	station_id bigint not null,
	name varchar(255) not null,
	system_id bigint not null,
	region_id bigint not null,
	is_preset boolean not null default false,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create unique index idx_trading_stations_station_id on trading_stations(station_id);
create index idx_trading_stations_region_id on trading_stations(region_id);

insert into trading_stations (station_id, name, system_id, region_id, is_preset)
values (60003760, 'Jita IV - Moon 4 - Caldari Navy Assembly Plant', 30000142, 10000002, true);
