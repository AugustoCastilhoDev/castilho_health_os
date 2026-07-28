-- Metadata only (see internal/domain/models/patient_document.go) — the
-- file bytes live in Cloudflare R2, never in this table. file_key is the R2
-- object key, not a URL: every read/write goes through a short-lived
-- presigned URL generated on demand.
CREATE TABLE patient_documents (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    patient_id     UUID NOT NULL REFERENCES patients (id),
    uploaded_by_id UUID NOT NULL REFERENCES users (id),
    file_key       VARCHAR(500) NOT NULL,
    file_name      VARCHAR(255) NOT NULL,
    file_size      BIGINT       NOT NULL,
    mime_type      VARCHAR(150) NOT NULL,
    description    TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_patient_documents_file_key ON patient_documents (file_key);
CREATE INDEX idx_patient_documents_tenant ON patient_documents (tenant_id);
CREATE INDEX idx_patient_documents_patient ON patient_documents (patient_id);
