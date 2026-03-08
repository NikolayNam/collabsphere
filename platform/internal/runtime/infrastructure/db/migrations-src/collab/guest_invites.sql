-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'guest_invites'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.guest_invites" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd


CREATE TABLE collab.guest_invites
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id           uuid        NOT NULL,
    email                text      NOT NULL,
    token_hash           text        NOT NULL,
    can_post             boolean     NOT NULL DEFAULT true,
    visible_from_seq     bigint      NOT NULL DEFAULT 0,
    expires_at           timestamptz NOT NULL,
    accepted_at          timestamptz NULL,
    revoked_at           timestamptz NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NULL,
    invited_by           uuid        NOT NULL,
    accepted_by_guest_id uuid        NULL,
    CONSTRAINT fk_collab_guest_invites_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_guest_invites_invited_by
        FOREIGN KEY (invited_by)
            REFERENCES iam.accounts (id)
            ON DELETE RESTRICT,
    CONSTRAINT uq_collab_guest_invites_token_hash
        UNIQUE (token_hash),
    CONSTRAINT chk_collab_guest_invites_token_hash_not_blank
        CHECK (btrim(token_hash) <> ''),
    CONSTRAINT chk_collab_guest_invites_visible_from_seq_nonneg
        CHECK (visible_from_seq >= 0),
    CONSTRAINT chk_collab_guest_invites_expiry_after_create
        CHECK (expires_at > created_at),
    CONSTRAINT chk_collab_guest_invites_accepted_valid
        CHECK (accepted_at IS NULL OR accepted_at >= created_at),
    CONSTRAINT chk_collab_guest_invites_revoked_valid
        CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT chk_collab_guest_invites_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_collab_guest_invites_channel_id
    ON collab.guest_invites (channel_id);

CREATE INDEX idx_collab_guest_invites_email
    ON collab.guest_invites (email);

-- +goose Down

DROP TABLE collab.guest_invites;
