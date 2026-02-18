-- Migration: decimal_prices
-- Created: Tue Feb 17 06:38:56 PM PST 2026

alter table for_sale_items alter column price_per_unit type numeric(20,2);
alter table purchase_transactions alter column price_per_unit type numeric(20,2);
alter table purchase_transactions alter column total_price type numeric(20,2);
alter table buy_orders alter column max_price_per_unit type numeric(20,2);
