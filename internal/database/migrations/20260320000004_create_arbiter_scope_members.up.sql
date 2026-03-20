-- Migration: create_arbiter_scope_members
-- Associates characters and corporations with an arbiter scope.
-- member_type: 'character' or 'corporation'
-- member_id: character.id or player_corporations.id — no FK because
--   characters has a composite PK (id, user_id), making a bare REFERENCES characters(id)
--   invalid. Application layer enforces membership belongs to the scope's user.

begin;

create table arbiter_scope_members (
    id          bigserial   primary key,
    scope_id    bigint      not null references arbiter_scopes(id) on delete cascade,
    member_type text        not null,
    member_id   bigint      not null
);

-- A given character/corporation can only appear once per scope.
create unique index idx_arbiter_scope_members_unique
    on arbiter_scope_members(scope_id, member_type, member_id);

-- Fast lookup of all members for a scope.
create index idx_arbiter_scope_members_scope_id
    on arbiter_scope_members(scope_id);

commit;
