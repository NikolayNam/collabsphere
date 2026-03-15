-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'attachment_limits'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.attachment_limits" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

-- Upload limits for chat attachments. Scope: platform (default), organization, or account.
-- Resolution order: account > organization > platform.
CREATE TABLE collab.attachment_limits
(
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type           text        NOT NULL,
    scope_id              uuid        NULL,
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

CREATE INDEX idx_attachment_limits_scope
    ON collab.attachment_limits (scope_type, scope_id);

-- Platform default: 10MB doc, 15MB photo, 100MB video, 1GB total
INSERT INTO collab.attachment_limits (scope_type, scope_id, document_limit_bytes, photo_limit_bytes, video_limit_bytes, total_limit_bytes)
VALUES ('platform', NULL, 10485760, 15728640, 104857600, 1073741824);

-- +goose Down

DROP TABLE collab.attachment_limits;
