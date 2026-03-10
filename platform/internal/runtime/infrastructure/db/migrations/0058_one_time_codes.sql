-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'auth') THEN
            RAISE EXCEPTION 'schema "auth" does not exist';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run iam/accounts migration first';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relname = 'one_time_codes'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "auth.one_time_codes" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.one_time_codes
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    purpose        varchar(64) NOT NULL,
    code_hash      text        NOT NULL,
    account_id     uuid        NOT NULL,
    provider       varchar(64) NOT NULL,
    intent         varchar(32) NOT NULL,
    is_new_account boolean     NOT NULL DEFAULT false,
    expires_at     timestamptz NOT NULL,
    used_at        timestamptz NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NULL,
    CONSTRAINT fk_auth_one_time_codes_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_auth_one_time_codes_code_hash
        UNIQUE (code_hash),
    CONSTRAINT chk_auth_one_time_codes_purpose_not_blank
        CHECK (btrim(purpose) <> ''),
    CONSTRAINT chk_auth_one_time_codes_code_hash_not_blank
        CHECK (btrim(code_hash) <> ''),
    CONSTRAINT chk_auth_one_time_codes_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_auth_one_time_codes_intent_valid
        CHECK (intent IN ('login', 'signup')),
    CONSTRAINT chk_auth_one_time_codes_expiry_valid
        CHECK (expires_at > created_at),
    CONSTRAINT chk_auth_one_time_codes_used_valid
        CHECK (used_at IS NULL OR used_at >= created_at),
    CONSTRAINT chk_auth_one_time_codes_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX ix_auth_one_time_codes_account_id
    ON auth.one_time_codes (account_id);

CREATE INDEX ix_auth_one_time_codes_purpose
    ON auth.one_time_codes (purpose);

CREATE INDEX ix_auth_one_time_codes_expires_at
    ON auth.one_time_codes (expires_at);

-- +goose Down

DROP TABLE auth.one_time_codes;
