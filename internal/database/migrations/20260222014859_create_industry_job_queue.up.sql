create table industry_job_queue (
	id			bigserial	primary key,
	user_id			bigint		not null,
	character_id		bigint,
	blueprint_type_id	bigint		not null,
	activity		text		not null,
	runs			int		not null,
	me_level		int		not null default 0,
	te_level		int		not null default 0,
	system_id		bigint,
	facility_tax		float8		not null default 0,
	status			text		not null default 'planned',
	esi_job_id		bigint,
	product_type_id		bigint,
	estimated_cost		float8,
	estimated_duration	int,
	notes			text,
	created_at		timestamptz	not null default now(),
	updated_at		timestamptz	not null default now()
);

create index idx_job_queue_user on industry_job_queue (user_id);
create index idx_job_queue_status on industry_job_queue (status);
