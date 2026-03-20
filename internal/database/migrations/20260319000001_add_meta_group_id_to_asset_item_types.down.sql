drop index if exists idx_asset_item_types_meta_group_id;

alter table asset_item_types
    drop column if exists meta_group_id;
