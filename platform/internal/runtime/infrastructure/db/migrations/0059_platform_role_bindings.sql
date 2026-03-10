-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
        RAISE EXCEPTION 'schema "iam" does not exist';
    END IF;

    IF to_regclass('iam.accounts') IS NULL THEN
        RAISE EXCEPTION 'table "iam.accounts" does not exist';
    END IF;

    IF to_regclass('iam.platform_role_bindings') IS NOT NULL THEN
        RAISE EXCEPTION 'table "iam.platform_role_bindings" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE iam.platform_role_bindings
(
    id                    uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    account_id            uuid         NOT NULL,
    role                  varchar(64)  NOT NULL,
    granted_by_account_id uuid         NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_iam_platform_role_bindings_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_iam_platform_role_bindings_granted_by_account
        FOREIGN KEY (granted_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_iam_platform_role_bindings_account_role
        UNIQUE (account_id, role),
    CONSTRAINT chk_iam_platform_role_bindings_role_allowed
        CHECK (role IN ('platform_admin', 'support_operator', 'review_operator')),
    CONSTRAINT chk_iam_platform_role_bindings_updated_at_valid
        CHECK (updated_at >= created_at)
);

CREATE INDEX ix_iam_platform_role_bindings_account_id
    ON iam.platform_role_bindings (account_id);

CREATE INDEX ix_iam_platform_role_bindings_role
    ON iam.platform_role_bindings (role);

-- +goose Down

DROP INDEX iam.ix_iam_platform_role_bindings_role;
DROP INDEX iam.ix_iam_platform_role_bindings_account_id;
DROP TABLE iam.platform_role_bindings;
