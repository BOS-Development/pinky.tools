-- Migration: create_arbiter_tax_profile
-- Per-user pricing and fee configuration for Arbiter profit calculations.
--
-- trader_character_id: if set, sales_tax_rate is overridden at query time by
--   deriving the effective rate from the character's Accounting skill level.
--   Stored bare (no FK) because characters has a composite PK (id, user_id).
--
-- input_price_type / output_price_type: controls whether buy or sell price is used
--   when costing materials ('buy' or 'sell') and valuing outputs ('buy' or 'sell').
--
-- Default rates reflect Accounting 5 (3.6% sales tax) and typical broker fees.

begin;

create table arbiter_tax_profile (
    user_id                 bigint              primary key references users(id) on delete cascade,
    trader_character_id     bigint,
    sales_tax_rate          double precision    not null default 0.036,
    broker_fee_rate         double precision    not null default 0.03,
    structure_broker_fee    double precision    not null default 0.02,
    input_price_type        text                not null default 'sell',
    output_price_type       text                not null default 'buy',
    updated_at              timestamptz         not null default now()
);

commit;
