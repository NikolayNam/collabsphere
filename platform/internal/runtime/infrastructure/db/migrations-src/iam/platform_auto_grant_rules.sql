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

    IF to_regclass('iam.platform_auto_grant_rules') IS NOT NULL THEN
        RAISE EXCEPTION 'table "iam.platform_auto_grant_rules" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE iam.platform_auto_grant_rules
(
    id                    uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    role                  varchar(64)  NOT NULL,
    match_type            varchar(32)  NOT NULL,
    match_value           varchar(255) NOT NULL,
    granted_by_account_id uuid         NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_iam_platform_auto_grant_rules_granted_by_account
        FOREIGN KEY (granted_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_iam_platform_auto_grant_rules_role_match
        UNIQUE (role, match_type, match_value),
    CONSTRAINT chk_iam_platform_auto_grant_rules_role_allowed
        CHECK (role IN ('platform_admin', 'support_operator', 'review_operator')),
    CONSTRAINT chk_iam_platform_auto_grant_rules_match_type_allowed
        CHECK (match_type IN ('email', 'subject')),
    CONSTRAINT chk_iam_platform_auto_grant_rules_match_value_not_blank
        CHECK (btrim(match_value) <> ''),
    CONSTRAINT chk_iam_platform_auto_grant_rules_updated_at_valid
        CHECK (updated_at >= created_at)
);

CREATE INDEX ix_iam_platform_auto_grant_rules_role
    ON iam.platform_auto_grant_rules (role);

CREATE INDEX ix_iam_platform_auto_grant_rules_match_lookup
    ON iam.platform_auto_grant_rules (match_type, match_value);

-- +goose Down

DROP INDEX iam.ix_iam_platform_auto_grant_rules_match_lookup;
DROP INDEX iam.ix_iam_platform_auto_grant_rules_role;
DROP TABLE iam.platform_auto_grant_rules;
