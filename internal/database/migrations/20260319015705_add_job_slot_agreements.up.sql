-- Migration: add_job_slot_agreements
-- Created: Thu Mar 19 01:57:05 AM PDT 2026

begin;

create table job_slot_agreements (
    id                   bigserial primary key,
    interest_request_id  bigint not null references job_slot_interest_requests(id) on delete restrict,
    listing_id           bigint not null references job_slot_rental_listings(id) on delete restrict,
    seller_user_id       bigint not null references users(id) on delete restrict,
    renter_user_id       bigint not null references users(id) on delete restrict,
    slots_agreed         int not null,
    price_amount         numeric(12,2) not null,
    pricing_unit         text not null,
    agreed_at            timestamptz not null default now(),
    expected_end_at      timestamptz,
    status               text not null default 'active',
    cancellation_reason  text,
    created_at           timestamptz not null default now(),
    updated_at           timestamptz not null default now(),
    constraint job_slot_agreements_positive_slots check (slots_agreed > 0),
    constraint job_slot_agreements_positive_price check (price_amount >= 0),
    constraint job_slot_agreements_valid_pricing_unit check (
        pricing_unit in ('per_slot_day', 'per_job', 'flat_fee')
    ),
    constraint job_slot_agreements_valid_status check (
        status in ('active', 'completed', 'cancelled')
    )
);

create index idx_job_slot_agreements_interest_request on job_slot_agreements(interest_request_id);
create index idx_job_slot_agreements_listing on job_slot_agreements(listing_id);
create index idx_job_slot_agreements_seller on job_slot_agreements(seller_user_id);
create index idx_job_slot_agreements_renter on job_slot_agreements(renter_user_id);
create index idx_job_slot_agreements_seller_status on job_slot_agreements(seller_user_id, status);
create index idx_job_slot_agreements_renter_status on job_slot_agreements(renter_user_id, status);
create unique index idx_job_slot_agreements_unique_active_request
    on job_slot_agreements(interest_request_id)
    where status = 'active';

commit;
