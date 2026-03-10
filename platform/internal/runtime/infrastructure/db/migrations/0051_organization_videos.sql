-- +goose Up
CREATE TABLE org.organization_videos
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid        NOT NULL,
    object_id       uuid        NOT NULL,
    sort_order      integer     NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    uploaded_by     uuid        NULL,
    deleted_at      timestamptz NULL,
    CONSTRAINT fk_org_organization_videos_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_org_organization_videos_object
        FOREIGN KEY (organization_id, object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE CASCADE,
    CONSTRAINT fk_org_organization_videos_uploaded_by
        FOREIGN KEY (uploaded_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_org_organization_videos_organization_object
        UNIQUE (organization_id, object_id),
    CONSTRAINT chk_org_organization_videos_sort_order_nonneg
        CHECK (sort_order >= 0)
);

CREATE INDEX ix_org_organization_videos_organization_id
    ON org.organization_videos (organization_id, sort_order, created_at, id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_org_organization_videos_object_id
    ON org.organization_videos (object_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS org.ix_org_organization_videos_object_id;
DROP INDEX IF EXISTS org.ix_org_organization_videos_organization_id;
DROP TABLE IF EXISTS org.organization_videos;
