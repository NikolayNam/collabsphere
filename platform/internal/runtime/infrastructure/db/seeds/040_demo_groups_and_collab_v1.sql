-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF to_regclass('iam.groups') IS NULL THEN
        RAISE EXCEPTION 'table "iam.groups" does not exist';
    END IF;
    IF to_regclass('iam.group_account_members') IS NULL THEN
        RAISE EXCEPTION 'table "iam.group_account_members" does not exist';
    END IF;
    IF to_regclass('iam.group_organization_members') IS NULL THEN
        RAISE EXCEPTION 'table "iam.group_organization_members" does not exist';
    END IF;
    IF to_regclass('collab.channels') IS NULL THEN
        RAISE EXCEPTION 'table "collab.channels" does not exist';
    END IF;
    IF to_regclass('collab.channel_admins') IS NULL THEN
        RAISE EXCEPTION 'table "collab.channel_admins" does not exist';
    END IF;
    IF to_regclass('collab.messages') IS NULL THEN
        RAISE EXCEPTION 'table "collab.messages" does not exist';
    END IF;
END
$$;
-- +goose StatementEnd

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

-- +goose Down

DELETE FROM iam.groups
WHERE id = '60000000-0000-0000-0000-000000000001';
