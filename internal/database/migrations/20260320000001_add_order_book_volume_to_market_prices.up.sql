-- Migration: add_order_book_volume_to_market_prices
-- Adds total sell order quantity for Jita to market_prices.
-- Used by Arbiter for D.O.S. calculation: order_book_volume / 30d avg daily volume.
-- Populated during the market price refresh alongside buy/sell prices.

alter table market_prices
    add column if not exists order_book_volume bigint;
