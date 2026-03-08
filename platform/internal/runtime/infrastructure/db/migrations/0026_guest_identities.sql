-- +goose Up

CREATE TABLE auth.guest_identities
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invite_id        uuid        NOT NULL,
    channel_id       uuid        NOT NULL,
    email            citext      NOT NULL,
    display_name     text        NOT NULL,
    accepted_at      timestamptz NOT NULL DEFAULT now(),
    expires_at       timestamptz NOT NULL,
    last_seen_at     timestamptz NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NULL,
    CONSTRAINT fk_auth_guest_identities_invite
        FOREIGN KEY (invite_id)
            REFERENCES collab.guest_invites (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_auth_guest_identities_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_auth_guest_identities_invite
        UNIQUE (invite_id),
    CONSTRAINT chk_auth_guest_identities_display_name_not_blank
        CHECK (btrim(display_name) <> ''),
    CONSTRAINT chk_auth_guest_identities_expiry_after_accept
        CHECK (expires_at > accepted_at),
    CONSTRAINT chk_auth_guest_identities_last_seen_valid
        CHECK (last_seen_at IS NULL OR last_seen_at >= accepted_at),
    CONSTRAINT chk_auth_guest_identities_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_auth_guest_identities_channel_id
    ON auth.guest_identities (channel_id);

-- +goose Down

DROP TABLE auth.guest_identities;
