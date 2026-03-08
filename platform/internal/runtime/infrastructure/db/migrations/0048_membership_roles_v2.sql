-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_class c
                                JOIN pg_namespace n ON n.oid = c.relnamespace
                       WHERE n.nspname = 'iam'
                         AND c.relname = 'memberships'
                         AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "iam.memberships" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM iam.memberships
                   WHERE role NOT IN ('owner', 'admin', 'manager', 'member', 'viewer')) THEN
            RAISE EXCEPTION 'iam.memberships contains unsupported role values';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE iam.memberships
    DROP CONSTRAINT IF EXISTS chk_iam_memberships_role_not_blank;

ALTER TABLE iam.memberships
    DROP CONSTRAINT IF EXISTS chk_iam_memberships_role_allowed;

ALTER TABLE iam.memberships
    ADD CONSTRAINT chk_iam_memberships_role_allowed
        CHECK (role IN ('owner', 'admin', 'manager', 'member', 'viewer'));

-- +goose Down

UPDATE iam.memberships
SET role = 'member'
WHERE role IN ('admin', 'manager', 'viewer');

ALTER TABLE iam.memberships
    DROP CONSTRAINT IF EXISTS chk_iam_memberships_role_allowed;

ALTER TABLE iam.memberships
    ADD CONSTRAINT chk_iam_memberships_role_not_blank
        CHECK (btrim(role) <> '');
