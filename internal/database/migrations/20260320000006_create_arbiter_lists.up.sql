-- Migration: create_arbiter_lists
-- Blacklist: items the user never wants Arbiter to build (always buy instead).
-- Whitelist: items the user always wants Arbiter to build (ignore profitability gate).
--
-- Both use composite PKs for natural deduplication and efficient per-user lookup.
-- No FK on type_id — consistent with market_prices pattern (type_ids are SDE-sourced
-- and may not all exist in asset_item_types).

begin;

create table arbiter_blacklist (
    user_id     bigint          not null references users(id) on delete cascade,
    type_id     bigint          not null,
    added_at    timestamptz     not null default now(),
    primary key (user_id, type_id)
);

create table arbiter_whitelist (
    user_id     bigint          not null references users(id) on delete cascade,
    type_id     bigint          not null,
    added_at    timestamptz     not null default now(),
    primary key (user_id, type_id)
);

commit;
