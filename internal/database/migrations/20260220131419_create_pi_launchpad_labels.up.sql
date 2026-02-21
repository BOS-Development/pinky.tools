create table pi_launchpad_labels (
	user_id bigint not null references users(id),
	character_id bigint not null,
	planet_id bigint not null,
	pin_id bigint not null,
	label varchar(100) not null,
	primary key(user_id, character_id, planet_id, pin_id)
);
