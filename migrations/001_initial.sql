CREATE TABLE clients (
    id          TEXT PRIMARY KEY,
    document    TEXT NOT NULL UNIQUE,
    full_name   TEXT NOT NULL,
    email       TEXT NOT NULL UNIQUE,
    phone       TEXT NOT NULL,
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE client_documents (
    client_id  TEXT        PRIMARY KEY REFERENCES clients(id),
    number     TEXT        NOT NULL,
    type       TEXT        NOT NULL  -- 'CPF' ou 'CNPJ'
);

CREATE TABLE client_addresses (
    client_id    TEXT    PRIMARY KEY REFERENCES clients(id),
    zip_code     TEXT    NOT NULL,
    street       TEXT    NOT NULL,
    number       TEXT    NOT NULL,
    complement   TEXT,
    neighborhood TEXT    NOT NULL,
    city         TEXT    NOT NULL,
    state        CHAR(2) NOT NULL
);

CREATE TABLE outbox (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_name   TEXT NOT NULL,
    payload      JSONB NOT NULL,
    occurred_at  TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);

-- partial index: só indexa registros pendentes, mantendo a query do worker eficiente
CREATE INDEX outbox_pending_idx ON outbox (occurred_at) WHERE processed_at IS NULL;
