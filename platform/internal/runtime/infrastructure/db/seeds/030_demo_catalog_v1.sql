-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF to_regclass('catalog.product_category_templates') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.product_category_templates" does not exist';
    END IF;
    IF to_regclass('catalog.product_categories') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.product_categories" does not exist';
    END IF;
    IF to_regclass('catalog.products') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.products" does not exist';
    END IF;
    IF to_regclass('org.organizations') IS NULL THEN
        RAISE EXCEPTION 'table "org.organizations" does not exist';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TEMP TABLE tmp_demo_organizations
(
    organization_id uuid PRIMARY KEY
) ON COMMIT DROP;

INSERT INTO tmp_demo_organizations (organization_id)
VALUES
    ('30000000-0000-0000-0000-000000000001'),
    ('30000000-0000-0000-0000-000000000002')
ON CONFLICT (organization_id) DO NOTHING;

-- +goose StatementBegin
DO
$$
DECLARE
    v_inserted_count integer;
BEGIN
    LOOP
        WITH ready AS (
            SELECT d.organization_id,
                   t.id AS template_id,
                   parent.id AS parent_id,
                   t.code,
                   t.name,
                   t.sort_order
            FROM tmp_demo_organizations d
                     JOIN catalog.product_category_templates t ON TRUE
                     LEFT JOIN catalog.product_categories existing
                               ON existing.organization_id = d.organization_id
                                   AND existing.template_id = t.id
                     LEFT JOIN catalog.product_categories parent
                               ON parent.organization_id = d.organization_id
                                   AND parent.template_id = t.parent_id
            WHERE existing.id IS NULL
              AND (t.parent_id IS NULL OR parent.id IS NOT NULL)
        ),
             inserted AS (
                 INSERT INTO catalog.product_categories (
                     organization_id,
                     parent_id,
                     template_id,
                     code,
                     name,
                     sort_order,
                     created_at,
                     updated_at,
                     deleted_at
                 )
                 SELECT r.organization_id,
                        r.parent_id,
                        r.template_id,
                        r.code,
                        r.name,
                        r.sort_order,
                        '2026-03-08T09:24:00Z'::timestamptz,
                        '2026-03-08T09:24:00Z'::timestamptz,
                        NULL
                 FROM ready r
                 ON CONFLICT (organization_id, template_id) DO NOTHING
                 RETURNING id
             )
        SELECT count(*)
        INTO v_inserted_count
        FROM inserted;

        EXIT WHEN v_inserted_count = 0;
    END LOOP;
END
$$;
-- +goose StatementEnd

INSERT INTO catalog.products (id, organization_id, product_type_id, name, description, sku, price_amount, currency_code, is_active, created_at, updated_at, deleted_at)
SELECT *
FROM (
    SELECT '65000000-0000-0000-0000-000000000001'::uuid AS id,
           '30000000-0000-0000-0000-000000000001'::uuid AS organization_id,
           c.id                                         AS product_type_id,
           'Пельмени домашние 900 г'                    AS name,
           'Флагманский SKU для теста каталога и импорта.' AS description,
           'SEV-PEL-900'                                AS sku,
           349.90::numeric(14, 2)                       AS price_amount,
           'RUB'                                        AS currency_code,
           true                                         AS is_active,
           '2026-03-08T09:25:00Z'::timestamptz          AS created_at,
           '2026-03-08T09:25:00Z'::timestamptz          AS updated_at,
           NULL::timestamptz                            AS deleted_at
    FROM catalog.product_categories c
    WHERE c.organization_id = '30000000-0000-0000-0000-000000000001'::uuid
      AND c.code = 'dumplings'

    UNION ALL

    SELECT '65000000-0000-0000-0000-000000000002'::uuid,
           '30000000-0000-0000-0000-000000000001'::uuid,
           c.id,
           'Куриный суп шоковой заморозки',
           'Позиция для проверки нескольких категорий и цен.',
           'SEV-SOUP-001',
           229.00::numeric(14, 2),
           'RUB',
           true,
           '2026-03-08T09:26:00Z'::timestamptz,
           '2026-03-08T09:26:00Z'::timestamptz,
           NULL::timestamptz
    FROM catalog.product_categories c
    WHERE c.organization_id = '30000000-0000-0000-0000-000000000001'::uuid
      AND c.code = 'soups'

    UNION ALL

    SELECT '65000000-0000-0000-0000-000000000003'::uuid,
           '30000000-0000-0000-0000-000000000002'::uuid,
           c.id,
           'Апельсиновый сок 1 л',
           'Тестовый SKU покупателя для проверки cross-org каталога.',
           'GM-JUICE-001',
           159.50::numeric(14, 2),
           'RUB',
           true,
           '2026-03-08T09:27:00Z'::timestamptz,
           '2026-03-08T09:27:00Z'::timestamptz,
           NULL::timestamptz
    FROM catalog.product_categories c
    WHERE c.organization_id = '30000000-0000-0000-0000-000000000002'::uuid
      AND c.code = 'juice'
) AS seed_products
ON CONFLICT (id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    product_type_id = EXCLUDED.product_type_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    sku = EXCLUDED.sku,
    price_amount = EXCLUDED.price_amount,
    currency_code = EXCLUDED.currency_code,
    is_active = EXCLUDED.is_active,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;

-- +goose Down

DELETE FROM catalog.products
WHERE id IN (
    '65000000-0000-0000-0000-000000000001',
    '65000000-0000-0000-0000-000000000002',
    '65000000-0000-0000-0000-000000000003'
);
