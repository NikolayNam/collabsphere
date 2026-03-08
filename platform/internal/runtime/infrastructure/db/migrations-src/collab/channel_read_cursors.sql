CREATE TABLE collab.channel_read_cursors
(
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id         uuid        NOT NULL,
    actor_type         text        NOT NULL,
    account_id         uuid        NULL,
    guest_id           uuid        NULL,
    last_read_seq      bigint      NOT NULL DEFAULT 0,
    last_read_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_collab_channel_read_cursors_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_channel_read_cursors_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_channel_read_cursors_guest
        FOREIGN KEY (guest_id)
            REFERENCES auth.guest_identities (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_collab_channel_read_cursors_actor
        UNIQUE NULLS NOT DISTINCT (channel_id, actor_type, account_id, guest_id),
    CONSTRAINT chk_collab_channel_read_cursors_actor_type
        CHECK (actor_type IN ('account', 'guest')),
    CONSTRAINT chk_collab_channel_read_cursors_actor_match
        CHECK (
            (actor_type = 'account' AND account_id IS NOT NULL AND guest_id IS NULL)
            OR (actor_type = 'guest' AND guest_id IS NOT NULL AND account_id IS NULL)
        ),
    CONSTRAINT chk_collab_channel_read_cursors_last_read_seq_nonneg
        CHECK (last_read_seq >= 0)
);
