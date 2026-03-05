-- Migration: hauling_run_items_sell_tracking
-- Created: Thu Mar  5 12:47:40 AM PST 2026

alter table hauling_run_items
	add column sell_order_id bigint,
	add column qty_sold bigint not null default 0,
	add column actual_sell_price_isk double precision;

alter table hauling_run_items
	add constraint hauling_run_items_non_negative_sold check (qty_sold >= 0);
