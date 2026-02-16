DROP INDEX IF EXISTS idx_purchase_transactions_contract_key;

ALTER TABLE purchase_transactions
DROP COLUMN contract_key;
