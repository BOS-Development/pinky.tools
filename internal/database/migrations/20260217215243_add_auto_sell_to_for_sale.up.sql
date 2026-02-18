alter table for_sale_items add column auto_sell_container_id bigint references auto_sell_containers(id);
create index idx_for_sale_auto_sell on for_sale_items(auto_sell_container_id) where auto_sell_container_id is not null;
