-- Migration: revert_augmentation_decryptor_fix
-- Created: Mon Mar 23 07:33:51 PM PDT 2026

-- Re-apply the (incorrect) swap so rolling back this migration undoes the revert.
update sde_decryptors
set me_modifier = run_modifier, run_modifier = me_modifier
where name like '%Augmentation%';
