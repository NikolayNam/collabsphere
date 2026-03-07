-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('catalog.product_categories') IS NULL THEN
            RAISE EXCEPTION 'table "catalog.product_categories" does not exist';
        END IF;

        IF to_regclass('catalog.products') IS NULL THEN
            RAISE EXCEPTION 'table "catalog.products" does not exist';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist';
        END IF;

        IF to_regclass('catalog.product_category_templates') IS NOT NULL THEN
            RAISE EXCEPTION 'table "catalog.product_category_templates" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE catalog.product_categories
    RENAME TO product_category_templates;

CREATE TABLE catalog.product_categories
(
    id              uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id uuid         NOT NULL,
    parent_id       uuid         NULL,
    template_id     uuid         NULL,
    code            varchar(128) NOT NULL,
    name            varchar(255) NOT NULL,
    sort_order      bigint       NOT NULL DEFAULT 0,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),
    deleted_at      timestamptz  NULL,

    CONSTRAINT fk_catalog_product_categories_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_catalog_product_categories_organization_id_id
        UNIQUE (organization_id, id),

    CONSTRAINT fk_catalog_product_categories_parent
        FOREIGN KEY (organization_id, parent_id)
            REFERENCES catalog.product_categories (organization_id, id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_catalog_product_categories_template
        FOREIGN KEY (template_id)
            REFERENCES catalog.product_category_templates (id)
            ON DELETE SET NULL,

    CONSTRAINT uq_catalog_product_categories_organization_id_code
        UNIQUE (organization_id, code),

    CONSTRAINT uq_catalog_product_categories_organization_id_template_id
        UNIQUE (organization_id, template_id),

    CONSTRAINT chk_catalog_product_categories_code_not_blank
        CHECK (btrim(code) <> ''),

    CONSTRAINT chk_catalog_product_categories_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_catalog_product_categories_sort_order_nonneg
        CHECK (sort_order >= 0)
);

CREATE INDEX ix_catalog_product_categories_organization_id
    ON catalog.product_categories (organization_id);

CREATE INDEX ix_catalog_product_categories_organization_parent_id
    ON catalog.product_categories (organization_id, parent_id);

CREATE INDEX ix_catalog_product_categories_organization_sort_order
    ON catalog.product_categories (organization_id, sort_order);

CREATE INDEX ix_catalog_product_categories_template_id
    ON catalog.product_categories (template_id);

CREATE TEMP TABLE tmp_catalog_product_category_map
(
    organization_id uuid NOT NULL,
    template_id     uuid NOT NULL,
    category_id     uuid NOT NULL,
    PRIMARY KEY (organization_id, template_id)
) ON COMMIT DROP;

-- +goose StatementBegin
DO
$$
    DECLARE
        v_inserted_count integer;
    BEGIN
        LOOP
            WITH ready AS (SELECT o.id           AS organization_id,
                                  t.id           AS template_id,
                                  pm.category_id AS parent_category_id,
                                  t.code,
                                  t.name,
                                  t.sort_order,
                                  t.created_at,
                                  t.updated_at,
                                  t.deleted_at
                           FROM org.organizations o
                                    JOIN catalog.product_category_templates t ON TRUE
                                    LEFT JOIN tmp_catalog_product_category_map mapped
                                              ON mapped.organization_id = o.id
                                                  AND mapped.template_id = t.id
                                    LEFT JOIN tmp_catalog_product_category_map pm
                                              ON pm.organization_id = o.id
                                                  AND pm.template_id = t.parent_id
                           WHERE mapped.category_id IS NULL
                             AND (t.parent_id IS NULL OR pm.category_id IS NOT NULL)),
                 inserted AS (
                     INSERT INTO catalog.product_categories (organization_id,
                                                             parent_id,
                                                             template_id,
                                                             code,
                                                             name,
                                                             sort_order,
                                                             created_at,
                                                             updated_at,
                                                             deleted_at)
                         SELECT r.organization_id,
                                r.parent_category_id,
                                r.template_id,
                                r.code,
                                r.name,
                                r.sort_order,
                                r.created_at,
                                r.updated_at,
                                r.deleted_at
                         FROM ready r
                         RETURNING organization_id, template_id, id)
            INSERT
            INTO tmp_catalog_product_category_map (organization_id, template_id, category_id)
            SELECT i.organization_id, i.template_id, i.id
            FROM inserted i;

            GET DIAGNOSTICS v_inserted_count = ROW_COUNT;
            EXIT WHEN v_inserted_count = 0;
        END LOOP;

        IF EXISTS (SELECT 1
                   FROM org.organizations o
                            CROSS JOIN catalog.product_category_templates t
                            LEFT JOIN tmp_catalog_product_category_map mapped
                                      ON mapped.organization_id = o.id
                                          AND mapped.template_id = t.id
                   WHERE mapped.category_id IS NULL) THEN
            RAISE EXCEPTION 'failed to clone product category templates for all organizations';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE catalog.products
    DROP CONSTRAINT fk_catalog_products_product_type;

UPDATE catalog.products p
SET product_type_id = mapped.category_id
FROM tmp_catalog_product_category_map mapped
WHERE p.organization_id = mapped.organization_id
  AND p.product_type_id = mapped.template_id;

-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM catalog.products p
                            LEFT JOIN catalog.product_categories c
                                      ON c.organization_id = p.organization_id
                                          AND c.id = p.product_type_id
                   WHERE p.product_type_id IS NOT NULL
                     AND c.id IS NULL) THEN
            RAISE EXCEPTION 'failed to remap catalog.products.product_type_id to organization-scoped categories';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE catalog.products
    ADD CONSTRAINT fk_catalog_products_product_type
        FOREIGN KEY (organization_id, product_type_id)
            REFERENCES catalog.product_categories (organization_id, id)
            ON DELETE SET NULL;

-- +goose Down

-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM catalog.product_categories c
                   WHERE c.template_id IS NULL) THEN
            RAISE EXCEPTION 'cannot downgrade product categories: found organization-specific categories without template mapping';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE catalog.products
    DROP CONSTRAINT fk_catalog_products_product_type;

UPDATE catalog.products p
SET product_type_id = c.template_id
FROM catalog.product_categories c
WHERE p.organization_id = c.organization_id
  AND p.product_type_id = c.id;

DROP INDEX catalog.ix_catalog_product_categories_template_id;
DROP INDEX catalog.ix_catalog_product_categories_organization_sort_order;
DROP INDEX catalog.ix_catalog_product_categories_organization_parent_id;
DROP INDEX catalog.ix_catalog_product_categories_organization_id;
DROP TABLE catalog.product_categories;

ALTER TABLE catalog.product_category_templates
    RENAME TO product_categories;

ALTER TABLE catalog.products
    ADD CONSTRAINT fk_catalog_products_product_type
        FOREIGN KEY (product_type_id)
            REFERENCES catalog.product_categories (id)
            ON DELETE SET NULL;
