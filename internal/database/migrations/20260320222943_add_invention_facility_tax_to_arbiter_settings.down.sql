-- Migration: add_invention_facility_tax_to_arbiter_settings
-- Created: Fri Mar 20 10:29:43 PM PDT 2026

alter table arbiter_settings drop column if exists invention_facility_tax;
