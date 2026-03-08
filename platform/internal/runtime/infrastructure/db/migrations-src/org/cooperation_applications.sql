-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
        RAISE EXCEPTION 'schema "org" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
        RAISE EXCEPTION 'schema "iam" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
        RAISE EXCEPTION 'schema "storage" does not exist';
    END IF;

    IF to_regclass('org.organizations') IS NULL THEN
        RAISE EXCEPTION 'table "org.organizations" does not exist';
    END IF;

    IF to_regclass('iam.accounts') IS NULL THEN
        RAISE EXCEPTION 'table "iam.accounts" does not exist';
    END IF;

    IF to_regclass('storage.objects') IS NULL THEN
        RAISE EXCEPTION 'table "storage.objects" does not exist';
    END IF;

    IF to_regclass('org.cooperation_applications') IS NOT NULL THEN
        RAISE EXCEPTION 'table "org.cooperation_applications" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE org.cooperation_applications
(
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id        uuid         NOT NULL,
    status                 varchar(32)  NOT NULL DEFAULT 'draft',
    confirmation_email     varchar(320) NULL,
    company_name           varchar(255) NULL,
    represented_categories text         NULL,
    minimum_order_amount   varchar(128) NULL,
    delivery_geography     text         NULL,
    sales_channels         jsonb        NOT NULL DEFAULT '[]'::jsonb,
    storefront_url         varchar(512) NULL,
    contact_first_name     varchar(128) NULL,
    contact_last_name      varchar(128) NULL,
    contact_job_title      varchar(128) NULL,
    price_list_object_id   uuid         NULL,
    contact_email          varchar(320) NULL,
    contact_phone          varchar(32)  NULL,
    partner_code           varchar(128) NULL,
    review_note            text         NULL,
    reviewer_account_id    uuid         NULL,
    submitted_at           timestamptz  NULL,
    reviewed_at            timestamptz  NULL,
    created_at             timestamptz  NOT NULL DEFAULT now(),
    updated_at             timestamptz  NULL,
    CONSTRAINT uq_org_cooperation_applications_organization
        UNIQUE (organization_id),
    CONSTRAINT fk_org_cooperation_applications_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_org_cooperation_applications_price_list_object
        FOREIGN KEY (organization_id, price_list_object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_org_cooperation_applications_reviewer_account
        FOREIGN KEY (reviewer_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_org_cooperation_applications_status
        CHECK (status IN ('draft', 'submitted', 'under_review', 'approved', 'rejected', 'needs_info')),
    CONSTRAINT chk_org_cooperation_applications_confirmation_email_not_blank
        CHECK (confirmation_email IS NULL OR btrim(confirmation_email) <> ''),
    CONSTRAINT chk_org_cooperation_applications_company_name_not_blank
        CHECK (company_name IS NULL OR btrim(company_name) <> ''),
    CONSTRAINT chk_org_cooperation_applications_represented_categories_not_blank
        CHECK (represented_categories IS NULL OR btrim(represented_categories) <> ''),
    CONSTRAINT chk_org_cooperation_applications_minimum_order_amount_not_blank
        CHECK (minimum_order_amount IS NULL OR btrim(minimum_order_amount) <> ''),
    CONSTRAINT chk_org_cooperation_applications_delivery_geography_not_blank
        CHECK (delivery_geography IS NULL OR btrim(delivery_geography) <> ''),
    CONSTRAINT chk_org_cooperation_applications_storefront_url_not_blank
        CHECK (storefront_url IS NULL OR btrim(storefront_url) <> ''),
    CONSTRAINT chk_org_cooperation_applications_contact_first_name_not_blank
        CHECK (contact_first_name IS NULL OR btrim(contact_first_name) <> ''),
    CONSTRAINT chk_org_cooperation_applications_contact_last_name_not_blank
        CHECK (contact_last_name IS NULL OR btrim(contact_last_name) <> ''),
    CONSTRAINT chk_org_cooperation_applications_contact_job_title_not_blank
        CHECK (contact_job_title IS NULL OR btrim(contact_job_title) <> ''),
    CONSTRAINT chk_org_cooperation_applications_contact_email_not_blank
        CHECK (contact_email IS NULL OR btrim(contact_email) <> ''),
    CONSTRAINT chk_org_cooperation_applications_contact_phone_not_blank
        CHECK (contact_phone IS NULL OR btrim(contact_phone) <> ''),
    CONSTRAINT chk_org_cooperation_applications_partner_code_not_blank
        CHECK (partner_code IS NULL OR btrim(partner_code) <> ''),
    CONSTRAINT chk_org_cooperation_applications_review_note_not_blank
        CHECK (review_note IS NULL OR btrim(review_note) <> ''),
    CONSTRAINT chk_org_cooperation_applications_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),
    CONSTRAINT chk_org_cooperation_applications_reviewed_at_valid
        CHECK (reviewed_at IS NULL OR reviewed_at >= created_at),
    CONSTRAINT chk_org_cooperation_applications_submitted_at_valid
        CHECK (submitted_at IS NULL OR submitted_at >= created_at),
    CONSTRAINT chk_org_cooperation_applications_sales_channels_array
        CHECK (jsonb_typeof(sales_channels) = 'array')
);

CREATE INDEX ix_org_cooperation_applications_status
    ON org.cooperation_applications (status);

CREATE INDEX ix_org_cooperation_applications_submitted_at
    ON org.cooperation_applications (submitted_at DESC);

-- +goose Down

DROP INDEX org.ix_org_cooperation_applications_submitted_at;
DROP INDEX org.ix_org_cooperation_applications_status;
DROP TABLE org.cooperation_applications;
