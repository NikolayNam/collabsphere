-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF to_regclass('org.organizations') IS NULL THEN
        RAISE EXCEPTION 'table "org.organizations" does not exist';
    END IF;
    IF to_regclass('iam.memberships') IS NULL THEN
        RAISE EXCEPTION 'table "iam.memberships" does not exist';
    END IF;
    IF to_regclass('storage.objects') IS NULL THEN
        RAISE EXCEPTION 'table "storage.objects" does not exist';
    END IF;
END
$$;
-- +goose StatementEnd

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

-- +goose Down

DELETE FROM iam.memberships
WHERE id IN (
    '50000000-0000-0000-0000-000000000001',
    '50000000-0000-0000-0000-000000000002',
    '50000000-0000-0000-0000-000000000003'
);

DELETE FROM org.organizations
WHERE id IN (
    '30000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000002'
);

DELETE FROM storage.objects
WHERE id IN (
    '40000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000010',
    '40000000-0000-0000-0000-000000000011',
    '40000000-0000-0000-0000-000000000012',
    '40000000-0000-0000-0000-000000000013'
);
