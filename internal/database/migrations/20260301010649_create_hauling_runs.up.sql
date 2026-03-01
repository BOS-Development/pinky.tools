-- Migration: create_hauling_runs

-- Hauling runs: multi-step workflow for profitable item hauling
create table hauling_runs (
	id bigserial primary key,
	user_id bigint not null references users(id),
	name text not null,
	status text not null default 'PLANNING',
	from_region_id bigint not null,
	from_system_id bigint,
	to_region_id bigint not null,
	max_volume_m3 double precision,
	haul_threshold_isk double precision,
	notify_tier2 boolean not null default false,
	notify_tier3 boolean not null default false,
	daily_digest boolean not null default false,
	notes text,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint hauling_runs_valid_status check (
		status in ('PLANNING', 'ACCUMULATING', 'READY', 'IN_TRANSIT', 'SELLING', 'COMPLETE', 'CANCELLED')
	)
);

create index idx_hauling_runs_user_id on hauling_runs(user_id);
create index idx_hauling_runs_status on hauling_runs(status);

-- Hauling run items: individual items tracked within a run
create table hauling_run_items (
	id bigserial primary key,
	run_id bigint not null references hauling_runs(id) on delete cascade,
	type_id bigint not null,
	type_name text not null,
	quantity_planned bigint not null default 1,
	quantity_acquired bigint not null default 0,
	buy_price_isk double precision,
	sell_price_isk double precision,
	volume_m3 double precision,
	character_id bigint,
	notes text,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint hauling_run_items_positive_planned check (quantity_planned > 0),
	constraint hauling_run_items_non_negative_acquired check (quantity_acquired >= 0),
	constraint hauling_run_items_unique_type unique (run_id, type_id)
);

create index idx_hauling_run_items_run_id on hauling_run_items(run_id);
create index idx_hauling_run_items_type_id on hauling_run_items(type_id);

-- Hauling market snapshots: cached market scanner data per type/region/system
-- system_id = 0 means region-wide snapshot; non-zero means system-specific
create table hauling_market_snapshots (
	type_id bigint not null,
	region_id bigint not null,
	system_id bigint not null default 0,
	buy_price double precision,
	sell_price double precision,
	volume_available bigint,
	avg_daily_volume double precision,
	days_to_sell double precision,
	updated_at timestamptz not null default now(),
	primary key (type_id, region_id, system_id)
);

create index idx_hauling_market_snapshots_region_system_updated
	on hauling_market_snapshots(region_id, system_id, updated_at);

-- Hauling run P&L: post-run profit and loss records (data model only, Phase 1)
create table hauling_run_pnl (
	id bigserial primary key,
	run_id bigint not null references hauling_runs(id) on delete cascade,
	type_id bigint not null,
	quantity_sold bigint not null default 0,
	avg_sell_price_isk double precision,
	total_revenue_isk double precision,
	total_cost_isk double precision,
	net_profit_isk double precision generated always as (total_revenue_isk - total_cost_isk) stored,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint hauling_run_pnl_non_negative_sold check (quantity_sold >= 0)
);

create index idx_hauling_run_pnl_run_id on hauling_run_pnl(run_id);
