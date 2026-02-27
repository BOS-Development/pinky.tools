-- Migration: widen_contract_key_to_text
-- Created: Fri Feb 27 10:26:30 AM PST 2026

alter table purchase_transactions alter column contract_key type varchar(50);
