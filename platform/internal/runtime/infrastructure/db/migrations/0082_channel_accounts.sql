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
                     AND c.relname = 'channel_accounts'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.channel_accounts" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

-- Channel visibility: if a channel has rows here, only these accounts (within the group) can access.
-- If empty and channel_organizations is also empty, all group members can access (current behavior).
CREATE TABLE collab.channel_accounts
(
    channel_id uuid        NOT NULL,
    account_id uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_channel_accounts_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_channel_accounts_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_channel_accounts_channel_account
        UNIQUE (channel_id, account_id)
);

CREATE INDEX idx_channel_accounts_channel_id
    ON collab.channel_accounts (channel_id);

CREATE INDEX idx_channel_accounts_account_id
    ON collab.channel_accounts (account_id);

-- +goose Down

DROP TABLE collab.channel_accounts;
