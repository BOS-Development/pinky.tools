-- Migration: hauling_run_items_sell_tracking
-- Created: Thu Mar  5 12:47:40 AM PST 2026

alter table hauling_run_items
	drop constraint if exists hauling_run_items_non_negative_sold,
	drop column if exists actual_sell_price_isk,
	drop column if exists qty_sold,
	drop column if exists sell_order_id;
