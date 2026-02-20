-- Migration: auto_buy_price_range
-- Created: Fri Feb 20 12:51:30 AM PST 2026

alter table buy_orders drop column min_price_per_unit;
alter table auto_buy_configs drop column min_price_percentage;
alter table auto_buy_configs rename column max_price_percentage to price_percentage;
