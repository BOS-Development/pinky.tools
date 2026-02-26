create or replace function resolve_owner_name(p_owner_type text, p_owner_id bigint)
returns text
language sql
stable
as $$
    select coalesce(
        case when p_owner_type = 'character' then
            (select name from characters where id = p_owner_id limit 1)
        when p_owner_type = 'corporation' then
            (select name from player_corporations where id = p_owner_id limit 1)
        end,
        'Unknown'
    )
$$;
