create table pi_planets (
	id bigserial primary key,
	character_id bigint not null,
	user_id bigint not null,
	planet_id bigint not null,
	planet_type varchar(20) not null,
	solar_system_id bigint not null,
	upgrade_level int not null default 0,
	num_pins int not null default 0,
	last_update timestamp not null,
	last_stall_notified_at timestamp,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now(),
	unique(character_id, planet_id)
);

create index idx_pi_planets_user on pi_planets(user_id);

create table pi_pins (
	id bigserial primary key,
	character_id bigint not null,
	planet_id bigint not null,
	pin_id bigint not null,
	type_id bigint not null,
	schematic_id int,
	latitude double precision,
	longitude double precision,
	install_time timestamp,
	expiry_time timestamp,
	last_cycle_start timestamp,
	extractor_cycle_time int,
	extractor_head_radius double precision,
	extractor_product_type_id bigint,
	extractor_qty_per_cycle int,
	extractor_num_heads int,
	pin_category varchar(20) not null,
	updated_at timestamp not null default now(),
	unique(character_id, planet_id, pin_id)
);

create index idx_pi_pins_planet on pi_pins(character_id, planet_id);

create table pi_pin_contents (
	character_id bigint not null,
	planet_id bigint not null,
	pin_id bigint not null,
	type_id bigint not null,
	amount bigint not null,
	primary key(character_id, planet_id, pin_id, type_id)
);

create table pi_routes (
	character_id bigint not null,
	planet_id bigint not null,
	route_id bigint not null,
	source_pin_id bigint not null,
	destination_pin_id bigint not null,
	content_type_id bigint not null,
	quantity bigint not null,
	primary key(character_id, planet_id, route_id)
);

create table pi_tax_config (
	id bigserial primary key,
	user_id bigint not null references users(id),
	planet_id bigint,
	tax_rate double precision not null default 10.0,
	unique(user_id, planet_id)
);
