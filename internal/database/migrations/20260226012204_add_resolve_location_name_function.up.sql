create or replace function resolve_location_name(p_location_id bigint)
returns text
language sql
stable
as $$
    select coalesce(
        (select name from stations where station_id = p_location_id limit 1),
        (select name from solar_systems where solar_system_id = p_location_id limit 1),
        'Unknown Location'
    )
$$;
