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

CREATE TABLE auth.refresh_sessions
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    account_id uuid        NOT NULL REFERENCES iam.accounts (id) ON DELETE CASCADE,
    token_hash text        NOT NULL,

    user_agent text        NULL,
    ip         varchar(64) NULL,

    expires_at timestamptz NOT NULL,
    revoked_at timestamptz NULL,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NULL,

    CONSTRAINT chk_auth_refresh_sessions_token_hash_not_blank
        CHECK (btrim(token_hash) <> ''),

    CONSTRAINT chk_auth_refresh_sessions_expiry_valid
        CHECK (expires_at > created_at),

    CONSTRAINT chk_auth_refresh_sessions_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),

    CONSTRAINT chk_auth_refresh_sessions_revoked_at_valid
        CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE UNIQUE INDEX ux_auth_refresh_sessions_token_hash
    ON auth.refresh_sessions (token_hash);

CREATE INDEX ix_auth_refresh_sessions_account_id
    ON auth.refresh_sessions (account_id);

CREATE INDEX ix_auth_refresh_sessions_expires_at
    ON auth.refresh_sessions (expires_at);

CREATE INDEX ix_auth_refresh_sessions_revoked_at
    ON auth.refresh_sessions (revoked_at);

-- +goose Down
DROP TABLE auth.refresh_sessions;