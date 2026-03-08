-- +goose Up

CREATE TABLE collab.message_mentions
(
    message_id   uuid        NOT NULL,
    account_id   uuid        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, account_id),
    CONSTRAINT fk_collab_message_mentions_message
        FOREIGN KEY (message_id)
            REFERENCES collab.messages (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_message_mentions_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_collab_message_mentions_account_id
    ON collab.message_mentions (account_id, created_at DESC);

-- +goose Down

DROP TABLE collab.message_mentions;
