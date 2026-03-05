-- Migration: create_hauling_structure_snapshots
-- Created: Thu Mar  5 02:32:56 AM PST 2026

create table hauling_structure_snapshots (
	type_id bigint not null,
	structure_id bigint not null,
	buy_price double precision,
	sell_price double precision,
	volume_available bigint,
	avg_daily_volume double precision,
	updated_at timestamptz not null default now(),
	primary key (type_id, structure_id)
);

create index idx_hauling_structure_snapshots_structure_updated
	on hauling_structure_snapshots(structure_id, updated_at);
