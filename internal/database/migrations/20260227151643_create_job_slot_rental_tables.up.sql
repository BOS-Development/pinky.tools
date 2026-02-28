-- Migration: create_job_slot_rental_tables
-- Created: Thu Feb 27 03:16:43 PM UTC 2026

-- Job Slot Rental Listings
-- Users list their idle industry job slots for rent
create table job_slot_rental_listings (
	id bigserial primary key,
	user_id bigint not null references users(id),
	character_id bigint not null,
	activity_type text not null,
	slots_listed int not null,
	price_amount double precision not null,
	pricing_unit text not null,
	location_id bigint,
	notes text,
	is_active boolean not null default true,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint job_slot_listings_positive_slots check (slots_listed > 0),
	constraint job_slot_listings_positive_price check (price_amount >= 0),
	constraint job_slot_listings_valid_activity check (
		activity_type in ('manufacturing', 'reaction', 'copying', 'invention', 'me_research', 'te_research')
	),
	constraint job_slot_listings_valid_pricing check (
		pricing_unit in ('per_slot_day', 'per_job', 'flat_fee')
	)
);

-- Prevent duplicate active listings for same user/character/activity combo
create unique index idx_job_slot_listings_unique_active on job_slot_rental_listings(
	user_id, character_id, activity_type
) where is_active = true;

create index idx_job_slot_listings_user on job_slot_rental_listings(user_id);
create index idx_job_slot_listings_character on job_slot_rental_listings(character_id);
create index idx_job_slot_listings_activity on job_slot_rental_listings(activity_type) where is_active = true;

-- Job Slot Interest Requests
-- Buyers express interest in renting slots from a listing
create table job_slot_interest_requests (
	id bigserial primary key,
	listing_id bigint not null references job_slot_rental_listings(id) on delete cascade,
	requester_user_id bigint not null references users(id),
	slots_requested int not null,
	duration_days int,
	message text,
	status text not null default 'pending',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint job_slot_requests_positive_slots check (slots_requested > 0),
	constraint job_slot_requests_valid_status check (
		status in ('pending', 'accepted', 'declined', 'withdrawn')
	)
);

create index idx_job_slot_requests_listing on job_slot_interest_requests(listing_id);
create index idx_job_slot_requests_requester on job_slot_interest_requests(requester_user_id);
create index idx_job_slot_requests_status on job_slot_interest_requests(status);
