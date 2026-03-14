-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('auth.refresh_sessions') IS NULL THEN
            RAISE EXCEPTION 'table "auth.refresh_sessions" does not exist; run refresh_sessions migration first';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relname = 'refresh_session_tokens'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "auth.refresh_session_tokens" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE auth.refresh_session_tokens
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid        NOT NULL,
    token_hash text        NOT NULL,
    used_at    timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_auth_refresh_session_tokens_session
        FOREIGN KEY (session_id)
            REFERENCES auth.refresh_sessions (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_auth_refresh_session_tokens_token_hash
        UNIQUE (token_hash),

    CONSTRAINT chk_auth_refresh_session_tokens_token_hash_not_blank
        CHECK (btrim(token_hash) <> ''),

    CONSTRAINT chk_auth_refresh_session_tokens_used_valid
        CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX ix_auth_refresh_session_tokens_session_id
    ON auth.refresh_session_tokens (session_id);

CREATE INDEX ix_auth_refresh_session_tokens_used_at
    ON auth.refresh_session_tokens (used_at);

INSERT INTO auth.refresh_session_tokens (id, session_id, token_hash, created_at)
SELECT gen_random_uuid(), id, token_hash, created_at
FROM auth.refresh_sessions;

-- +goose Down

DROP TABLE auth.refresh_session_tokens;
