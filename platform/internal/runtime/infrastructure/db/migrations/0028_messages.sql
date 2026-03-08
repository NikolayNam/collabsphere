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
                     AND c.relname = 'messages'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.messages" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd


CREATE TABLE collab.messages
(
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id          uuid        NOT NULL,
    channel_seq         bigint      NOT NULL,
    message_type        text        NOT NULL DEFAULT 'user',
    author_type         text        NOT NULL,
    author_account_id   uuid        NULL,
    author_guest_id     uuid        NULL,
    body                text        NOT NULL,
    reply_to_message_id uuid        NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    edited_at           timestamptz NULL,
    deleted_at          timestamptz NULL,
    CONSTRAINT fk_collab_messages_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_messages_author_account
        FOREIGN KEY (author_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_collab_messages_author_guest
        FOREIGN KEY (author_guest_id)
            REFERENCES auth.guest_identities (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_collab_messages_reply_to
        FOREIGN KEY (reply_to_message_id)
            REFERENCES collab.messages (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_collab_messages_channel_seq
        UNIQUE (channel_id, channel_seq),
    CONSTRAINT chk_collab_messages_type
        CHECK (message_type IN ('user', 'system')),
    CONSTRAINT chk_collab_messages_author_type
        CHECK (author_type IN ('account', 'guest', 'system')),
    CONSTRAINT chk_collab_messages_author_match
        CHECK (
            (author_type = 'account' AND author_account_id IS NOT NULL AND author_guest_id IS NULL)
            OR (author_type = 'guest' AND author_guest_id IS NOT NULL AND author_account_id IS NULL)
            OR (author_type = 'system' AND author_account_id IS NULL AND author_guest_id IS NULL)
        ),
    CONSTRAINT chk_collab_messages_channel_seq_positive
        CHECK (channel_seq > 0),
    CONSTRAINT chk_collab_messages_edited_valid
        CHECK (edited_at IS NULL OR edited_at >= created_at),
    CONSTRAINT chk_collab_messages_deleted_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE INDEX idx_collab_messages_channel_created_at
    ON collab.messages (channel_id, created_at DESC);

-- +goose Down

DROP TABLE collab.messages;
