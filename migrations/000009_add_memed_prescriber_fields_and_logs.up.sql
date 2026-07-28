-- Fields the Memed prescriber-registration API requires that we never
-- needed to collect before (see internal/domain/models/user.go): cpf and
-- birth_date are mandatory on their side, phone/sex are optional. Nullable
-- because they only apply to DOCTOR/DENTIST users actually issuing
-- prescriptions through Memed, not every user.
ALTER TABLE users ADD COLUMN cpf VARCHAR(11);
ALTER TABLE users ADD COLUMN birth_date DATE;
ALTER TABLE users ADD COLUMN sex VARCHAR(1) CHECK (sex IN ('M', 'F'));
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- Audit trail only (see internal/domain/models/memed_prescription_log.go) —
-- Memed itself is the system of record for the prescription's content; this
-- table just answers "quem prescreveu o quê e quando" for the clinic.
CREATE TABLE memed_prescription_logs (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             UUID NOT NULL REFERENCES tenants (id),
    patient_id            UUID NOT NULL REFERENCES patients (id),
    professional_id       UUID NOT NULL REFERENCES users (id),
    memed_prescription_id VARCHAR(100) NOT NULL,
    status                VARCHAR(20)  NOT NULL DEFAULT 'ISSUED' CHECK (status IN ('ISSUED', 'CANCELLED')),
    issued_at             TIMESTAMPTZ  NOT NULL,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_memed_prescription_logs_memed_id ON memed_prescription_logs (memed_prescription_id);
CREATE INDEX idx_memed_prescription_logs_tenant ON memed_prescription_logs (tenant_id);
CREATE INDEX idx_memed_prescription_logs_patient ON memed_prescription_logs (patient_id);
CREATE INDEX idx_memed_prescription_logs_status ON memed_prescription_logs (status);
