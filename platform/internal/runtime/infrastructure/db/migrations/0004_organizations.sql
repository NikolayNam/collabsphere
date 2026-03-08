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
                     AND c.relname = 'organizations'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "org.organizations" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE org.organizations
(
    id             uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    name           varchar(255) NOT NULL,
    slug           varchar(255) NOT NULL,
    logo_object_id uuid         NULL,
    is_active      boolean      NOT NULL DEFAULT true,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    updated_at     timestamptz  NOT NULL DEFAULT now(),
    deleted_at     timestamptz  NULL,

    CONSTRAINT uq_org_organizations_slug
        UNIQUE (slug),

    CONSTRAINT chk_org_organizations_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_org_organizations_slug_not_blank
        CHECK (btrim(slug) <> '')
);

CREATE INDEX ix_org_organizations_is_active
    ON org.organizations (is_active);

CREATE INDEX ix_org_organizations_created_at
    ON org.organizations (created_at);


-- +goose Down

DROP INDEX org.ix_org_organizations_created_at;
DROP INDEX org.ix_org_organizations_is_active;
DROP TABLE org.organizations;