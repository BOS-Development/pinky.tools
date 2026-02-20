drop index if exists idx_buy_orders_auto_buy_unique;
alter table buy_orders drop column if exists auto_buy_config_id;
