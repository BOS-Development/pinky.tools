-- Migration: backfill_job_slot_browse_permissions
-- Created: Sat Feb 28 05:04:04 PM PST 2026

-- Backfill job_slot_browse permission rows for existing contacts
-- Mirrors for_sale_browse rows to ensure users can see job slot listings
insert into contact_permissions (contact_id, granting_user_id, receiving_user_id, service_type, can_access, created_at, updated_at)
select
	contact_id,
	granting_user_id,
	receiving_user_id,
	'job_slot_browse' as service_type,
	can_access,
	now() as created_at,
	now() as updated_at
from contact_permissions
where service_type = 'for_sale_browse'
on conflict (contact_id, granting_user_id, receiving_user_id, service_type) do nothing;
