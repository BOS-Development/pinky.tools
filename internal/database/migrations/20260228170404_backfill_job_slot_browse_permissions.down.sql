-- Migration: backfill_job_slot_browse_permissions
-- Created: Sat Feb 28 05:04:04 PM PST 2026

-- Remove all job_slot_browse permission rows
-- Safe to remove all since feature would be rolled back anyway
delete from contact_permissions where service_type = 'job_slot_browse';
