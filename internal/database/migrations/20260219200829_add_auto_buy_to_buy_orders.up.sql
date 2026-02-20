alter table buy_orders add column auto_buy_config_id bigint references auto_buy_configs(id);

create unique index idx_buy_orders_auto_buy_unique on buy_orders(
	buyer_user_id, type_id, location_id, auto_buy_config_id
) where auto_buy_config_id is not null and is_active = true;
