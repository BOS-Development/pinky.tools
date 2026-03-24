-- Migration: revert_augmentation_decryptor_fix
-- Created: Mon Mar 23 07:33:51 PM PDT 2026

-- Revert the incorrect fix from 20260323191828.
-- The original SDE values were correct: Augmentation gives +9 runs and -2 ME.
update sde_decryptors
set me_modifier = run_modifier, run_modifier = me_modifier
where name like '%Augmentation%';
