-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('auth.oauth_states') IS NULL THEN
            RAISE EXCEPTION 'table "auth.oauth_states" does not exist; run oauth_states migration first';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relname = 'oidc_nonces'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "auth.oidc_nonces" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.oidc_nonces
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider       varchar(64) NOT NULL,
    oauth_state_id uuid        NOT NULL,
    nonce_hash     text        NOT NULL,
    expires_at     timestamptz NOT NULL,
    used_at        timestamptz NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NULL,
    CONSTRAINT fk_auth_oidc_nonces_oauth_state
        FOREIGN KEY (oauth_state_id)
            REFERENCES auth.oauth_states (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_auth_oidc_nonces_state
        UNIQUE (oauth_state_id),
    CONSTRAINT uq_auth_oidc_nonces_nonce_hash
        UNIQUE (nonce_hash),
    CONSTRAINT chk_auth_oidc_nonces_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_auth_oidc_nonces_nonce_hash_not_blank
        CHECK (btrim(nonce_hash) <> ''),
    CONSTRAINT chk_auth_oidc_nonces_expiry_valid
        CHECK (expires_at > created_at),
    CONSTRAINT chk_auth_oidc_nonces_used_valid
        CHECK (used_at IS NULL OR used_at >= created_at),
    CONSTRAINT chk_auth_oidc_nonces_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX ix_auth_oidc_nonces_provider
    ON auth.oidc_nonces (provider);

CREATE INDEX ix_auth_oidc_nonces_expires_at
    ON auth.oidc_nonces (expires_at);

-- +goose Down

DROP TABLE auth.oidc_nonces;
