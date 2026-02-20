alter table auto_sell_containers
	add column price_source varchar(20) not null default 'jita_buy';
