-- Migration: cancel_stale_auto_fulfill_purchases
-- Created: Sat Feb 21 12:45:55 PM PST 2026

-- Cancel all pending auto-fulfilled purchases where the for-sale item
-- is no longer active. These are leftovers from the duplicate purchase bug
-- where auto-sell re-listed items, auto-fulfill matched again, and the
-- original listing was eventually deactivated (quantity reached 0).

update purchase_transactions
set status = 'cancelled'
where is_auto_fulfilled = true
	and status = 'pending'
	and for_sale_item_id in (
		select id from for_sale_items where is_active = false
	);
