-- Migration: create_user_trading_structures
-- Created: Thu Mar  5 02:32:56 AM PST 2026

create table user_trading_structures (
	id bigserial primary key,
	user_id bigint not null references users(id) on delete cascade,
	structure_id bigint not null,
	name varchar(255) not null,
	system_id bigint not null,
	region_id bigint not null,
	character_id bigint not null,
	access_ok boolean not null default true,
	last_scanned_at timestamptz,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create unique index idx_user_trading_structures_unique on user_trading_structures(user_id, structure_id);
create index idx_user_trading_structures_user_id on user_trading_structures(user_id);
