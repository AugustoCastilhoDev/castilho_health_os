-- Estoque (see internal/domain/models/stock.go). StockItem holds the running
-- QuantityOnHand; StockMovement is the append-only ledger that explains how
-- it got there — the same "ledger + derived balance" shape already used by
-- financial_transactions, kept in sync inside one DB transaction per
-- movement (StockMovementRepository.RecordMovement).
CREATE TABLE stock_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants (id),
    name              VARCHAR(255) NOT NULL,
    unit              VARCHAR(20)  NOT NULL,
    quantity_on_hand  INTEGER      NOT NULL DEFAULT 0,
    min_quantity      INTEGER,
    is_active         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_stock_items_tenant ON stock_items (tenant_id);

CREATE TABLE stock_movements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants (id),
    item_id        UUID NOT NULL REFERENCES stock_items (id),
    type           VARCHAR(10) NOT NULL CHECK (type IN ('IN', 'OUT')),
    quantity       INTEGER     NOT NULL CHECK (quantity > 0),
    note           TEXT,
    created_by_id  UUID        NOT NULL REFERENCES users (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_stock_movements_tenant ON stock_movements (tenant_id);
CREATE INDEX idx_stock_movements_item ON stock_movements (item_id);
