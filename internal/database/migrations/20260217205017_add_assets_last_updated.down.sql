-- Migration: add_assets_last_updated
-- Created: Tue Feb 17 08:50:17 PM PST 2026

alter table users drop column assets_last_updated_at;
