create table auto_sell_containers (
	id bigserial primary key,
	user_id bigint not null references users(id),
	owner_type varchar(20) not null,
	owner_id bigint not null,
	location_id bigint not null,
	container_id bigint not null,
	division_number int,
	price_percentage numeric(5, 2) not null default 90.00,
	is_active boolean not null default true,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now(),
	constraint auto_sell_valid_percentage check (price_percentage > 0 and price_percentage <= 200)
);

create unique index idx_auto_sell_unique_container on auto_sell_containers(
	user_id, owner_type, owner_id, location_id, container_id,
	coalesce(division_number, 0)
) where is_active = true;

create index idx_auto_sell_user on auto_sell_containers(user_id);
