CREATE TABLE collab.channels
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id         uuid        NOT NULL,
    slug             text        NOT NULL,
    name             text        NOT NULL,
    description      text        NULL,
    is_default       boolean     NOT NULL DEFAULT false,
    last_message_seq bigint      NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NULL,
    deleted_at       timestamptz NULL,
    created_by       uuid        NULL,
    updated_by       uuid        NULL,
    CONSTRAINT fk_collab_channels_group
        FOREIGN KEY (group_id)
            REFERENCES iam.groups (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_channels_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_collab_channels_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_collab_channels_group_slug
        UNIQUE (group_id, slug),
    CONSTRAINT chk_collab_channels_slug_not_blank
        CHECK (btrim(slug) <> ''),
    CONSTRAINT chk_collab_channels_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT chk_collab_channels_last_message_seq_nonneg
        CHECK (last_message_seq >= 0),
    CONSTRAINT chk_collab_channels_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at),
    CONSTRAINT chk_collab_channels_deleted_at_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT chk_collab_channels_default_flag
        CHECK (is_default IN (true, false))
);

CREATE INDEX idx_collab_channels_group_id
    ON collab.channels (group_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_collab_channels_group_default
    ON collab.channels (group_id)
    WHERE is_default = true AND deleted_at IS NULL;

INSERT INTO collab.channels (group_id, slug, name, description, is_default)
SELECT g.id, 'general', 'General', 'Default channel', true
FROM iam.groups g
WHERE NOT EXISTS (
    SELECT 1
    FROM collab.channels c
    WHERE c.group_id = g.id
      AND c.is_default = true
      AND c.deleted_at IS NULL
);
