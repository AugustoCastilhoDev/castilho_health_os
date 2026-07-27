-- Lets a PATIENT_PAYMENT carry the same scope dimensions FinancialRule
-- matches on (procedure/insurance), so automatic settlement can resolve a
-- non-wildcard rule instead of always falling back to the professional's
-- default.
ALTER TABLE financial_transactions ADD COLUMN procedure_code VARCHAR(50);
ALTER TABLE financial_transactions ADD COLUMN insurance_plan VARCHAR(100);

CREATE INDEX idx_financial_tx_procedure_code ON financial_transactions (procedure_code);
CREATE INDEX idx_financial_tx_insurance_plan ON financial_transactions (insurance_plan);
