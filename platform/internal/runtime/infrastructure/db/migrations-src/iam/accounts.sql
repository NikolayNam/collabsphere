-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'iam'
                     AND c.relname = 'accounts'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.accounts" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE iam.accounts
(
    id               uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    email            varchar(320) NOT NULL,
    display_name     varchar(255) NULL,
    avatar_object_id uuid         NULL,
    is_active        boolean      NOT NULL DEFAULT true,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now(),
    deleted_at       timestamptz  NULL,

    CONSTRAINT uq_iam_accounts_email
        UNIQUE (email),

    CONSTRAINT chk_iam_accounts_email_not_blank
        CHECK (btrim(email) <> '')
);

CREATE INDEX ix_iam_accounts_is_active
    ON iam.accounts (is_active);

CREATE INDEX ix_iam_accounts_created_at
    ON iam.accounts (created_at);


-- +goose Down

DROP INDEX iam.ix_iam_accounts_created_at;
DROP INDEX iam.ix_iam_accounts_is_active;
DROP TABLE iam.accounts;