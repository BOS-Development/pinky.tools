-- Migration: widen_auto_fulfill_unique_index
-- Created: Sat Feb 21 11:38:33 AM PST 2026

-- Revert to the original narrower unique index (pending only)

drop index if exists idx_auto_fulfill_unique;

create unique index idx_auto_fulfill_unique
	on purchase_transactions(buy_order_id, for_sale_item_id)
	where is_auto_fulfilled = true and status = 'pending';
