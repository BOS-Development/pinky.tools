create table auto_buy_configs (
	id bigserial primary key,
	user_id bigint not null references users(id),
	owner_type varchar(20) not null,
	owner_id bigint not null,
	location_id bigint not null,
	container_id bigint,
	division_number int,
	price_percentage numeric(5, 2) not null default 100.00,
	price_source varchar(20) not null default 'jita_sell',
	is_active boolean not null default true,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now()
);

create unique index idx_auto_buy_unique_active on auto_buy_configs(
	user_id, owner_type, owner_id, location_id,
	coalesce(container_id, 0), coalesce(division_number, 0)
) where is_active = true;

create index idx_auto_buy_user on auto_buy_configs(user_id);
