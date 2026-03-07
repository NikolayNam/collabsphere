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

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'storage'
                     AND c.relname = 'objects'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "storage.objects" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE storage.objects
(
    id              uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id uuid         NULL,
    bucket          varchar(128) NOT NULL,
    object_key      text         NOT NULL,
    file_name       varchar(512) NOT NULL,
    content_type    varchar(255) NULL,
    size_bytes      bigint       NOT NULL,
    checksum_sha256 varchar(64)  NULL,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    deleted_at      timestamptz  NULL,

    CONSTRAINT fk_storage_objects_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE SET NULL,

    CONSTRAINT uq_storage_objects_bucket_object_key
        UNIQUE (bucket, object_key),

    CONSTRAINT uq_storage_objects_organization_id_id
        UNIQUE (organization_id, id),

    CONSTRAINT chk_storage_objects_bucket_not_blank
        CHECK (btrim(bucket) <> ''),

    CONSTRAINT chk_storage_objects_object_key_not_blank
        CHECK (btrim(object_key) <> ''),

    CONSTRAINT chk_storage_objects_file_name_not_blank
        CHECK (btrim(file_name) <> ''),

    CONSTRAINT chk_storage_objects_size_bytes_nonneg
        CHECK (size_bytes >= 0)
);

CREATE INDEX ix_storage_objects_organization_id
    ON storage.objects (organization_id);

CREATE INDEX ix_storage_objects_created_at
    ON storage.objects (created_at);


-- +goose Down

DROP INDEX storage.ix_storage_objects_created_at;
DROP INDEX storage.ix_storage_objects_organization_id;
DROP TABLE storage.objects;
