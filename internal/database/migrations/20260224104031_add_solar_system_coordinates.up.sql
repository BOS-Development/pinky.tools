-- Migration: add_solar_system_coordinates
-- Created: Tue Feb 24 10:40:31 AM PST 2026

alter table solar_systems add column x double precision;
alter table solar_systems add column y double precision;
alter table solar_systems add column z double precision;
