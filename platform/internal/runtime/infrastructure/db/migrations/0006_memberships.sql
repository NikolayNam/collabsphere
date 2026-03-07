-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'iam'
                     AND c.relname = 'memberships'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.memberships" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE iam.memberships
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    organization_id uuid        NOT NULL,
    account_id      uuid        NOT NULL,
    role            varchar(64) NOT NULL,
    is_active       boolean     NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz NULL,

    CONSTRAINT fk_iam_memberships_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_iam_memberships_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_iam_memberships_org_account
        UNIQUE (organization_id, account_id),

    CONSTRAINT chk_iam_memberships_role_not_blank
        CHECK (btrim(role) <> '')
);

CREATE INDEX ix_iam_memberships_account_id
    ON iam.memberships (account_id);

CREATE INDEX ix_iam_memberships_organization_id
    ON iam.memberships (organization_id);

CREATE INDEX ix_iam_memberships_role
    ON iam.memberships (role);


-- +goose Down

DROP INDEX iam.ix_iam_memberships_role;
DROP INDEX iam.ix_iam_memberships_organization_id;
DROP INDEX iam.ix_iam_memberships_account_id;
DROP TABLE iam.memberships;