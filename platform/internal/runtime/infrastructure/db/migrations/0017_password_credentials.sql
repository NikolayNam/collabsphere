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
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.password_credentials
(
    account_id    uuid PRIMARY KEY,
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_auth_password_credentials_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,

    CONSTRAINT chk_auth_password_credentials_password_hash_not_blank
        CHECK (btrim(password_hash) <> ''),

    CONSTRAINT chk_auth_password_credentials_updated_at_valid
        CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX ux_auth_password_credentials_account_id
    ON auth.password_credentials (account_id);

-- +goose Down
DROP INDEX auth.ux_auth_password_credentials_account_id;
DROP TABLE auth.password_credentials;
