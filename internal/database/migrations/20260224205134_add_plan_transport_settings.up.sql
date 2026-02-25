-- Migration: add_plan_transport_settings
-- Created: Tue Feb 24 08:51:34 PM PST 2026

alter table production_plans
	add column transport_fulfillment text,
	add column transport_method text,
	add column transport_profile_id bigint references transport_profiles(id) on delete set null,
	add column courier_rate_per_m3 numeric(12,2) not null default 0,
	add column courier_collateral_rate numeric(6,4) not null default 0;
