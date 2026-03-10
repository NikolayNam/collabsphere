-- +goose Up
CREATE TABLE iam.account_videos
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid        NOT NULL,
    object_id  uuid        NOT NULL,
    sort_order integer     NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT fk_iam_account_videos_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_iam_account_videos_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_iam_account_videos_account_object
        UNIQUE (account_id, object_id),
    CONSTRAINT chk_iam_account_videos_sort_order_nonneg
        CHECK (sort_order >= 0)
);

CREATE INDEX ix_iam_account_videos_account_id
    ON iam.account_videos (account_id, sort_order, created_at, id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_iam_account_videos_object_id
    ON iam.account_videos (object_id)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS iam.ix_iam_account_videos_object_id;
DROP INDEX IF EXISTS iam.ix_iam_account_videos_account_id;
DROP TABLE IF EXISTS iam.account_videos;
