-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_class c
                                JOIN pg_namespace n ON n.oid = c.relnamespace
                       WHERE n.nspname = 'iam'
                         AND c.relname = 'groups'
                         AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.groups" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'iam'
                     AND c.relname = 'group_account_members'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.group_account_members" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE iam.group_account_members
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    group_id    uuid        NOT NULL,
    account_id  uuid        NOT NULL,
    role        varchar(64) NOT NULL DEFAULT 'member',
    is_active   boolean     NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz NULL,

    CONSTRAINT fk_iam_group_account_members_group
        FOREIGN KEY (group_id)
            REFERENCES iam.groups (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_iam_group_account_members_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_iam_group_account_members_group_account
        UNIQUE (group_id, account_id),

    CONSTRAINT chk_iam_group_account_members_role
        CHECK (role IN ('owner', 'member'))
);

CREATE INDEX ix_iam_group_account_members_group_id
    ON iam.group_account_members (group_id);

CREATE INDEX ix_iam_group_account_members_account_id
    ON iam.group_account_members (account_id);

CREATE INDEX ix_iam_group_account_members_role
    ON iam.group_account_members (role);


-- +goose Down

DROP INDEX iam.ix_iam_group_account_members_role;
DROP INDEX iam.ix_iam_group_account_members_account_id;
DROP INDEX iam.ix_iam_group_account_members_group_id;
DROP TABLE iam.group_account_members;
