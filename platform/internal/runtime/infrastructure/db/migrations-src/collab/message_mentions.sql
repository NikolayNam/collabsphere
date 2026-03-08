-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run orders migration first';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'message_mentions'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.message_mentions" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd


CREATE TABLE collab.message_mentions
(
    message_id uuid        NOT NULL,
    account_id uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
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
