CREATE TABLE collab.channel_admins
(
    channel_id  uuid        NOT NULL,
    account_id  uuid        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    created_by  uuid        NULL,
    PRIMARY KEY (channel_id, account_id),
    CONSTRAINT fk_collab_channel_admins_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_channel_admins_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_channel_admins_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL
);

CREATE INDEX idx_collab_channel_admins_account_id
    ON collab.channel_admins (account_id);
