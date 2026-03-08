-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'catalog') THEN
            RAISE EXCEPTION 'schema "catalog" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'catalog'
                     AND c.relname = 'product_categories'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "catalog.product_categories" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE catalog.product_categories
(
    id         uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    parent_id  uuid         NULL,
    code       varchar(128) NOT NULL,
    name       varchar(255) NOT NULL,
    sort_order bigint       NOT NULL DEFAULT 0,
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz  NOT NULL DEFAULT now(),
    deleted_at timestamptz  NULL,

    CONSTRAINT fk_catalog_product_types_parent
        FOREIGN KEY (parent_id)
            REFERENCES catalog.product_categories (id)
            ON DELETE RESTRICT,

    CONSTRAINT uq_catalog_product_types_code
        UNIQUE (code),

    CONSTRAINT chk_catalog_product_types_code_not_blank
        CHECK (btrim(code) <> ''),

    CONSTRAINT chk_catalog_product_types_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_catalog_product_types_sort_order_nonneg
        CHECK (sort_order >= 0)
);

CREATE INDEX ix_catalog_product_types_parent_id
    ON catalog.product_categories (parent_id);

CREATE INDEX ix_catalog_product_types_sort_order
    ON catalog.product_categories (sort_order);


-- +goose Down

DROP INDEX catalog.ix_catalog_product_types_sort_order;
DROP INDEX catalog.ix_catalog_product_types_parent_id;
DROP TABLE catalog.product_categories;