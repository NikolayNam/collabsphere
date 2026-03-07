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
                     AND c.relname = 'group_organization_members'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.group_organization_members" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE iam.group_organization_members
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        uuid        NOT NULL,
    organization_id uuid        NOT NULL,
    is_active       boolean     NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz NULL,

    CONSTRAINT fk_iam_group_organization_members_group
        FOREIGN KEY (group_id)
            REFERENCES iam.groups (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_iam_group_organization_members_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_iam_group_organization_members_group_organization
        UNIQUE (group_id, organization_id)
);

CREATE INDEX ix_iam_group_organization_members_group_id
    ON iam.group_organization_members (group_id);

CREATE INDEX ix_iam_group_organization_members_organization_id
    ON iam.group_organization_members (organization_id);


-- +goose Down

DROP INDEX iam.ix_iam_group_organization_members_organization_id;
DROP INDEX iam.ix_iam_group_organization_members_group_id;
DROP TABLE iam.group_organization_members;
