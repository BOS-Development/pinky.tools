-- Migration: create_arbiter_scopes
-- Named asset scopes per user for Arbiter. A scope groups one or more characters
-- and/or corporations whose assets are pooled together when evaluating inventory
-- coverage and build decisions.
--
-- arbiter_scope_members links characters/corporations to a scope.
-- arbiter_settings.default_scope_id references this table (added in a later migration).

begin;

create table arbiter_scopes (
    id          bigserial       primary key,
    user_id     bigint          not null references users(id) on delete cascade,
    name        text            not null,
    is_default  boolean         not null default false,
    created_at  timestamptz     not null default now()
);

-- Each user's scope names must be unique.
create unique index idx_arbiter_scopes_user_name
    on arbiter_scopes(user_id, name);

-- Fast lookup of all scopes for a user.
create index idx_arbiter_scopes_user_id
    on arbiter_scopes(user_id);

commit;
