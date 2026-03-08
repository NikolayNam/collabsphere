-- +goose Up

CREATE TABLE auth.guest_sessions
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_id       uuid        NOT NULL,
    token_hash     text        NOT NULL,
    user_agent     text        NULL,
    ip_address     inet        NULL,
    expires_at     timestamptz NOT NULL,
    last_used_at   timestamptz NULL,
    revoked_at     timestamptz NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NULL,
    CONSTRAINT fk_auth_guest_sessions_guest
        FOREIGN KEY (guest_id)
            REFERENCES auth.guest_identities (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_auth_guest_sessions_token_hash
        UNIQUE (token_hash),
    CONSTRAINT chk_auth_guest_sessions_token_hash_not_blank
        CHECK (btrim(token_hash) <> ''),
    CONSTRAINT chk_auth_guest_sessions_expiry_after_create
        CHECK (expires_at > created_at),
    CONSTRAINT chk_auth_guest_sessions_last_used_valid
        CHECK (last_used_at IS NULL OR last_used_at >= created_at),
    CONSTRAINT chk_auth_guest_sessions_revoked_valid
        CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT chk_auth_guest_sessions_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_auth_guest_sessions_guest_id
    ON auth.guest_sessions (guest_id);

-- +goose Down

DROP TABLE auth.guest_sessions;
