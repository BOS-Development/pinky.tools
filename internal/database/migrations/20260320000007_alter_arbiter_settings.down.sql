begin;

alter table arbiter_settings
    drop column if exists default_scope_id,
    drop column if exists decryptor_type_id,
    drop column if exists use_blacklist,
    drop column if exists use_whitelist;

alter table arbiter_settings
    add column if not exists reaction_security   text not null default 'null',
    add column if not exists invention_security  text not null default 'high',
    add column if not exists component_security  text not null default 'null',
    add column if not exists final_security      text not null default 'null';

commit;
