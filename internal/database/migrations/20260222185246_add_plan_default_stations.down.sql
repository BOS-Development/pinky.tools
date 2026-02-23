-- Migration: add_plan_default_stations
-- Created: Sun Feb 22 06:52:46 PM PST 2026

alter table production_plans drop column default_manufacturing_station_id;
alter table production_plans drop column default_reaction_station_id;
