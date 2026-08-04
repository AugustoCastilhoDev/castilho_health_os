-- Odontograma (see internal/domain/models/odontograma.go). Append-only
-- ledger of tooth findings/procedures — the chart shown to the dentist is
-- always the most recent entry per tooth, derived at read time with
-- DISTINCT ON (OdontogramaRepository.CurrentChart), not a denormalized
-- current-state table: unlike stock_items.quantity_on_hand there's no
-- numeric balance to keep in sync, a condition change is a plain
-- replacement.
CREATE TABLE odontograma_entries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    patient_id     UUID NOT NULL REFERENCES patients (id),
    -- FDI two-digit notation (quadrants 1-4 permanent, 5-8 deciduous). The
    -- exact whitelist (models.IsValidToothNumber) is enforced in the
    -- service layer; this CHECK only rejects gross shape errors.
    tooth_number   VARCHAR(2)  NOT NULL CHECK (tooth_number ~ '^[1-8][1-8]$'),
    condition      VARCHAR(20) NOT NULL CHECK (condition IN (
        'SAUDAVEL', 'CARIE', 'RESTAURADO', 'AUSENTE', 'CANAL', 'COROA',
        'IMPLANTE', 'FRATURADO', 'A_EXTRAIR'
    )),
    note           TEXT,
    recorded_by_id UUID        NOT NULL REFERENCES users (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_odontograma_entries_tenant ON odontograma_entries (tenant_id);
CREATE INDEX idx_odontograma_entries_patient ON odontograma_entries (patient_id, tooth_number);
