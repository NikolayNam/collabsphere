-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
            RAISE EXCEPTION 'schema "storage" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'storage'
                     AND c.relname = 'attachment_limits'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "storage.attachment_limits" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

-- Upload limits for attachments. Scope: platform (default), organization, or account.
-- Resolution order: account > organization > platform.
CREATE TABLE storage.attachment_limits
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type           text        NOT NULL,
    scope_id             uuid        NULL,
    document_limit_bytes bigint      NOT NULL,
    photo_limit_bytes    bigint      NOT NULL,
    video_limit_bytes    bigint      NOT NULL,
    total_limit_bytes    bigint      NOT NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_storage_attachment_limits_scope_type
        CHECK (scope_type IN ('platform', 'organization', 'account')),

    CONSTRAINT chk_storage_attachment_limits_platform_scope_id_null
        CHECK (scope_type != 'platform' OR scope_id IS NULL),

    CONSTRAINT chk_storage_attachment_limits_org_account_scope_id_not_null
        CHECK (scope_type = 'platform' OR scope_id IS NOT NULL),

    CONSTRAINT chk_storage_attachment_limits_positive
        CHECK (document_limit_bytes > 0 AND photo_limit_bytes > 0 AND video_limit_bytes > 0 AND total_limit_bytes > 0),

    CONSTRAINT uq_storage_attachment_limits_scope
        UNIQUE (scope_type, scope_id)
);

CREATE INDEX idx_storage_attachment_limits_scope
    ON storage.attachment_limits (scope_type, scope_id);

-- Migrate data from collab.attachment_limits if it exists
INSERT INTO storage.attachment_limits (scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes)
SELECT scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes
FROM collab.attachment_limits
ON CONFLICT (scope_type, scope_id) DO NOTHING;

-- Ensure platform default exists
INSERT INTO storage.attachment_limits (scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes)
VALUES ('platform', NULL, 10485760, 15728640, 104857600, 1073741824)
ON CONFLICT (scope_type, scope_id) DO NOTHING;

DROP TABLE IF EXISTS collab.attachment_limits;

-- +goose Down

-- Restore collab.attachment_limits (structure only, data loss on rollback)
CREATE TABLE IF NOT EXISTS collab.attachment_limits
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type           text        NOT NULL,
    scope_id             uuid        NULL,
    document_limit_bytes bigint      NOT NULL,
    photo_limit_bytes    bigint      NOT NULL,
    video_limit_bytes    bigint      NOT NULL,
    total_limit_bytes    bigint      NOT NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_attachment_limits_scope_type
        CHECK (scope_type IN ('platform', 'organization', 'account')),

    CONSTRAINT chk_attachment_limits_platform_scope_id_null
        CHECK (scope_type != 'platform' OR scope_id IS NULL),

    CONSTRAINT chk_attachment_limits_org_account_scope_id_not_null
        CHECK (scope_type = 'platform' OR scope_id IS NOT NULL),

    CONSTRAINT chk_attachment_limits_positive
        CHECK (document_limit_bytes > 0 AND photo_limit_bytes > 0 AND video_limit_bytes > 0 AND total_limit_bytes > 0),

    CONSTRAINT uq_attachment_limits_scope
        UNIQUE (scope_type, scope_id)
);

INSERT INTO collab.attachment_limits (scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes)
SELECT scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes
FROM storage.attachment_limits;

DROP TABLE storage.attachment_limits;
