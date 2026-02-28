-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE accounts
(
    id            uuid PRIMARY KEY      DEFAULT gen_random_uuid(),

    email         varchar(254) NOT NULL,
    password_hash text         NOT NULL,

    first_name    varchar(200) NOT NULL,
    last_name     varchar(200) NOT NULL,

    status        text         NOT NULL DEFAULT 'active',

    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NULL,

    CONSTRAINT chk_accounts_email_not_blank
        CHECK (btrim(email) <> ''),

    CONSTRAINT chk_accounts_email_trimmed
        CHECK (email = btrim(email)),

    CONSTRAINT chk_accounts_email_lower
        CHECK (email = lower(email)),

    CONSTRAINT chk_accounts_password_hash_not_blank
        CHECK (btrim(password_hash) <> ''),

    CONSTRAINT chk_accounts_first_name_not_blank
        CHECK (btrim(first_name) <> ''),

    CONSTRAINT chk_accounts_last_name_not_blank
        CHECK (btrim(last_name) <> ''),

    CONSTRAINT chk_accounts_status
        CHECK (status IN ('active', 'suspended', 'blocked')),

    CONSTRAINT chk_accounts_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE UNIQUE INDEX ux_accounts_email ON accounts (email);
CREATE INDEX ix_accounts_created_at ON accounts (created_at);
CREATE INDEX ix_accounts_status ON accounts (status);

-- +goose Down
DROP TABLE IF EXISTS accounts;