CREATE TABLE patients (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    name           VARCHAR(255) NOT NULL,
    document       VARCHAR(20),
    birth_date     DATE,
    phone          VARCHAR(20),
    email          VARCHAR(255),
    address_zip    VARCHAR(10),
    address_street VARCHAR(255),
    address_city   VARCHAR(100),
    address_state  VARCHAR(2),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_patients_tenant ON patients (tenant_id);
CREATE INDEX idx_patients_name ON patients (name);
CREATE INDEX idx_patients_document ON patients (document);
