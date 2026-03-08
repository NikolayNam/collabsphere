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

    IF to_regclass('org.organization_legal_documents') IS NOT NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_documents" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE org.organization_legal_documents
(
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id       uuid         NOT NULL,
    document_type         varchar(64)  NOT NULL,
    status                varchar(32)  NOT NULL DEFAULT 'pending',
    object_id             uuid         NOT NULL,
    title                 varchar(255) NOT NULL,
    uploaded_by_account_id uuid        NULL,
    reviewer_account_id   uuid         NULL,
    review_note           text         NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NULL,
    reviewed_at           timestamptz  NULL,
    deleted_at            timestamptz  NULL,
    CONSTRAINT fk_org_organization_legal_documents_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_org_organization_legal_documents_object
        FOREIGN KEY (organization_id, object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_org_organization_legal_documents_uploaded_by
        FOREIGN KEY (uploaded_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_org_organization_legal_documents_reviewer
        FOREIGN KEY (reviewer_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_org_organization_legal_documents_document_type_not_blank
        CHECK (btrim(document_type) <> ''),
    CONSTRAINT chk_org_organization_legal_documents_status
        CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_org_organization_legal_documents_title_not_blank
        CHECK (btrim(title) <> ''),
    CONSTRAINT chk_org_organization_legal_documents_review_note_not_blank
        CHECK (review_note IS NULL OR btrim(review_note) <> ''),
    CONSTRAINT chk_org_organization_legal_documents_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),
    CONSTRAINT chk_org_organization_legal_documents_reviewed_at_valid
        CHECK (reviewed_at IS NULL OR reviewed_at >= created_at),
    CONSTRAINT chk_org_organization_legal_documents_deleted_at_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX ux_org_organization_legal_documents_object_active
    ON org.organization_legal_documents (organization_id, object_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_org_organization_legal_documents_type_status
    ON org.organization_legal_documents (organization_id, document_type, status)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_org_organization_legal_documents_created_at
    ON org.organization_legal_documents (organization_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down

DROP INDEX org.ix_org_organization_legal_documents_created_at;
DROP INDEX org.ix_org_organization_legal_documents_type_status;
DROP INDEX org.ux_org_organization_legal_documents_object_active;
DROP TABLE org.organization_legal_documents;
