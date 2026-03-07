-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'catalog') THEN
            RAISE EXCEPTION 'schema "catalog" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist';
        END IF;

        IF to_regclass('catalog.product_categories') IS NULL THEN
            RAISE EXCEPTION 'table "catalog.product_categories" does not exist';
        END IF;

        IF to_regclass('catalog.products') IS NOT NULL THEN
            RAISE EXCEPTION 'table "catalog.products" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE catalog.products
(
    id              uuid PRIMARY KEY        DEFAULT gen_random_uuid(),
    organization_id uuid           NOT NULL,
    product_type_id uuid           NULL,
    name            varchar(255)   NOT NULL,
    description     text           NULL,
    sku             varchar(128)   NULL,
    price_amount    numeric(14, 2) NULL,
    currency_code   varchar(3)     NULL,
    is_active       boolean        NOT NULL DEFAULT true,
    created_at      timestamptz    NOT NULL DEFAULT now(),
    updated_at      timestamptz    NOT NULL DEFAULT now(),
    deleted_at      timestamptz    NULL,

    CONSTRAINT fk_catalog_products_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_catalog_products_product_type
        FOREIGN KEY (product_type_id)
            REFERENCES catalog.product_categories (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_catalog_products_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_catalog_products_price_nonneg
        CHECK (price_amount IS NULL OR price_amount >= 0),

    CONSTRAINT chk_catalog_products_currency_code_format
        CHECK (currency_code IS NULL OR currency_code ~ '^[A-Z]{3}$')
);

CREATE INDEX ix_catalog_products_organization_id
    ON catalog.products (organization_id);

CREATE INDEX ix_catalog_products_product_categories_id
    ON catalog.products (product_type_id);

CREATE INDEX ix_catalog_products_name
    ON catalog.products (name);

CREATE INDEX ix_catalog_products_sku
    ON catalog.products (sku);

-- +goose Down

DROP INDEX catalog.ix_catalog_products_sku;
DROP INDEX catalog.ix_catalog_products_name;
DROP INDEX catalog.ix_catalog_products_product_categories_id;
DROP INDEX catalog.ix_catalog_products_organization_id;
DROP TABLE catalog.products;