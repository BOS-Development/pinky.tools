begin;

create table job_slot_job_notifications (
    id             bigserial primary key,
    character_id   bigint not null,
    esi_job_id     bigint not null,
    notified_at    timestamptz not null default now(),
    unique (character_id, esi_job_id)
);

create index idx_job_slot_job_notifications_character on job_slot_job_notifications(character_id);

commit;
