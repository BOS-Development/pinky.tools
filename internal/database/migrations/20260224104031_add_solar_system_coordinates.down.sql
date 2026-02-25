-- Migration: add_solar_system_coordinates
-- Created: Tue Feb 24 10:40:31 AM PST 2026

alter table solar_systems drop column if exists x;
alter table solar_systems drop column if exists y;
alter table solar_systems drop column if exists z;
