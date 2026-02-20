-- Migration: add_corporation_id_to_characters
-- Created: Fri Feb 20 12:14:37 AM PST 2026

alter table characters drop column corporation_id;
