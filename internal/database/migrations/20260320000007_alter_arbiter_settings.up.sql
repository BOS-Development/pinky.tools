-- Migration: alter_arbiter_settings
-- Removes the four hardcoded security class text columns (reaction_security,
-- invention_security, component_security, final_security). Security class is now
-- derived at query time from solar_systems.security via the system_id columns already
-- present on the table.
--
-- Adds:
--   use_whitelist / use_blacklist: toggle Arbiter list behaviour per user.
--   decryptor_type_id: preferred decryptor FK into sde_decryptors. NULL = maximize ROI.
--   default_scope_id: the scope used by default in the Arbiter UI.

begin;

alter table arbiter_settings
    drop column if exists reaction_security,
    drop column if exists invention_security,
    drop column if exists component_security,
    drop column if exists final_security;

alter table arbiter_settings
    add column if not exists use_whitelist       boolean not null default true,
    add column if not exists use_blacklist       boolean not null default true,
    add column if not exists decryptor_type_id   bigint  references sde_decryptors(type_id) on delete set null,
    add column if not exists default_scope_id    bigint  references arbiter_scopes(id) on delete set null;

commit;
