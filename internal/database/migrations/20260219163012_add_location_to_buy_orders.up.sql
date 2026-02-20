-- Migration: add_location_to_buy_orders
-- Created: Thu Feb 19 04:30:12 PM PST 2026

delete from buy_orders;
alter table buy_orders add column location_id bigint not null;
