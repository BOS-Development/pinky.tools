-- Migration: widen_auto_fulfill_unique_index
-- Created: Sat Feb 21 11:38:33 AM PST 2026

-- Widen the auto-fulfill unique index to also cover 'contract_created' status.
-- The old index only prevented duplicates while status = 'pending', allowing
-- a new pending purchase for the same (buy_order_id, for_sale_item_id) pair
-- once the original moved to 'contract_created'.

drop index if exists idx_auto_fulfill_unique;

create unique index idx_auto_fulfill_unique
	on purchase_transactions(buy_order_id, for_sale_item_id)
	where is_auto_fulfilled = true and status in ('pending', 'contract_created');
