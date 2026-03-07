-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'iam'
                     AND c.relname = 'groups'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.groups" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE iam.groups
(
    id          uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    name        varchar(255) NOT NULL,
    slug        varchar(255) NOT NULL,
    description text         NULL,
    is_active   boolean      NOT NULL DEFAULT true,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz  NULL,
    created_by  uuid         NULL,
    updated_by  uuid         NULL,

    CONSTRAINT uq_iam_groups_slug
        UNIQUE (slug),

    CONSTRAINT fk_iam_groups_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_iam_groups_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_iam_groups_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_iam_groups_slug_not_blank
        CHECK (btrim(slug) <> ''),

    CONSTRAINT chk_iam_groups_description_not_blank
        CHECK (description IS NULL OR btrim(description) <> '')
);

CREATE INDEX ix_iam_groups_is_active
    ON iam.groups (is_active);

CREATE INDEX ix_iam_groups_created_at
    ON iam.groups (created_at);

CREATE INDEX ix_iam_groups_created_by
    ON iam.groups (created_by);


-- +goose Down

DROP INDEX iam.ix_iam_groups_created_by;
DROP INDEX iam.ix_iam_groups_created_at;
DROP INDEX iam.ix_iam_groups_is_active;
DROP TABLE iam.groups;
