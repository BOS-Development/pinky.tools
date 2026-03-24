-- Migration: create_arbiter_settings
-- Per-user configuration for the Arbiter T2 profit advisor.
-- Stores 4 structure slots: reaction, invention, component build, final build.
-- Each slot records structure type, rig tier, and security class for applying
-- the correct industry bonuses. system_id drives cost index lookup from
-- industry_cost_indices (activity = 'manufacturing' or 'reaction' etc.).

begin;

create table arbiter_settings (
    user_id             bigint      primary key references users(id) on delete cascade,

    -- Reaction structure (composite reactions, moon goo reactions)
    reaction_structure  text        not null default 'athanor',
    reaction_rig        text        not null default 't1',
    reaction_security   text        not null default 'null',
    reaction_system_id  bigint      references solar_systems(solar_system_id) on delete set null,

    -- Invention structure (T2 BPC invention)
    invention_structure text        not null default 'raitaru',
    invention_rig       text        not null default 't1',
    invention_security  text        not null default 'high',
    invention_system_id bigint      references solar_systems(solar_system_id) on delete set null,

    -- Component build structure (T2 components)
    component_structure text        not null default 'raitaru',
    component_rig       text        not null default 't2',
    component_security  text        not null default 'null',
    component_system_id bigint      references solar_systems(solar_system_id) on delete set null,

    -- Final build structure (T2 ships and modules)
    final_structure     text        not null default 'raitaru',
    final_rig           text        not null default 't2',
    final_security      text        not null default 'null',
    final_system_id     bigint      references solar_systems(solar_system_id) on delete set null,

    updated_at          timestamptz not null default now()
);

commit;
