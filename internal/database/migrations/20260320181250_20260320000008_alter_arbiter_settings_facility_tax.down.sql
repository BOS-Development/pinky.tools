-- Migration: 20260320000008_alter_arbiter_settings_facility_tax
-- Created: Fri Mar 20 06:12:50 PM PDT 2026

alter table arbiter_settings
    drop column final_facility_tax,
    drop column component_facility_tax,
    drop column reaction_facility_tax;
