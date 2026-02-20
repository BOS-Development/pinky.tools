create table discord_links (
	id bigserial primary key,
	user_id bigint not null references users(id) unique,
	discord_user_id varchar(50) not null,
	discord_username varchar(100) not null,
	access_token text not null,
	refresh_token text not null,
	token_expires_at timestamp not null,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now()
);

create table discord_notification_targets (
	id bigserial primary key,
	user_id bigint not null references users(id),
	target_type varchar(20) not null,
	channel_id varchar(50),
	guild_name varchar(100) not null default '',
	channel_name varchar(100) not null default '',
	is_active boolean not null default true,
	created_at timestamp not null default now(),
	updated_at timestamp not null default now()
);

create index idx_discord_targets_user on discord_notification_targets(user_id);

create table notification_preferences (
	id bigserial primary key,
	target_id bigint not null references discord_notification_targets(id) on delete cascade,
	event_type varchar(50) not null,
	is_enabled boolean not null default true,
	unique(target_id, event_type)
);
