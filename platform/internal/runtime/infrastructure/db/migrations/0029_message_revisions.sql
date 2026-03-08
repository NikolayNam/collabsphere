CREATE TABLE collab.message_revisions
(
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  uuid        NOT NULL,
    body        text        NOT NULL,
    edited_by   uuid        NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_collab_message_revisions_message
        FOREIGN KEY (message_id)
            REFERENCES collab.messages (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_message_revisions_edited_by
        FOREIGN KEY (edited_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL
);

CREATE INDEX idx_collab_message_revisions_message_id
    ON collab.message_revisions (message_id, created_at DESC);
