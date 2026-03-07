-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_namespace
            WHERE nspname = 'sales'
        ) THEN
            RAISE EXCEPTION 'schema "sales" does not exist';
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_namespace
            WHERE nspname = 'org'
        ) THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_namespace
            WHERE nspname = 'storage'
        ) THEN
            RAISE EXCEPTION 'schema "storage" does not exist';
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_namespace
            WHERE nspname = 'iam'
        ) THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF to_regclass('sales.orders') IS NULL THEN
            RAISE EXCEPTION 'table "sales.orders" does not exist; run orders migration first';
        END IF;

        IF to_regclass('storage.objects') IS NULL THEN
            RAISE EXCEPTION 'table "storage.objects" does not exist; run objects migration first';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist; run organizations migration first';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run accounts migration first';
        END IF;

        IF to_regclass('sales.order_documents') IS NOT NULL THEN
            RAISE EXCEPTION 'table "sales.order_documents" already exists; migration already applied or schema drift';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE sales.order_documents
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id uuid        NOT NULL,
    order_id        uuid        NOT NULL,
    object_id       uuid        NOT NULL,

    doc_type        text        NOT NULL DEFAULT 'attachment',
    title           text        NULL,

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NULL,
    created_by      uuid        NULL,
    updated_by      uuid        NULL,
    deleted_at      timestamptz NULL,

    CONSTRAINT fk_sales_order_documents_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_sales_order_documents_order
        FOREIGN KEY (organization_id, order_id)
            REFERENCES sales.orders (organization_id, id)
            ON DELETE CASCADE,

    CONSTRAINT fk_sales_order_documents_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_sales_order_documents_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_sales_order_documents_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_sales_order_documents_doc_type
        CHECK (doc_type IN ('tz', 'brief', 'attachment', 'contract', 'invoice', 'other')),

    CONSTRAINT chk_sales_order_documents_title_not_blank
        CHECK (title IS NULL OR btrim(title) <> ''),

    CONSTRAINT chk_sales_order_documents_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),

    CONSTRAINT chk_sales_order_documents_deleted_at_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX ux_sales_order_documents_order_object_active
    ON sales.order_documents (order_id, object_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_sales_order_documents_order_created_at
    ON sales.order_documents (order_id, created_at DESC);

CREATE INDEX ix_sales_order_documents_organization_created_at
    ON sales.order_documents (organization_id, created_at DESC);

CREATE INDEX ix_sales_order_documents_doc_type
    ON sales.order_documents (doc_type);


-- +goose Down

DROP INDEX sales.ix_sales_order_documents_doc_type;
DROP INDEX sales.ix_sales_order_documents_organization_created_at;
DROP INDEX sales.ix_sales_order_documents_order_created_at;
DROP INDEX sales.ux_sales_order_documents_order_object_active;

DROP TABLE sales.order_documents;