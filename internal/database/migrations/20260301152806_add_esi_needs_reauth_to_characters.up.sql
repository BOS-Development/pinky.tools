-- Migration: add_esi_needs_reauth_to_characters
-- Created: Sun Mar  1 03:28:06 PM PST 2026

alter table characters add column esi_needs_reauth boolean not null default false;
