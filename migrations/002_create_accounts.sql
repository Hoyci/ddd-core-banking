CREATE SEQUENCE account_number_seq START 1;

CREATE TABLE accounts (
    id         TEXT        PRIMARY KEY,
    client_id  TEXT        NOT NULL UNIQUE REFERENCES clients(id),
    number     TEXT        NOT NULL UNIQUE,
    blocked    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);
