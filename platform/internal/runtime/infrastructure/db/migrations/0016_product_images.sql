-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_namespace
            WHERE nspname = 'catalog'
        ) THEN
            RAISE EXCEPTION 'schema "catalog" does not exist';
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

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist; run organizations migration first';
        END IF;

        IF to_regclass('catalog.products') IS NULL THEN
            RAISE EXCEPTION 'table "catalog.products" does not exist; run products migration first';
        END IF;

        IF to_regclass('storage.objects') IS NULL THEN
            RAISE EXCEPTION 'table "storage.objects" does not exist; run objects migration first';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run accounts migration first';
        END IF;

        IF to_regclass('catalog.product_images') IS NOT NULL THEN
            RAISE EXCEPTION 'table "catalog.product_images" already exists; migration already applied or schema drift';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE catalog.product_images
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id uuid        NOT NULL,
    product_id      uuid        NOT NULL,
    object_id       uuid        NOT NULL,

    sort_order      integer     NOT NULL DEFAULT 0,
    is_primary      boolean     NOT NULL DEFAULT false,
    alt_text        text        NULL,

    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NULL,
    created_by      uuid        NULL,
    updated_by      uuid        NULL,
    deleted_at      timestamptz NULL,

    CONSTRAINT fk_catalog_product_images_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_catalog_product_images_product
        FOREIGN KEY (organization_id, product_id)
            REFERENCES catalog.products (organization_id, id)
            ON DELETE CASCADE,

    CONSTRAINT fk_catalog_product_images_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_catalog_product_images_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_catalog_product_images_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_catalog_product_images_sort_order_nonneg
        CHECK (sort_order >= 0),

    CONSTRAINT chk_catalog_product_images_alt_text_not_blank
        CHECK (alt_text IS NULL OR btrim(alt_text) <> ''),

    CONSTRAINT chk_catalog_product_images_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),

    CONSTRAINT chk_catalog_product_images_deleted_at_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE UNIQUE INDEX ux_catalog_product_images_product_object_active
    ON catalog.product_images (product_id, object_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_catalog_product_images_one_primary_active
    ON catalog.product_images (product_id)
    WHERE is_primary = true AND deleted_at IS NULL;

CREATE INDEX ix_catalog_product_images_product_sort
    ON catalog.product_images (product_id, sort_order, created_at);

CREATE INDEX ix_catalog_product_images_organization_product
    ON catalog.product_images (organization_id, product_id);


-- +goose Down

DROP INDEX catalog.ix_catalog_product_images_organization_product;
DROP INDEX catalog.ix_catalog_product_images_product_sort;
DROP INDEX catalog.ux_catalog_product_images_one_primary_active;
DROP INDEX catalog.ux_catalog_product_images_product_object_active;

DROP TABLE catalog.product_images;