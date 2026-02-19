alter table contact_rules add column permissions jsonb not null default '["for_sale_browse"]';
