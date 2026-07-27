CREATE TABLE appointments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL REFERENCES tenants (id),
    patient_id           UUID NOT NULL REFERENCES patients (id),
    professional_id      UUID NOT NULL REFERENCES users (id),
    scheduled_at         TIMESTAMPTZ NOT NULL,
    duration_min         INT         NOT NULL DEFAULT 30,
    status               VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED'
        CHECK (status IN ('SCHEDULED', 'CONFIRMED', 'CANCELLED', 'WAITING', 'IN_PROGRESS', 'COMPLETED', 'NO_SHOW')),
    confirmed_at         TIMESTAMPTZ,
    checked_in_at        TIMESTAMPTZ,
    started_at           TIMESTAMPTZ,
    completed_at         TIMESTAMPTZ,
    cancelled_at         TIMESTAMPTZ,
    no_show_at           TIMESTAMPTZ,
    cancellation_reason  TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX idx_appointments_tenant ON appointments (tenant_id);
CREATE INDEX idx_appointments_patient ON appointments (patient_id);
CREATE INDEX idx_appointments_professional ON appointments (professional_id);
CREATE INDEX idx_appointments_scheduled_at ON appointments (scheduled_at);
CREATE INDEX idx_appointments_status ON appointments (status);

-- Append-only audit trail for every status transition (see
-- internal/domain/models/appointment.go). changed_by_id is not a hard FK to
-- users because automated actors (e.g. a WhatsApp webhook) may write here
-- using a well-known non-user UUID.
CREATE TABLE appointment_status_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    appointment_id UUID NOT NULL REFERENCES appointments (id),
    from_status    VARCHAR(20),
    to_status      VARCHAR(20) NOT NULL
        CHECK (to_status IN ('SCHEDULED', 'CONFIRMED', 'CANCELLED', 'WAITING', 'IN_PROGRESS', 'COMPLETED', 'NO_SHOW')),
    changed_by_id  UUID NOT NULL,
    reason         TEXT,
    changed_at     TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_status_logs_tenant ON appointment_status_logs (tenant_id);
CREATE INDEX idx_status_logs_appointment ON appointment_status_logs (appointment_id);
CREATE INDEX idx_status_logs_changed_at ON appointment_status_logs (changed_at);
