ALTER TABLE purchase_transactions
ADD COLUMN contract_key VARCHAR(50);

-- Create index for lookups (not unique since multiple items share same key)
CREATE INDEX idx_purchase_transactions_contract_key ON purchase_transactions(contract_key) WHERE contract_key IS NOT NULL;
