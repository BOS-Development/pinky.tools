-- Migration: create_sde_decryptors
-- Denormalized table of T2 invention decryptors derived from sde_type_dogma_attributes.
-- Populated by the SDE importer using dogma attribute IDs:
--   1112 = inventionPropabilityMultiplier  (float multiplier, e.g. 1.2 = +20% chance)
--   1113 = inventionMEModifier             (integer offset applied to BPC ME, e.g. +2)
--   1114 = inventionTEModifier             (integer offset applied to BPC TE, e.g. +10)
--   1124 = inventionMaxRunModifier         (integer offset applied to BPC run count, e.g. +1)
--
-- The "no decryptor" option is NOT stored as a row here. It is handled in application
-- code by using probability_multiplier=1.0, me_modifier=0, te_modifier=0, run_modifier=0.
-- This avoids polluting FK references with a sentinel null row.
--
-- The standard T2 decryptors are type_ids 34201-34208 (Accelerant, Attainment,
-- Augmentation, Optimized Attainment, Optimized Augmentation, Parity, Process, Symmetry).
-- The SDE importer should populate this table for all types that carry attribute 1112.

begin;

create table sde_decryptors (
    type_id                 bigint              primary key references asset_item_types(type_id),
    name                    text                not null,
    probability_multiplier  double precision    not null,
    me_modifier             int                 not null,
    te_modifier             int                 not null,
    run_modifier            int                 not null
);

commit;
