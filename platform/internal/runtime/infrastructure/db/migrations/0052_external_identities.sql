-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run accounts migration first';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'auth') THEN
            RAISE EXCEPTION 'schema "auth" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relname = 'external_identities'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "auth.external_identities" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.external_identities
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider         varchar(64) NOT NULL,
    external_subject text        NOT NULL,
    account_id       uuid        NOT NULL,
    email            varchar(320) NULL,
    email_verified   boolean     NOT NULL DEFAULT false,
    display_name     varchar(255) NULL,
    claims_json      jsonb       NOT NULL DEFAULT '{}'::jsonb,
    last_login_at    timestamptz NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NULL,
    CONSTRAINT fk_auth_external_identities_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_auth_external_identities_provider_subject
        UNIQUE (provider, external_subject),
    CONSTRAINT chk_auth_external_identities_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_auth_external_identities_subject_not_blank
        CHECK (btrim(external_subject) <> ''),
    CONSTRAINT chk_auth_external_identities_email_not_blank
        CHECK (email IS NULL OR btrim(email) <> ''),
    CONSTRAINT chk_auth_external_identities_last_login_valid
        CHECK (last_login_at IS NULL OR last_login_at >= created_at),
    CONSTRAINT chk_auth_external_identities_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX ix_auth_external_identities_account_id
    ON auth.external_identities (account_id);

CREATE INDEX ix_auth_external_identities_email
    ON auth.external_identities (email);

-- +goose Down

DROP TABLE auth.external_identities;
