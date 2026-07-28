-- Prontuário entries (see internal/domain/models/medical_record.go). Content
-- is opaque text/JSON produced by the frontend editor; this table only owns
-- identity, classification and the lock state.
CREATE TABLE medical_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants (id),
    patient_id      UUID NOT NULL REFERENCES patients (id),
    professional_id UUID NOT NULL REFERENCES users (id),
    appointment_id  UUID REFERENCES appointments (id),
    type            VARCHAR(20) NOT NULL
        CHECK (type IN ('MEDICA', 'ODONTOLOGICA', 'PSICOLOGICA', 'PSIQUIATRICA')),
    content         TEXT        NOT NULL,
    is_locked       BOOLEAN     NOT NULL DEFAULT FALSE,
    locked_at       TIMESTAMPTZ,
    locked_by_id    UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_medical_records_tenant ON medical_records (tenant_id);
CREATE INDEX idx_medical_records_patient ON medical_records (patient_id);
CREATE INDEX idx_medical_records_professional ON medical_records (professional_id);
CREATE INDEX idx_medical_records_appointment ON medical_records (appointment_id);
CREATE INDEX idx_medical_records_type ON medical_records (type);
CREATE INDEX idx_medical_records_is_locked ON medical_records (is_locked);

-- Clinic-level reusable document bodies (see
-- internal/domain/models/document_template.go) — Content carries {{tag}}
-- placeholders resolved at PDF-generation time, never at rest.
CREATE TABLE document_templates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants (id),
    name       VARCHAR(255) NOT NULL,
    type       VARCHAR(20)  NOT NULL
        CHECK (type IN ('ATESTADO', 'DECLARACAO', 'LAUDO', 'OUTRO')),
    content    TEXT         NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_document_templates_tenant ON document_templates (tenant_id);
CREATE INDEX idx_document_templates_type ON document_templates (type);
