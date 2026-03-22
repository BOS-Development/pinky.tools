-- Migration: add_adjusted_prices_table
-- Created: Sun Mar 22 12:25:00 AM PDT 2026

BEGIN;

create table if not exists adjusted_prices (
	type_id			bigint			primary key,
	adjusted_price	double precision	not null,
	updated_at		timestamp		not null default now()
);

-- Backfill from existing market_prices data
insert into adjusted_prices (type_id, adjusted_price, updated_at)
select type_id, adjusted_price, updated_at
from market_prices
where adjusted_price is not null
on conflict (type_id) do update
	set adjusted_price = excluded.adjusted_price,
		updated_at = excluded.updated_at;

COMMIT;
