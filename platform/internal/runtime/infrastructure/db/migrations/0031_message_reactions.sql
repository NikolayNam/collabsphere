-- +goose Up

CREATE TABLE collab.message_reactions
(
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  uuid        NOT NULL,
    actor_type  text        NOT NULL,
    account_id  uuid        NULL,
    guest_id    uuid        NULL,
    emoji       text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_collab_message_reactions_message
        FOREIGN KEY (message_id)
            REFERENCES collab.messages (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_message_reactions_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_message_reactions_guest
        FOREIGN KEY (guest_id)
            REFERENCES auth.guest_identities (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_collab_message_reactions_actor
        UNIQUE NULLS NOT DISTINCT (message_id, actor_type, account_id, guest_id, emoji),
    CONSTRAINT chk_collab_message_reactions_actor_type
        CHECK (actor_type IN ('account', 'guest')),
    CONSTRAINT chk_collab_message_reactions_actor_match
        CHECK (
            (actor_type = 'account' AND account_id IS NOT NULL AND guest_id IS NULL)
            OR (actor_type = 'guest' AND guest_id IS NOT NULL AND account_id IS NULL)
        ),
    CONSTRAINT chk_collab_message_reactions_emoji_not_blank
        CHECK (btrim(emoji) <> '')
);

-- +goose Down

DROP TABLE collab.message_reactions;
