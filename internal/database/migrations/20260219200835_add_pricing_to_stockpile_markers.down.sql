alter table stockpile_markers
	drop column if exists price_source,
	drop column if exists price_percentage;
