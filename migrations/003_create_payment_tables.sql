CREATE TABLE invoice_payments (
    id         TEXT        PRIMARY KEY,
    account_id TEXT        NOT NULL REFERENCES accounts(id),
    barcode    TEXT        NOT NULL,
    amount     BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE transfers (
    id         TEXT        PRIMARY KEY,
    amount     BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE entries (
    id          TEXT        PRIMARY KEY,
    account_id  TEXT        NOT NULL REFERENCES accounts(id),
    type        TEXT        NOT NULL,  -- DEBIT or CREDIT
    amount      BIGINT      NOT NULL,
    source_type TEXT        NOT NULL,  -- TRANSFER or INVOICE_PAYMENT
    source_id   TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE payment_outbox (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_name   TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    occurred_at  TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);

CREATE INDEX payment_outbox_pending_idx ON payment_outbox (occurred_at) WHERE processed_at IS NULL;
