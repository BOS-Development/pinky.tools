-- Migration: auto_fulfill_purchases
-- Created: Fri Feb 20 12:51:34 AM PST 2026

alter table purchase_transactions
	add column buy_order_id bigint references buy_orders(id),
	add column is_auto_fulfilled boolean not null default false;

-- Prevent duplicate auto-fulfills for same buy_order + for_sale_item pair
create unique index idx_auto_fulfill_unique
	on purchase_transactions(buy_order_id, for_sale_item_id)
	where is_auto_fulfilled = true and status = 'pending';
