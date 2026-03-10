-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
            RAISE EXCEPTION 'schema "storage" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'storage'
                     AND c.relname = 'uploads'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "storage.uploads" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE storage.uploads
(
    id                    uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id       uuid         NULL,
    object_id             uuid         NOT NULL,
    created_by_account_id uuid         NOT NULL,
    purpose               varchar(64)  NOT NULL,
    status                varchar(32)  NOT NULL,
    bucket                varchar(128) NOT NULL,
    object_key            text         NOT NULL,
    file_name             varchar(512) NOT NULL,
    content_type          varchar(255) NULL,
    declared_size_bytes   bigint       NOT NULL DEFAULT 0,
    actual_size_bytes     bigint       NULL,
    checksum_sha256       varchar(64)  NULL,
    metadata              jsonb        NOT NULL DEFAULT '{}'::jsonb,
    error_code            varchar(128) NULL,
    error_message         text         NULL,
    result_kind           varchar(64)  NULL,
    result_id             uuid         NULL,
    completed_at          timestamptz  NULL,
    expires_at            timestamptz  NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NULL,

    CONSTRAINT fk_storage_uploads_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_storage_uploads_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_storage_uploads_created_by_account
        FOREIGN KEY (created_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE RESTRICT,

    CONSTRAINT uq_storage_uploads_object_id
        UNIQUE (object_id),

    CONSTRAINT chk_storage_uploads_purpose_not_blank
        CHECK (btrim(purpose) <> ''),

    CONSTRAINT chk_storage_uploads_status_not_blank
        CHECK (btrim(status) <> ''),

    CONSTRAINT chk_storage_uploads_bucket_not_blank
        CHECK (btrim(bucket) <> ''),

    CONSTRAINT chk_storage_uploads_object_key_not_blank
        CHECK (btrim(object_key) <> ''),

    CONSTRAINT chk_storage_uploads_file_name_not_blank
        CHECK (btrim(file_name) <> ''),

    CONSTRAINT chk_storage_uploads_declared_size_nonneg
        CHECK (declared_size_bytes >= 0),

    CONSTRAINT chk_storage_uploads_actual_size_nonneg
        CHECK (actual_size_bytes IS NULL OR actual_size_bytes >= 0)
);

CREATE INDEX ix_storage_uploads_organization_id
    ON storage.uploads (organization_id);

CREATE INDEX ix_storage_uploads_created_by_account_id
    ON storage.uploads (created_by_account_id);

CREATE INDEX ix_storage_uploads_status
    ON storage.uploads (status);

CREATE INDEX ix_storage_uploads_created_at
    ON storage.uploads (created_at DESC);


-- +goose Down

DROP INDEX storage.ix_storage_uploads_created_at;
DROP INDEX storage.ix_storage_uploads_status;
DROP INDEX storage.ix_storage_uploads_created_by_account_id;
DROP INDEX storage.ix_storage_uploads_organization_id;
DROP TABLE storage.uploads;
