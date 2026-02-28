-- Migration: create_job_slot_rental_tables
-- Created: Thu Feb 27 03:16:43 PM UTC 2026

-- Drop tables in reverse order (child before parent due to FK)
drop table if exists job_slot_interest_requests;
drop table if exists job_slot_rental_listings;
