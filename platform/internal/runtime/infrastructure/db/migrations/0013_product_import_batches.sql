-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_namespace
                       WHERE nspname = 'catalog') THEN
            RAISE EXCEPTION 'schema "catalog" does not exist';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist; run organizations migration first';
        END IF;

        IF to_regclass('storage.objects') IS NULL THEN
            RAISE EXCEPTION 'table "storage.objects" does not exist; run storage_objects migration first';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run accounts migration first';
        END IF;

        IF to_regclass('catalog.product_import_batches') IS NOT NULL THEN
            RAISE EXCEPTION 'table "catalog.product_import_batches" already exists; migration already applied or schema drift';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE catalog.product_import_batches
(
    id                    uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    organization_id       uuid        NOT NULL,
    source_object_id      uuid        NOT NULL,
    created_by_account_id uuid        NOT NULL,

    status                varchar(32) NOT NULL,
    total_rows            integer     NULL,
    processed_rows        integer     NOT NULL DEFAULT 0,
    success_rows          integer     NOT NULL DEFAULT 0,
    error_rows            integer     NOT NULL DEFAULT 0,


    started_by            varchar(32) NULL,

    started_at            timestamptz NOT NULL DEFAULT now(),
    finished_at           timestamptz NULL,

    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NULL,

    mode                  varchar(32) NULL, -- append / replace / upsert
    result_summary        jsonb       NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT fk_catalog_product_import_batches_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_catalog_product_import_batches_source_object
        FOREIGN KEY (organization_id, source_object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_catalog_product_import_batches_created_by_account
        FOREIGN KEY (created_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE RESTRICT,

    CONSTRAINT chk_catalog_product_import_batches_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),

    CONSTRAINT chk_catalog_product_import_batches_total_rows_non_negative
        CHECK (total_rows IS NULL OR total_rows >= 0),

    CONSTRAINT chk_catalog_product_import_batches_processed_rows_non_negative
        CHECK (processed_rows >= 0),

    CONSTRAINT chk_catalog_product_import_batches_success_rows_non_negative
        CHECK (success_rows >= 0),

    CONSTRAINT chk_catalog_product_import_batches_error_rows_non_negative
        CHECK (error_rows >= 0),

    CONSTRAINT chk_catalog_product_import_batches_processed_le_total
        CHECK (total_rows IS NULL OR processed_rows <= total_rows),

    CONSTRAINT chk_catalog_product_import_batches_success_le_processed
        CHECK (success_rows <= processed_rows),

    CONSTRAINT chk_catalog_product_import_batches_error_le_processed
        CHECK (error_rows <= processed_rows),

    CONSTRAINT chk_catalog_product_import_batches_success_error_sum_le_processed
        CHECK (success_rows + error_rows <= processed_rows),

    CONSTRAINT chk_catalog_product_import_batches_finished_at_valid
        CHECK (finished_at IS NULL OR finished_at >= started_at),

    CONSTRAINT chk_catalog_product_import_batches_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX ix_catalog_product_import_batches_organization_created_at
    ON catalog.product_import_batches (organization_id, created_at DESC);

CREATE INDEX ix_catalog_product_import_batches_status_created_at
    ON catalog.product_import_batches (status, created_at DESC);

CREATE INDEX ix_catalog_product_import_batches_source_object_id
    ON catalog.product_import_batches (source_object_id);

CREATE INDEX ix_catalog_product_import_batches_created_by_account_id
    ON catalog.product_import_batches (created_by_account_id);


-- +goose Down

DROP INDEX catalog.ix_catalog_product_import_batches_created_by_account_id;
DROP INDEX catalog.ix_catalog_product_import_batches_source_object_id;
DROP INDEX catalog.ix_catalog_product_import_batches_status_created_at;
DROP INDEX catalog.ix_catalog_product_import_batches_organization_created_at;

DROP TABLE catalog.product_import_batches;
