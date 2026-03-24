-- Migration: add_arbiter_enabled_to_users
-- Gates the Arbiter feature — only users with arbiter_enabled = true can access it.

alter table users
	add column if not exists arbiter_enabled boolean not null default false;
