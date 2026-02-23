-- Migration: create_user_stations
-- Created: Sun Feb 22 05:53:30 PM PST 2026

alter table production_plan_steps drop column if exists user_station_id;
drop table if exists user_station_services;
drop table if exists user_station_rigs;
drop table if exists user_stations;
