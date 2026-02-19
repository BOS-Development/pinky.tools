create table contact_rules (
	id bigserial primary key,
	user_id bigint not null references users(id),
	rule_type varchar(20) not null,
	entity_id bigint,
	entity_name varchar(500),
	is_active boolean not null default true,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now(),
	constraint contact_rules_valid_type check (
		rule_type in ('corporation', 'alliance', 'everyone')
	),
	constraint contact_rules_entity_check check (
		(rule_type = 'everyone' and entity_id is null) or
		(rule_type != 'everyone' and entity_id is not null)
	)
);

create unique index idx_contact_rules_unique
	on contact_rules(user_id, rule_type, coalesce(entity_id, 0))
	where is_active = true;

create index idx_contact_rules_entity
	on contact_rules(rule_type, entity_id)
	where is_active = true;
