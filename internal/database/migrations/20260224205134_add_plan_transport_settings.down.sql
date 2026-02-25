-- Migration: add_plan_transport_settings
-- Created: Tue Feb 24 08:51:34 PM PST 2026

alter table production_plans
	drop column transport_fulfillment,
	drop column transport_method,
	drop column transport_profile_id,
	drop column courier_rate_per_m3,
	drop column courier_collateral_rate;
