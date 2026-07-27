CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    name           VARCHAR(255) NOT NULL,
    email          VARCHAR(255) NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    role           VARCHAR(20)  NOT NULL CHECK (role IN ('TENANT_ADMIN', 'DOCTOR', 'DENTIST', 'RECEPTIONIST', 'FINANCE')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    council_type   VARCHAR(10),
    council_number VARCHAR(20),
    council_state  VARCHAR(2),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_users_tenant ON users (tenant_id);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_email ON users (email);

-- Email is unique per tenant, not globally (see internal/domain/models/user.go).
CREATE UNIQUE INDEX idx_users_tenant_email ON users (tenant_id, email) WHERE deleted_at IS NULL;
