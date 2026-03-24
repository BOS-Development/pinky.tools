-- Migration: add_meta_group_id_to_asset_item_types
-- Adds meta_group_id to asset_item_types so item tier (T1/T2/Faction/etc.) can be
-- queried directly. T2 items have meta_group_id = 2. Used by the Arbiter advisor
-- to scope its blueprint scan to T2 ships and modules.

alter table asset_item_types
    add column if not exists meta_group_id bigint references sde_meta_groups(meta_group_id);

create index idx_asset_item_types_meta_group_id on asset_item_types(meta_group_id);
