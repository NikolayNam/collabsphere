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
    IF to_regclass('storage.objects') IS NULL THEN
        RAISE EXCEPTION 'table "storage.objects" does not exist';
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

-- +goose Down

DELETE FROM auth.password_credentials
WHERE account_id IN (
    '10000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003'
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
    '20000000-0000-0000-0000-000000000003'
);
