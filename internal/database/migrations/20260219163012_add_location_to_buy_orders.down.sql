-- Migration: add_location_to_buy_orders
-- Created: Thu Feb 19 04:30:12 PM PST 2026

alter table buy_orders drop column location_id;
