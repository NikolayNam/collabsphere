-- +goose Up

-- Demo credentials:
-- owner@demo.collabsphere.local   / OwnerPass123!
-- manager@demo.collabsphere.local / ManagerPass123!
-- buyer@demo.collabsphere.local   / DemoPass123!

-- +goose StatementBegin
DO
$$
BEGIN
    IF to_regclass('iam.accounts') IS NULL THEN
        RAISE EXCEPTION 'table "iam.accounts" does not exist';
    END IF;
    IF to_regclass('auth.password_credentials') IS NULL THEN
        RAISE EXCEPTION 'table "auth.password_credentials" does not exist';
    END IF;
    IF to_regclass('org.organizations') IS NULL THEN
        RAISE EXCEPTION 'table "org.organizations" does not exist';
    END IF;
    IF to_regclass('iam.memberships') IS NULL THEN
        RAISE EXCEPTION 'table "iam.memberships" does not exist';
    END IF;
    IF to_regclass('catalog.product_category_templates') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.product_category_templates" does not exist';
    END IF;
    IF to_regclass('catalog.product_categories') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.product_categories" does not exist';
    END IF;
    IF to_regclass('catalog.products') IS NULL THEN
        RAISE EXCEPTION 'table "catalog.products" does not exist';
    END IF;
    IF to_regclass('iam.groups') IS NULL THEN
        RAISE EXCEPTION 'table "iam.groups" does not exist';
    END IF;
    IF to_regclass('collab.channels') IS NULL THEN
        RAISE EXCEPTION 'table "collab.channels" does not exist';
    END IF;
    IF to_regclass('collab.messages') IS NULL THEN
        RAISE EXCEPTION 'table "collab.messages" does not exist';
    END IF;
    IF to_regclass('org.cooperation_applications') IS NULL THEN
        RAISE EXCEPTION 'table "org.cooperation_applications" does not exist';
    END IF;
    IF to_regclass('org.organization_legal_documents') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_documents" does not exist';
    END IF;
    IF to_regclass('org.organization_legal_document_analysis') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_document_analysis" does not exist';
    END IF;
    IF to_regclass('integration.organization_document_analysis_jobs') IS NULL THEN
        RAISE EXCEPTION 'table "integration.organization_document_analysis_jobs" does not exist';
    END IF;
END
$$;
-- +goose StatementEnd

INSERT INTO storage.objects (id, organization_id, bucket, object_key, file_name, content_type, size_bytes, checksum_sha256, created_at, deleted_at)
VALUES
    ('20000000-0000-0000-0000-000000000001', NULL, 'test-assets', 'demo/accounts/owner/avatar.png', 'owner-avatar.png', 'image/png', 20480, NULL, '2026-03-08T09:00:00Z', NULL),
    ('20000000-0000-0000-0000-000000000002', NULL, 'test-assets', 'demo/accounts/manager/avatar.png', 'manager-avatar.png', 'image/png', 19456, NULL, '2026-03-08T09:01:00Z', NULL),
    ('20000000-0000-0000-0000-000000000003', NULL, 'test-assets', 'demo/accounts/buyer/avatar.png', 'buyer-avatar.png', 'image/png', 18765, NULL, '2026-03-08T09:02:00Z', NULL)
ON CONFLICT (id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    bucket = EXCLUDED.bucket,
    object_key = EXCLUDED.object_key,
    file_name = EXCLUDED.file_name,
    content_type = EXCLUDED.content_type,
    size_bytes = EXCLUDED.size_bytes,
    checksum_sha256 = EXCLUDED.checksum_sha256,
    created_at = EXCLUDED.created_at,
    deleted_at = EXCLUDED.deleted_at;

INSERT INTO iam.accounts (id, email, display_name, avatar_object_id, is_active, created_at, updated_at, deleted_at, bio, phone, locale, timezone, website)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'owner@demo.collabsphere.local', 'Анна Власова', '20000000-0000-0000-0000-000000000001', true, '2026-03-08T09:05:00Z', '2026-03-08T09:05:00Z', NULL, 'Владелец demo-организации и основной контакт по сотрудничеству.', '+79990000001', 'ru-RU', 'Europe/Moscow', 'https://anna.demo.collabsphere.local'),
    ('10000000-0000-0000-0000-000000000002', 'manager@demo.collabsphere.local', 'Илья Смирнов', '20000000-0000-0000-0000-000000000002', true, '2026-03-08T09:06:00Z', '2026-03-08T09:06:00Z', NULL, 'Категорийный менеджер, отвечает за каталог и закупки.', '+79990000002', 'ru-RU', 'Europe/Moscow', 'https://manager.demo.collabsphere.local'),
    ('10000000-0000-0000-0000-000000000003', 'buyer@demo.collabsphere.local', 'Мария Кузнецова', '20000000-0000-0000-0000-000000000003', true, '2026-03-08T09:07:00Z', '2026-03-08T09:07:00Z', NULL, 'Представитель розничной сети, тестирует procurement и collab.', '+79990000003', 'ru-RU', 'Europe/Moscow', 'https://buyer.demo.collabsphere.local')
ON CONFLICT (id) DO UPDATE
SET email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    avatar_object_id = EXCLUDED.avatar_object_id,
    is_active = EXCLUDED.is_active,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at,
    bio = EXCLUDED.bio,
    phone = EXCLUDED.phone,
    locale = EXCLUDED.locale,
    timezone = EXCLUDED.timezone,
    website = EXCLUDED.website;

INSERT INTO auth.password_credentials (account_id, password_hash, created_at, updated_at)
VALUES
    ('10000000-0000-0000-0000-000000000001', '$2a$10$4ZQNkiZ2J25.VRUjqp08o.3of00RA55dwT/l9Uy4dEAPz.374Y5YC', '2026-03-08T09:05:00Z', '2026-03-08T09:05:00Z'),
    ('10000000-0000-0000-0000-000000000002', '$2a$10$6z57SLBDVdc8m/5uZeXaQOU/IxEmhEj3yNuq0Mt7f4/HBicLRcwpS', '2026-03-08T09:06:00Z', '2026-03-08T09:06:00Z'),
    ('10000000-0000-0000-0000-000000000003', '$2a$10$UTGz2ViUArFp61ff6rrlHeL4jjDwyJb.owDf8BHo/uxBurFHmQNOu', '2026-03-08T09:07:00Z', '2026-03-08T09:07:00Z')
ON CONFLICT (account_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

INSERT INTO org.organizations (id, name, slug, logo_object_id, is_active, created_at, updated_at, deleted_at, description, website, primary_email, phone, address, industry)
VALUES
    ('30000000-0000-0000-0000-000000000001', 'Северный Фудс', 'severny-foods', NULL, true, '2026-03-08T09:10:00Z', '2026-03-08T09:10:00Z', NULL, 'Поставщик замороженных полуфабрикатов и готовых блюд для HoReCa.', 'https://severny-foods.demo.local', 'partners@severny-foods.demo.local', '+74950000001', 'Москва, ул. Промышленная, 10', 'food-manufacturing'),
    ('30000000-0000-0000-0000-000000000002', 'ГородМаркет', 'gorod-market', NULL, true, '2026-03-08T09:11:00Z', '2026-03-08T09:11:00Z', NULL, 'Сеть городских дарксторов и мини-маркетов.', 'https://gorod-market.demo.local', 'buyers@gorod-market.demo.local', '+74950000002', 'Санкт-Петербург, Невский пр., 1', 'retail')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    logo_object_id = EXCLUDED.logo_object_id,
    is_active = EXCLUDED.is_active,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at,
    description = EXCLUDED.description,
    website = EXCLUDED.website,
    primary_email = EXCLUDED.primary_email,
    phone = EXCLUDED.phone,
    address = EXCLUDED.address,
    industry = EXCLUDED.industry;

INSERT INTO storage.objects (id, organization_id, bucket, object_key, file_name, content_type, size_bytes, checksum_sha256, created_at, deleted_at)
VALUES
    ('40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'test-assets', 'demo/organizations/severny-foods/logo.png', 'severny-foods-logo.png', 'image/png', 40231, NULL, '2026-03-08T09:12:00Z', NULL),
    ('40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000002', 'test-assets', 'demo/organizations/gorod-market/logo.png', 'gorod-market-logo.png', 'image/png', 31877, NULL, '2026-03-08T09:13:00Z', NULL),
    ('40000000-0000-0000-0000-000000000010', '30000000-0000-0000-0000-000000000001', 'test-assets', 'demo/organizations/severny-foods/cooperation/price-list.xlsx', 'severny-foods-price-list.xlsx', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 81234, NULL, '2026-03-08T09:14:00Z', NULL),
    ('40000000-0000-0000-0000-000000000011', '30000000-0000-0000-0000-000000000001', 'test-assets', 'demo/organizations/severny-foods/legal/inn.pdf', 'inn-certificate.pdf', 'application/pdf', 121334, NULL, '2026-03-08T09:15:00Z', NULL),
    ('40000000-0000-0000-0000-000000000012', '30000000-0000-0000-0000-000000000001', 'test-assets', 'demo/organizations/severny-foods/legal/ogrn.pdf', 'ogrn-extract.pdf', 'application/pdf', 93221, NULL, '2026-03-08T09:16:00Z', NULL),
    ('40000000-0000-0000-0000-000000000013', '30000000-0000-0000-0000-000000000002', 'test-assets', 'demo/organizations/gorod-market/legal/charter.pdf', 'company-charter.pdf', 'application/pdf', 211443, NULL, '2026-03-08T09:17:00Z', NULL)
ON CONFLICT (id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    bucket = EXCLUDED.bucket,
    object_key = EXCLUDED.object_key,
    file_name = EXCLUDED.file_name,
    content_type = EXCLUDED.content_type,
    size_bytes = EXCLUDED.size_bytes,
    checksum_sha256 = EXCLUDED.checksum_sha256,
    created_at = EXCLUDED.created_at,
    deleted_at = EXCLUDED.deleted_at;

UPDATE org.organizations
SET logo_object_id = CASE id
    WHEN '30000000-0000-0000-0000-000000000001'::uuid THEN '40000000-0000-0000-0000-000000000001'::uuid
    WHEN '30000000-0000-0000-0000-000000000002'::uuid THEN '40000000-0000-0000-0000-000000000002'::uuid
    ELSE logo_object_id
END,
updated_at = '2026-03-08T09:20:00Z'
WHERE id IN ('30000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002');

INSERT INTO iam.memberships (id, organization_id, account_id, role, is_active, created_at, updated_at, deleted_at)
VALUES
    ('50000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'owner', true, '2026-03-08T09:21:00Z', '2026-03-08T09:21:00Z', NULL),
    ('50000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'member', true, '2026-03-08T09:22:00Z', '2026-03-08T09:22:00Z', NULL),
    ('50000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 'owner', true, '2026-03-08T09:23:00Z', '2026-03-08T09:23:00Z', NULL)
ON CONFLICT (organization_id, account_id) DO UPDATE
SET role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;

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

INSERT INTO iam.groups (id, name, slug, description, is_active, created_at, updated_at, deleted_at, created_by, updated_by)
VALUES
    ('60000000-0000-0000-0000-000000000001', 'Пилотная закупка март 2026', 'pilot-procurement-march-2026', 'Группа для проверки collab, каталога и переговоров между поставщиком и покупателем.', true, '2026-03-08T09:30:00Z', '2026-03-08T09:30:00Z', NULL, '10000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at,
    created_by = EXCLUDED.created_by,
    updated_by = EXCLUDED.updated_by;

INSERT INTO iam.group_account_members (id, group_id, account_id, role, is_active, created_at, updated_at, deleted_at)
VALUES
    ('61000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'owner', true, '2026-03-08T09:31:00Z', '2026-03-08T09:31:00Z', NULL),
    ('61000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 'member', true, '2026-03-08T09:32:00Z', '2026-03-08T09:32:00Z', NULL)
ON CONFLICT (group_id, account_id) DO UPDATE
SET role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;

INSERT INTO iam.group_organization_members (id, group_id, organization_id, is_active, created_at, updated_at, deleted_at)
VALUES
    ('62000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', true, '2026-03-08T09:33:00Z', '2026-03-08T09:33:00Z', NULL),
    ('62000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', true, '2026-03-08T09:34:00Z', '2026-03-08T09:34:00Z', NULL)
ON CONFLICT (group_id, organization_id) DO UPDATE
SET is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;

INSERT INTO collab.channels (id, group_id, slug, name, description, is_default, last_message_seq, created_at, updated_at, deleted_at, created_by, updated_by)
VALUES
    ('63000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', 'general', 'General', 'Общий канал пилотной группы.', true, 1, '2026-03-08T09:35:00Z', '2026-03-08T09:45:00Z', NULL, '10000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001'),
    ('63000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000001', 'procurement', 'Procurement', 'Обсуждение ассортимента, цен и документов.', false, 2, '2026-03-08T09:36:00Z', '2026-03-08T09:47:00Z', NULL, '10000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO UPDATE
SET group_id = EXCLUDED.group_id,
    slug = EXCLUDED.slug,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_default = EXCLUDED.is_default,
    last_message_seq = EXCLUDED.last_message_seq,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at,
    created_by = EXCLUDED.created_by,
    updated_by = EXCLUDED.updated_by;

INSERT INTO collab.channel_admins (channel_id, account_id, created_at, created_by)
VALUES
    ('63000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '2026-03-08T09:37:00Z', '10000000-0000-0000-0000-000000000001')
ON CONFLICT (channel_id, account_id) DO UPDATE
SET created_at = EXCLUDED.created_at,
    created_by = EXCLUDED.created_by;

INSERT INTO collab.messages (id, channel_id, channel_seq, message_type, author_type, author_account_id, author_guest_id, body, reply_to_message_id, created_at, edited_at, deleted_at)
VALUES
    ('64000000-0000-0000-0000-000000000001', '63000000-0000-0000-0000-000000000001', 1, 'user', 'account', '10000000-0000-0000-0000-000000000001', NULL, 'Добро пожаловать в тестовую группу. Здесь можно проверить каналы, сообщения и доступы.', NULL, '2026-03-08T09:40:00Z', NULL, NULL),
    ('64000000-0000-0000-0000-000000000002', '63000000-0000-0000-0000-000000000002', 1, 'user', 'account', '10000000-0000-0000-0000-000000000003', NULL, 'Мы готовы рассмотреть прайс по полуфабрикатам и готовой еде. Нужны документы и SKU с ценами.', NULL, '2026-03-08T09:45:00Z', NULL, NULL),
    ('64000000-0000-0000-0000-000000000003', '63000000-0000-0000-0000-000000000002', 2, 'user', 'account', '10000000-0000-0000-0000-000000000002', NULL, 'Прайс и базовые юрдокументы уже загружены. Можно тестировать onboarding и document analysis.', '64000000-0000-0000-0000-000000000002', '2026-03-08T09:47:00Z', NULL, NULL)
ON CONFLICT (id) DO UPDATE
SET channel_id = EXCLUDED.channel_id,
    channel_seq = EXCLUDED.channel_seq,
    message_type = EXCLUDED.message_type,
    author_type = EXCLUDED.author_type,
    author_account_id = EXCLUDED.author_account_id,
    author_guest_id = EXCLUDED.author_guest_id,
    body = EXCLUDED.body,
    reply_to_message_id = EXCLUDED.reply_to_message_id,
    created_at = EXCLUDED.created_at,
    edited_at = EXCLUDED.edited_at,
    deleted_at = EXCLUDED.deleted_at;

UPDATE collab.channels AS c
SET last_message_seq = m.max_seq,
    updated_at = CASE c.id
        WHEN '63000000-0000-0000-0000-000000000001'::uuid THEN '2026-03-08T09:40:00Z'::timestamptz
        WHEN '63000000-0000-0000-0000-000000000002'::uuid THEN '2026-03-08T09:47:00Z'::timestamptz
        ELSE c.updated_at
    END
FROM (
    SELECT channel_id, max(channel_seq) AS max_seq
    FROM collab.messages
    WHERE channel_id IN ('63000000-0000-0000-0000-000000000001', '63000000-0000-0000-0000-000000000002')
    GROUP BY channel_id
) AS m
WHERE c.id = m.channel_id;

INSERT INTO org.cooperation_applications (
    id,
    organization_id,
    status,
    confirmation_email,
    company_name,
    represented_categories,
    minimum_order_amount,
    delivery_geography,
    sales_channels,
    storefront_url,
    contact_first_name,
    contact_last_name,
    contact_job_title,
    price_list_object_id,
    contact_email,
    contact_phone,
    partner_code,
    review_note,
    reviewer_account_id,
    submitted_at,
    reviewed_at,
    created_at,
    updated_at
)
VALUES
    (
        '66000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'submitted',
        'partners@severny-foods.demo.local',
        'ООО Северный Фудс',
        'Полуфабрикаты, готовые блюда, HoReCa',
        '15000 RUB',
        'Москва и Санкт-Петербург',
        '["horeca", "retail", "marketplace"]'::jsonb,
        'https://severny-foods.demo.local/catalog',
        'Анна',
        'Власова',
        'Коммерческий директор',
        '40000000-0000-0000-0000-000000000010',
        'partners@severny-foods.demo.local',
        '+74950000001',
        'SEV-DEMO-2026',
        NULL,
        NULL,
        '2026-03-08T10:00:00Z',
        NULL,
        '2026-03-08T09:50:00Z',
        '2026-03-08T10:00:00Z'
    ),
    (
        '66000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000002',
        'draft',
        'buyers@gorod-market.demo.local',
        'ООО ГородМаркет',
        'Розничная сеть, соки, напитки',
        '5000 RUB',
        'Санкт-Петербург',
        '["retail", "darkstore"]'::jsonb,
        'https://gorod-market.demo.local/vendors',
        'Мария',
        'Кузнецова',
        'Руководитель закупок',
        NULL,
        'buyers@gorod-market.demo.local',
        '+74950000002',
        'GM-DEMO-2026',
        NULL,
        NULL,
        NULL,
        NULL,
        '2026-03-08T09:55:00Z',
        '2026-03-08T09:55:00Z'
    )
ON CONFLICT (organization_id) DO UPDATE
SET status = EXCLUDED.status,
    confirmation_email = EXCLUDED.confirmation_email,
    company_name = EXCLUDED.company_name,
    represented_categories = EXCLUDED.represented_categories,
    minimum_order_amount = EXCLUDED.minimum_order_amount,
    delivery_geography = EXCLUDED.delivery_geography,
    sales_channels = EXCLUDED.sales_channels,
    storefront_url = EXCLUDED.storefront_url,
    contact_first_name = EXCLUDED.contact_first_name,
    contact_last_name = EXCLUDED.contact_last_name,
    contact_job_title = EXCLUDED.contact_job_title,
    price_list_object_id = EXCLUDED.price_list_object_id,
    contact_email = EXCLUDED.contact_email,
    contact_phone = EXCLUDED.contact_phone,
    partner_code = EXCLUDED.partner_code,
    review_note = EXCLUDED.review_note,
    reviewer_account_id = EXCLUDED.reviewer_account_id,
    submitted_at = EXCLUDED.submitted_at,
    reviewed_at = EXCLUDED.reviewed_at,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

INSERT INTO org.organization_legal_documents (
    id,
    organization_id,
    document_type,
    status,
    object_id,
    title,
    uploaded_by_account_id,
    reviewer_account_id,
    review_note,
    created_at,
    updated_at,
    reviewed_at,
    deleted_at
)
VALUES
    (
        '67000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'inn_certificate',
        'approved',
        '40000000-0000-0000-0000-000000000011',
        'Свидетельство ИНН',
        '10000000-0000-0000-0000-000000000001',
        '10000000-0000-0000-0000-000000000002',
        'Поля извлечены корректно, документ принят.',
        '2026-03-08T10:05:00Z',
        '2026-03-08T10:20:00Z',
        '2026-03-08T10:20:00Z',
        NULL
    ),
    (
        '67000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000001',
        'ogrn_extract',
        'pending',
        '40000000-0000-0000-0000-000000000012',
        'Выписка ОГРН',
        '10000000-0000-0000-0000-000000000001',
        NULL,
        NULL,
        '2026-03-08T10:06:00Z',
        '2026-03-08T10:06:00Z',
        NULL,
        NULL
    ),
    (
        '67000000-0000-0000-0000-000000000003',
        '30000000-0000-0000-0000-000000000002',
        'charter',
        'pending',
        '40000000-0000-0000-0000-000000000013',
        'Устав компании',
        '10000000-0000-0000-0000-000000000003',
        NULL,
        NULL,
        '2026-03-08T10:07:00Z',
        '2026-03-08T10:07:00Z',
        NULL,
        NULL
    )
ON CONFLICT (id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    document_type = EXCLUDED.document_type,
    status = EXCLUDED.status,
    object_id = EXCLUDED.object_id,
    title = EXCLUDED.title,
    uploaded_by_account_id = EXCLUDED.uploaded_by_account_id,
    reviewer_account_id = EXCLUDED.reviewer_account_id,
    review_note = EXCLUDED.review_note,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    reviewed_at = EXCLUDED.reviewed_at,
    deleted_at = EXCLUDED.deleted_at;

INSERT INTO org.organization_legal_document_analysis (
    id,
    document_id,
    organization_id,
    status,
    provider,
    extracted_text,
    summary,
    extracted_fields_json,
    detected_document_type,
    confidence_score,
    requested_at,
    started_at,
    completed_at,
    updated_at,
    last_error
)
VALUES
    (
        '68000000-0000-0000-0000-000000000001',
        '67000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'completed',
        'generic-http',
        'ИНН 7701234567. ООО Северный Фудс. Дата регистрации 12.04.2022.',
        'Распознан ИНН поставщика и реквизиты юридического лица.',
        '{"inn":"7701234567","companyName":"ООО Северный Фудс","registrationDate":"2022-04-12"}'::jsonb,
        'inn_certificate',
        0.98,
        '2026-03-08T10:05:30Z',
        '2026-03-08T10:05:35Z',
        '2026-03-08T10:05:40Z',
        '2026-03-08T10:05:40Z',
        NULL
    ),
    (
        '68000000-0000-0000-0000-000000000002',
        '67000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000001',
        'failed',
        'generic-http',
        NULL,
        NULL,
        '{}'::jsonb,
        NULL,
        NULL,
        '2026-03-08T10:06:30Z',
        '2026-03-08T10:06:35Z',
        NULL,
        '2026-03-08T10:06:40Z',
        'Provider timeout while parsing scanned PDF'
    ),
    (
        '68000000-0000-0000-0000-000000000003',
        '67000000-0000-0000-0000-000000000003',
        '30000000-0000-0000-0000-000000000002',
        'pending',
        'generic-http',
        NULL,
        NULL,
        '{}'::jsonb,
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        NULL
    )
ON CONFLICT (document_id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    status = EXCLUDED.status,
    provider = EXCLUDED.provider,
    extracted_text = EXCLUDED.extracted_text,
    summary = EXCLUDED.summary,
    extracted_fields_json = EXCLUDED.extracted_fields_json,
    detected_document_type = EXCLUDED.detected_document_type,
    confidence_score = EXCLUDED.confidence_score,
    requested_at = EXCLUDED.requested_at,
    started_at = EXCLUDED.started_at,
    completed_at = EXCLUDED.completed_at,
    updated_at = EXCLUDED.updated_at,
    last_error = EXCLUDED.last_error;

INSERT INTO integration.organization_document_analysis_jobs (
    id,
    document_id,
    status,
    provider,
    attempts,
    available_at,
    leased_until,
    completed_at,
    last_error,
    created_at,
    updated_at
)
VALUES
    (
        '69000000-0000-0000-0000-000000000001',
        '67000000-0000-0000-0000-000000000001',
        'completed',
        'generic-http',
        1,
        '2026-03-08T10:05:30Z',
        NULL,
        '2026-03-08T10:05:40Z',
        NULL,
        '2026-03-08T10:05:30Z',
        '2026-03-08T10:05:40Z'
    ),
    (
        '69000000-0000-0000-0000-000000000002',
        '67000000-0000-0000-0000-000000000002',
        'failed',
        'generic-http',
        2,
        '2026-03-08T10:10:00Z',
        NULL,
        NULL,
        'Provider timeout while parsing scanned PDF',
        '2026-03-08T10:06:30Z',
        '2026-03-08T10:06:40Z'
    ),
    (
        '69000000-0000-0000-0000-000000000003',
        '67000000-0000-0000-0000-000000000003',
        'pending',
        'generic-http',
        0,
        '2026-03-08T10:07:30Z',
        NULL,
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        '2026-03-08T10:07:30Z'
    )
ON CONFLICT (document_id) DO UPDATE
SET status = EXCLUDED.status,
    provider = EXCLUDED.provider,
    attempts = EXCLUDED.attempts,
    available_at = EXCLUDED.available_at,
    leased_until = EXCLUDED.leased_until,
    completed_at = EXCLUDED.completed_at,
    last_error = EXCLUDED.last_error,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

-- +goose Down

DELETE FROM iam.groups
WHERE id = '60000000-0000-0000-0000-000000000001';

DELETE FROM org.organizations
WHERE id IN (
    '30000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000002'
);

DELETE FROM iam.accounts
WHERE id IN (
    '10000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003'
);

DELETE FROM storage.objects
WHERE id IN (
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003',
    '40000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000010',
    '40000000-0000-0000-0000-000000000011',
    '40000000-0000-0000-0000-000000000012',
    '40000000-0000-0000-0000-000000000013'
);
