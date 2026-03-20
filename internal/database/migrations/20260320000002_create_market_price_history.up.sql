-- Migration: create_market_price_history
-- Daily snapshots of Jita market data per type.
-- Used by Arbiter to calculate 30-day average daily volume (demand) and D.O.S.
-- D.O.S. = current order_book_volume / avg(daily_volume over last 30 days).
-- No FK on type_id — consistent with market_prices after remove_market_prices_fk.

begin;

create table market_price_history (
    type_id             bigint              not null,
    snapshot_date       date                not null,
    buy_price           double precision,
    sell_price          double precision,
    order_book_volume   bigint,
    daily_volume        bigint,

    primary key (type_id, snapshot_date)
);

-- Support efficient lookups for a single item's recent history (30-day window queries).
create index idx_market_price_history_type_date
    on market_price_history(type_id, snapshot_date desc);

-- Support bulk pruning of old snapshots by date.
create index idx_market_price_history_snapshot_date
    on market_price_history(snapshot_date);

commit;
