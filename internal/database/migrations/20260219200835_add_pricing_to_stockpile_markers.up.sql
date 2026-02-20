alter table stockpile_markers
	add column price_source varchar(20),
	add column price_percentage numeric(5, 2);
