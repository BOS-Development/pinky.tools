-- Migration: widen_auto_fulfill_unique_index
-- Created: Sat Feb 21 11:38:33 AM PST 2026

-- Widen the auto-fulfill unique index to also cover 'contract_created' status.
-- The old index only prevented duplicates while status = 'pending', allowing
-- a new pending purchase for the same (buy_order_id, for_sale_item_id) pair
-- once the original moved to 'contract_created'.

-- First, cancel duplicate pending purchases that were created while the
-- original was already in contract_created status. For each duplicated
-- (buy_order_id, for_sale_item_id) pair, keep the oldest row and cancel the rest.
update purchase_transactions
set status = 'cancelled'
where id in (
	select id from (
		select id,
			row_number() over (
				partition by buy_order_id, for_sale_item_id
				order by purchased_at asc
			) as rn
		from purchase_transactions
		where is_auto_fulfilled = true
			and status in ('pending', 'contract_created')
			and buy_order_id is not null
	) ranked
	where rn > 1
);

drop index if exists idx_auto_fulfill_unique;

create unique index idx_auto_fulfill_unique
	on purchase_transactions(buy_order_id, for_sale_item_id)
	where is_auto_fulfilled = true and status in ('pending', 'contract_created');
