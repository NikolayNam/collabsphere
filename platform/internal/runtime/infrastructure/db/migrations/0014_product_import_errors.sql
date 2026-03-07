-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('catalog.product_import_batches') IS NULL THEN
            RAISE EXCEPTION 'table "catalog.product_import_batches" does not exist; run product_import_batches migration first';
        END IF;

        IF to_regclass('catalog.product_import_errors') IS NOT NULL THEN
            RAISE EXCEPTION 'table "catalog.product_import_errors" already exists; migration already applied or schema drift';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE catalog.product_import_errors
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    batch_id   uuid        NOT NULL,
    row_no     integer     NULL,
    code       text        NULL,
    message    text        NOT NULL,
    details    jsonb       NOT NULL DEFAULT '{}'::jsonb,

    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_catalog_product_import_errors_batch
        FOREIGN KEY (batch_id)
            REFERENCES catalog.product_import_batches (id)
            ON DELETE CASCADE,

    CONSTRAINT chk_catalog_product_import_errors_row_no_positive
        CHECK (row_no IS NULL OR row_no > 0),

    CONSTRAINT chk_catalog_product_import_errors_message_not_blank
        CHECK (btrim(message) <> '')
);

CREATE INDEX ix_catalog_product_import_errors_batch_row
    ON catalog.product_import_errors (batch_id, row_no);

CREATE INDEX ix_catalog_product_import_errors_batch_created_at
    ON catalog.product_import_errors (batch_id, created_at);


-- +goose Down

DROP INDEX catalog.ix_catalog_product_import_errors_batch_created_at;
DROP INDEX catalog.ix_catalog_product_import_errors_batch_row;

DROP TABLE catalog.product_import_errors;