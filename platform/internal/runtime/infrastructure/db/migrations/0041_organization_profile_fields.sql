-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
        RAISE EXCEPTION 'schema "org" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
        RAISE EXCEPTION 'schema "storage" does not exist';
    END IF;

    IF to_regclass('org.organizations') IS NULL THEN
        RAISE EXCEPTION 'table "org.organizations" does not exist';
    END IF;

    IF to_regclass('storage.objects') IS NULL THEN
        RAISE EXCEPTION 'table "storage.objects" does not exist';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'org'
          AND table_name = 'organizations'
          AND column_name = 'logo_object_id'
    ) THEN
        RAISE EXCEPTION 'column "org.organizations.logo_object_id" does not exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM org.organizations AS o
        LEFT JOIN storage.objects AS so
            ON so.id = o.logo_object_id
           AND so.organization_id = o.id
        WHERE o.logo_object_id IS NOT NULL
          AND so.id IS NULL
    ) THEN
        RAISE EXCEPTION 'org.organizations contains invalid logo_object_id references';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE org.organizations
    ADD COLUMN IF NOT EXISTS description text NULL,
    ADD COLUMN IF NOT EXISTS website varchar(512) NULL,
    ADD COLUMN IF NOT EXISTS primary_email varchar(320) NULL,
    ADD COLUMN IF NOT EXISTS phone varchar(32) NULL,
    ADD COLUMN IF NOT EXISTS address text NULL,
    ADD COLUMN IF NOT EXISTS industry varchar(128) NULL;

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'fk_org_organizations_logo_object'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT fk_org_organizations_logo_object
                FOREIGN KEY (id, logo_object_id)
                    REFERENCES storage.objects (organization_id, id)
                    ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_description_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_description_not_blank
                CHECK (description IS NULL OR btrim(description) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_website_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_website_not_blank
                CHECK (website IS NULL OR btrim(website) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_primary_email_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_primary_email_not_blank
                CHECK (primary_email IS NULL OR btrim(primary_email) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_phone_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_phone_not_blank
                CHECK (phone IS NULL OR btrim(phone) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_address_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_address_not_blank
                CHECK (address IS NULL OR btrim(address) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'org.organizations'::regclass
          AND conname = 'chk_org_organizations_industry_not_blank'
    ) THEN
        ALTER TABLE org.organizations
            ADD CONSTRAINT chk_org_organizations_industry_not_blank
                CHECK (industry IS NULL OR btrim(industry) <> '');
    END IF;
END
$$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS ix_org_organizations_logo_object_id
    ON org.organizations (logo_object_id)
    WHERE logo_object_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS org.ix_org_organizations_logo_object_id;

ALTER TABLE org.organizations
    DROP CONSTRAINT IF EXISTS fk_org_organizations_logo_object,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_description_not_blank,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_website_not_blank,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_primary_email_not_blank,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_phone_not_blank,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_address_not_blank,
    DROP CONSTRAINT IF EXISTS chk_org_organizations_industry_not_blank,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS primary_email,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS industry;
