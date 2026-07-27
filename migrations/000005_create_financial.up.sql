CREATE TABLE financial_rules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants (id),
    professional_id     UUID NOT NULL REFERENCES users (id),
    type                VARCHAR(30) NOT NULL
        CHECK (type IN ('PERCENTAGE', 'FIXED_PER_APPOINTMENT', 'FIXED_PER_PROCEDURE')),
    percentage          NUMERIC(5, 4),
    fixed_amount_cents  BIGINT,
    procedure_code      VARCHAR(50),
    insurance_plan      VARCHAR(100),
    fee_deduction       VARCHAR(20) NOT NULL DEFAULT 'BEFORE_SPLIT'
        CHECK (fee_deduction IN ('BEFORE_SPLIT', 'AFTER_SPLIT')),
    priority            INT         NOT NULL DEFAULT 0,
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_financial_rules_tenant ON financial_rules (tenant_id);
CREATE INDEX idx_financial_rules_professional ON financial_rules (professional_id);
CREATE INDEX idx_financial_rules_procedure_code ON financial_rules (procedure_code);
CREATE INDEX idx_financial_rules_insurance_plan ON financial_rules (insurance_plan);

-- Ledger entries: patient income and professional payouts share this table
-- (see internal/domain/models/financial.go). A payout row references the
-- income row that funded it via source_transaction_id.
CREATE TABLE financial_transactions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              UUID NOT NULL REFERENCES tenants (id),
    appointment_id         UUID REFERENCES appointments (id),
    patient_id             UUID REFERENCES patients (id),
    professional_id        UUID REFERENCES users (id),
    source_transaction_id  UUID REFERENCES financial_transactions (id),
    financial_rule_id      UUID REFERENCES financial_rules (id),
    type                   VARCHAR(30) NOT NULL
        CHECK (type IN ('PATIENT_PAYMENT', 'PROFESSIONAL_PAYOUT')),
    status                 VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PAID', 'CANCELLED')),
    gross_amount_cents     BIGINT      NOT NULL,
    fee_amount_cents       BIGINT      NOT NULL DEFAULT 0,
    net_amount_cents       BIGINT      NOT NULL,
    payment_method         VARCHAR(20)
        CHECK (payment_method IN ('CASH', 'DEBIT_CARD', 'CREDIT_CARD', 'PIX', 'BANK_TRANSFER', 'INSURANCE')),
    due_date               TIMESTAMPTZ,
    paid_at                TIMESTAMPTZ,
    notes                  TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX idx_financial_tx_tenant ON financial_transactions (tenant_id);
CREATE INDEX idx_financial_tx_appointment ON financial_transactions (appointment_id);
CREATE INDEX idx_financial_tx_patient ON financial_transactions (patient_id);
CREATE INDEX idx_financial_tx_professional ON financial_transactions (professional_id);
CREATE INDEX idx_financial_tx_source ON financial_transactions (source_transaction_id);
CREATE INDEX idx_financial_tx_type ON financial_transactions (type);
CREATE INDEX idx_financial_tx_status ON financial_transactions (status);
