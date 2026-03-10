-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'auth') THEN
            RAISE EXCEPTION 'schema "auth" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relname = 'oauth_states'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "auth.oauth_states" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.oauth_states
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider   varchar(64) NOT NULL,
    state_hash text        NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NULL,
    CONSTRAINT uq_auth_oauth_states_state_hash
        UNIQUE (state_hash),
    CONSTRAINT chk_auth_oauth_states_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_auth_oauth_states_state_hash_not_blank
        CHECK (btrim(state_hash) <> ''),
    CONSTRAINT chk_auth_oauth_states_expiry_valid
        CHECK (expires_at > created_at),
    CONSTRAINT chk_auth_oauth_states_used_valid
        CHECK (used_at IS NULL OR used_at >= created_at),
    CONSTRAINT chk_auth_oauth_states_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX ix_auth_oauth_states_provider
    ON auth.oauth_states (provider);

CREATE INDEX ix_auth_oauth_states_expires_at
    ON auth.oauth_states (expires_at);

-- +goose Down

DROP TABLE auth.oauth_states;
