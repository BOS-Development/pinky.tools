-- Migration: cancel_duplicate_auto_fulfill_from_id_cycling
-- Created: Sat Feb 21 04:53:58 PM PST 2026

-- Cancel duplicate auto-fulfill purchases caused by buy order ID cycling.
-- When auto-buy deactivated and recreated buy orders with new IDs, the unique
-- index on (buy_order_id, for_sale_item_id) did not catch duplicates across
-- different buy_order_ids. This cancels the newer duplicates, keeping the oldest
-- purchase per (buyer_user_id, type_id, for_sale_item_id) combination.
update purchase_transactions
set status = 'cancelled'
where id in (
	select id from (
		select id,
			row_number() over (
				partition by buyer_user_id, type_id, for_sale_item_id
				order by purchased_at asc
			) as rn
		from purchase_transactions
		where is_auto_fulfilled = true
			and status in ('pending', 'contract_created')
	) ranked
	where rn > 1
);
