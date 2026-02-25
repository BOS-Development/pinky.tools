-- Migration: create_transport_tables
-- Created: Tue Feb 24 10:42:04 AM PST 2026

alter table industry_job_queue drop column if exists transport_job_id;
drop table if exists transport_trigger_config;
drop table if exists transport_job_items;
drop table if exists transport_jobs;
drop table if exists jf_route_waypoints;
drop table if exists jf_routes;
drop table if exists transport_profiles;
