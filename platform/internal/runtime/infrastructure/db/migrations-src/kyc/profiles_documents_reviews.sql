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
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
        RAISE EXCEPTION 'schema "storage" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'kyc') THEN
        EXECUTE 'CREATE SCHEMA kyc';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS kyc.account_profiles
(
    account_id           uuid PRIMARY KEY,
    status               varchar(32) NOT NULL DEFAULT 'draft',
    legal_name           varchar(255) NULL,
    country_code         varchar(8) NULL,
    document_number      varchar(128) NULL,
    residence_address    text NULL,
    review_note          text NULL,
    reviewer_account_id  uuid NULL,
    submitted_at         timestamptz NULL,
    reviewed_at          timestamptz NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_kyc_account_profiles_account
        FOREIGN KEY (account_id) REFERENCES iam.accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_kyc_account_profiles_reviewer
        FOREIGN KEY (reviewer_account_id) REFERENCES iam.accounts (id) ON DELETE SET NULL,
    CONSTRAINT chk_kyc_account_profiles_status
        CHECK (status IN ('draft', 'submitted', 'in_review', 'needs_info', 'approved', 'rejected')),
    CONSTRAINT chk_kyc_account_profiles_review_note_not_blank
        CHECK (review_note IS NULL OR btrim(review_note) <> ''),
    CONSTRAINT chk_kyc_account_profiles_country_not_blank
        CHECK (country_code IS NULL OR btrim(country_code) <> ''),
    CONSTRAINT chk_kyc_account_profiles_document_number_not_blank
        CHECK (document_number IS NULL OR btrim(document_number) <> '')
);

CREATE TABLE IF NOT EXISTS kyc.organization_profiles
(
    organization_id      uuid PRIMARY KEY,
    status               varchar(32) NOT NULL DEFAULT 'draft',
    legal_name           varchar(255) NULL,
    country_code         varchar(8) NULL,
    registration_number  varchar(128) NULL,
    tax_id               varchar(128) NULL,
    review_note          text NULL,
    reviewer_account_id  uuid NULL,
    submitted_at         timestamptz NULL,
    reviewed_at          timestamptz NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_kyc_organization_profiles_organization
        FOREIGN KEY (organization_id) REFERENCES org.organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_kyc_organization_profiles_reviewer
        FOREIGN KEY (reviewer_account_id) REFERENCES iam.accounts (id) ON DELETE SET NULL,
    CONSTRAINT chk_kyc_organization_profiles_status
        CHECK (status IN ('draft', 'submitted', 'in_review', 'needs_info', 'approved', 'rejected')),
    CONSTRAINT chk_kyc_organization_profiles_review_note_not_blank
        CHECK (review_note IS NULL OR btrim(review_note) <> ''),
    CONSTRAINT chk_kyc_organization_profiles_country_not_blank
        CHECK (country_code IS NULL OR btrim(country_code) <> ''),
    CONSTRAINT chk_kyc_organization_profiles_registration_not_blank
        CHECK (registration_number IS NULL OR btrim(registration_number) <> ''),
    CONSTRAINT chk_kyc_organization_profiles_tax_id_not_blank
        CHECK (tax_id IS NULL OR btrim(tax_id) <> '')
);

CREATE TABLE IF NOT EXISTS kyc.account_documents
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id           uuid NOT NULL,
    object_id            uuid NOT NULL,
    document_type        varchar(64) NOT NULL,
    title                varchar(255) NOT NULL,
    status               varchar(32) NOT NULL DEFAULT 'uploaded',
    review_note          text NULL,
    reviewer_account_id  uuid NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NULL,
    reviewed_at          timestamptz NULL,
    deleted_at           timestamptz NULL,
    CONSTRAINT fk_kyc_account_documents_account
        FOREIGN KEY (account_id) REFERENCES iam.accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_kyc_account_documents_object
        FOREIGN KEY (object_id) REFERENCES storage.objects (id) ON DELETE RESTRICT,
    CONSTRAINT fk_kyc_account_documents_reviewer
        FOREIGN KEY (reviewer_account_id) REFERENCES iam.accounts (id) ON DELETE SET NULL,
    CONSTRAINT chk_kyc_account_documents_document_type_not_blank
        CHECK (btrim(document_type) <> ''),
    CONSTRAINT chk_kyc_account_documents_title_not_blank
        CHECK (btrim(title) <> ''),
    CONSTRAINT chk_kyc_account_documents_status
        CHECK (status IN ('pending_upload', 'uploaded', 'verified', 'rejected'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_kyc_account_documents_object_active
    ON kyc.account_documents (object_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_kyc_account_documents_account_created
    ON kyc.account_documents (account_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kyc.organization_documents
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      uuid NOT NULL,
    object_id            uuid NOT NULL,
    document_type        varchar(64) NOT NULL,
    title                varchar(255) NOT NULL,
    status               varchar(32) NOT NULL DEFAULT 'uploaded',
    review_note          text NULL,
    reviewer_account_id  uuid NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NULL,
    reviewed_at          timestamptz NULL,
    deleted_at           timestamptz NULL,
    CONSTRAINT fk_kyc_organization_documents_organization
        FOREIGN KEY (organization_id) REFERENCES org.organizations (id) ON DELETE CASCADE,
    CONSTRAINT fk_kyc_organization_documents_object
        FOREIGN KEY (organization_id, object_id) REFERENCES storage.objects (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT fk_kyc_organization_documents_reviewer
        FOREIGN KEY (reviewer_account_id) REFERENCES iam.accounts (id) ON DELETE SET NULL,
    CONSTRAINT chk_kyc_organization_documents_document_type_not_blank
        CHECK (btrim(document_type) <> ''),
    CONSTRAINT chk_kyc_organization_documents_title_not_blank
        CHECK (btrim(title) <> ''),
    CONSTRAINT chk_kyc_organization_documents_status
        CHECK (status IN ('pending_upload', 'uploaded', 'verified', 'rejected'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_kyc_organization_documents_object_active
    ON kyc.organization_documents (organization_id, object_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_kyc_organization_documents_org_created
    ON kyc.organization_documents (organization_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS kyc.review_events
(
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope               varchar(32) NOT NULL,
    subject_id          uuid NOT NULL,
    decision            varchar(32) NOT NULL,
    reason              text NULL,
    reviewer_account_id uuid NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_kyc_review_events_reviewer
        FOREIGN KEY (reviewer_account_id) REFERENCES iam.accounts (id) ON DELETE RESTRICT,
    CONSTRAINT chk_kyc_review_events_scope
        CHECK (scope IN ('account', 'organization')),
    CONSTRAINT chk_kyc_review_events_decision
        CHECK (decision IN ('approve', 'reject', 'request_info')),
    CONSTRAINT chk_kyc_review_events_reason_not_blank
        CHECK (reason IS NULL OR btrim(reason) <> '')
);

CREATE INDEX IF NOT EXISTS ix_kyc_review_events_scope_subject_created
    ON kyc.review_events (scope, subject_id, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS kyc.ix_kyc_review_events_scope_subject_created;
DROP TABLE IF EXISTS kyc.review_events;

DROP INDEX IF EXISTS kyc.ix_kyc_organization_documents_org_created;
DROP INDEX IF EXISTS kyc.ux_kyc_organization_documents_object_active;
DROP TABLE IF EXISTS kyc.organization_documents;

DROP INDEX IF EXISTS kyc.ix_kyc_account_documents_account_created;
DROP INDEX IF EXISTS kyc.ux_kyc_account_documents_object_active;
DROP TABLE IF EXISTS kyc.account_documents;

DROP TABLE IF EXISTS kyc.organization_profiles;
DROP TABLE IF EXISTS kyc.account_profiles;
