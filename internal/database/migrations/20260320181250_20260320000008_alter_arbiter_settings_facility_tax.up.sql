-- Migration: 20260320000008_alter_arbiter_settings_facility_tax
-- Created: Fri Mar 20 06:12:50 PM PDT 2026

alter table arbiter_settings
    add column final_facility_tax     double precision not null default 0,
    add column component_facility_tax double precision not null default 0,
    add column reaction_facility_tax  double precision not null default 0;
