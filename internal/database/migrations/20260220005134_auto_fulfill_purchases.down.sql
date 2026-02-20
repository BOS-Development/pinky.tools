-- Migration: auto_fulfill_purchases
-- Created: Fri Feb 20 12:51:34 AM PST 2026

drop index if exists idx_auto_fulfill_unique;
alter table purchase_transactions drop column is_auto_fulfilled;
alter table purchase_transactions drop column buy_order_id;
