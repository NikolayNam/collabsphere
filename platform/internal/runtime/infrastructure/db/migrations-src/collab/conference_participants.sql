-- +goose Up

CREATE TABLE collab.conference_participants
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conference_id  uuid        NOT NULL,
    actor_type     text        NOT NULL,
    account_id     uuid        NULL,
    guest_id       uuid        NULL,
    joined_at      timestamptz NOT NULL DEFAULT now(),
    left_at        timestamptz NULL,
    role           text        NOT NULL DEFAULT 'participant',
    CONSTRAINT fk_collab_conference_participants_conference
        FOREIGN KEY (conference_id)
            REFERENCES collab.conferences (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_conference_participants_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_conference_participants_guest
        FOREIGN KEY (guest_id)
            REFERENCES auth.guest_identities (id)
            ON DELETE CASCADE,
    CONSTRAINT chk_collab_conference_participants_actor_type
        CHECK (actor_type IN ('account', 'guest')),
    CONSTRAINT chk_collab_conference_participants_actor_match
        CHECK (
            (actor_type = 'account' AND account_id IS NOT NULL AND guest_id IS NULL)
            OR (actor_type = 'guest' AND guest_id IS NOT NULL AND account_id IS NULL)
        ),
    CONSTRAINT chk_collab_conference_participants_left_valid
        CHECK (left_at IS NULL OR left_at >= joined_at),
    CONSTRAINT chk_collab_conference_participants_role
        CHECK (role IN ('participant', 'moderator'))
);

CREATE INDEX idx_collab_conference_participants_conference_id
    ON collab.conference_participants (conference_id, joined_at DESC);

-- +goose Down

DROP TABLE collab.conference_participants;
