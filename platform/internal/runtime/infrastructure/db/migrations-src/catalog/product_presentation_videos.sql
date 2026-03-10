-- +goose Up
CREATE TABLE catalog.product_videos
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid        NOT NULL,
    product_id      uuid        NOT NULL,
    object_id       uuid        NOT NULL,
    sort_order      integer     NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    uploaded_by     uuid        NULL,
    deleted_at      timestamptz NULL,
    CONSTRAINT fk_catalog_product_videos_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_catalog_product_videos_product
        FOREIGN KEY (organization_id, product_id)
            REFERENCES catalog.products (organization_id, id)
            ON DELETE CASCADE,
    CONSTRAINT fk_catalog_product_videos_object
        FOREIGN KEY (organization_id, object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE CASCADE,
    CONSTRAINT fk_catalog_product_videos_uploaded_by
        FOREIGN KEY (uploaded_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_catalog_product_videos_product_object
        UNIQUE (product_id, object_id),
    CONSTRAINT chk_catalog_product_videos_sort_order_nonneg
        CHECK (sort_order >= 0)
);

CREATE INDEX ix_catalog_product_videos_product_id
    ON catalog.product_videos (product_id, sort_order, created_at, id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_catalog_product_videos_object_id
    ON catalog.product_videos (object_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS catalog.ix_catalog_product_videos_object_id;
DROP INDEX IF EXISTS catalog.ix_catalog_product_videos_product_id;
DROP TABLE IF EXISTS catalog.product_videos;
