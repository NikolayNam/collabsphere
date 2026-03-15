-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'org'
                     AND c.relname = 'organization_roles'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "org.organization_roles" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE org.organization_roles
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    organization_id uuid        NOT NULL,
    code            varchar(64)  NOT NULL,
    name            varchar(255) NOT NULL,
    description     text,
    base_role       varchar(64)  NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz NULL,

    CONSTRAINT fk_org_organization_roles_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT chk_org_organization_roles_base_role
        CHECK (base_role IN ('owner', 'admin', 'manager', 'member', 'viewer')),

    CONSTRAINT chk_org_organization_roles_code_not_blank
        CHECK (btrim(code) <> '')
);

CREATE UNIQUE INDEX uq_org_organization_roles_org_code_active
    ON org.organization_roles (organization_id, code)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_org_organization_roles_organization_id
    ON org.organization_roles (organization_id);

CREATE INDEX ix_org_organization_roles_deleted_at
    ON org.organization_roles (organization_id, deleted_at)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE org.organization_roles IS 'Custom organization roles extending system base roles. Soft-deleted via deleted_at.';

-- Relax iam.memberships role constraint to allow custom role codes
-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM pg_constraint c
                            JOIN pg_class t ON t.oid = c.conrelid
                            JOIN pg_namespace n ON n.oid = t.relnamespace
                   WHERE n.nspname = 'iam'
                     AND t.relname = 'memberships'
                     AND c.conname = 'chk_iam_memberships_role_allowed') THEN
            ALTER TABLE iam.memberships DROP CONSTRAINT chk_iam_memberships_role_allowed;
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint c
                                JOIN pg_class t ON t.oid = c.conrelid
                                JOIN pg_namespace n ON n.oid = t.relnamespace
                       WHERE n.nspname = 'iam'
                         AND t.relname = 'memberships'
                         AND c.conname = 'chk_iam_memberships_role_not_blank') THEN
            ALTER TABLE iam.memberships
                ADD CONSTRAINT chk_iam_memberships_role_not_blank
                    CHECK (btrim(role) <> '');
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose Down

ALTER TABLE iam.memberships DROP CONSTRAINT IF EXISTS chk_iam_memberships_role_not_blank;

ALTER TABLE iam.memberships
    ADD CONSTRAINT chk_iam_memberships_role_allowed
        CHECK (role IN ('owner', 'admin', 'manager', 'member', 'viewer'));

DROP TABLE IF EXISTS org.organization_roles;
