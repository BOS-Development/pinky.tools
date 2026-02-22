create table esi_industry_jobs (
	job_id			bigint		primary key,
	installer_id		bigint		not null,
	user_id			bigint		not null,
	facility_id		bigint		not null,
	station_id		bigint		not null,
	activity_id		int		not null,
	blueprint_id		bigint		not null,
	blueprint_type_id	bigint		not null,
	blueprint_location_id	bigint		not null,
	output_location_id	bigint		not null,
	runs			int		not null,
	cost			float8,
	licensed_runs		int,
	probability		float8,
	product_type_id		bigint,
	status			text		not null,
	duration		int		not null,
	start_date		timestamptz	not null,
	end_date		timestamptz	not null,
	pause_date		timestamptz,
	completed_date		timestamptz,
	completed_character_id	bigint,
	successful_runs		int,
	solar_system_id		bigint,
	updated_at		timestamptz	not null default now()
);

create index idx_esi_jobs_user on esi_industry_jobs (user_id);
create index idx_esi_jobs_status on esi_industry_jobs (status);
