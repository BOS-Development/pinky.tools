-- Migration: auto_buy_price_range
-- Created: Fri Feb 20 12:51:30 AM PST 2026

-- Rename price_percentage → max_price_percentage on auto_buy_configs
alter table auto_buy_configs rename column price_percentage to max_price_percentage;
alter table auto_buy_configs add column min_price_percentage numeric(5, 2) not null default 0.00;

-- Add min_price_per_unit to buy_orders
alter table buy_orders add column min_price_per_unit numeric(20, 2) not null default 0.00;
