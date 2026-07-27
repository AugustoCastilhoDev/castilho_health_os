CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('MEDICA', 'ODONTO', 'MISTA')),
    document   VARCHAR(20)  NOT NULL,
    email      VARCHAR(255) NOT NULL,
    phone      VARCHAR(20),
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    settings   JSONB,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Partial unique indexes so a soft-deleted tenant doesn't block reuse of its slug/document.
CREATE UNIQUE INDEX idx_tenants_slug ON tenants (slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_tenants_document ON tenants (document) WHERE deleted_at IS NULL;
