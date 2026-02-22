create table character_skills (
	character_id	bigint		not null,
	user_id		bigint		not null,
	skill_id	bigint		not null,
	trained_level	int		not null default 0,
	active_level	int		not null default 0,
	skillpoints	bigint		not null default 0,
	updated_at	timestamptz	not null default now(),
	primary key (character_id, skill_id)
);

create index idx_character_skills_user on character_skills (user_id);
