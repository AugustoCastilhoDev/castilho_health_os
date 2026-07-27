DROP INDEX IF EXISTS idx_financial_tx_insurance_plan;
DROP INDEX IF EXISTS idx_financial_tx_procedure_code;

ALTER TABLE financial_transactions DROP COLUMN IF EXISTS insurance_plan;
ALTER TABLE financial_transactions DROP COLUMN IF EXISTS procedure_code;
