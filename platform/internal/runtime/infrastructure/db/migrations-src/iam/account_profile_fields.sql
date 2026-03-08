-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
        RAISE EXCEPTION 'schema "iam" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
        RAISE EXCEPTION 'schema "storage" does not exist';
    END IF;

    IF to_regclass('iam.accounts') IS NULL THEN
        RAISE EXCEPTION 'table "iam.accounts" does not exist';
    END IF;

    IF to_regclass('storage.objects') IS NULL THEN
        RAISE EXCEPTION 'table "storage.objects" does not exist';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'iam'
          AND table_name = 'accounts'
          AND column_name = 'avatar_object_id'
    ) THEN
        RAISE EXCEPTION 'column "iam.accounts.avatar_object_id" does not exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM iam.accounts AS a
        LEFT JOIN storage.objects AS so ON so.id = a.avatar_object_id
        WHERE a.avatar_object_id IS NOT NULL
          AND so.id IS NULL
    ) THEN
        RAISE EXCEPTION 'iam.accounts contains invalid avatar_object_id references';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE iam.accounts
    ADD COLUMN IF NOT EXISTS bio text NULL,
    ADD COLUMN IF NOT EXISTS phone varchar(32) NULL,
    ADD COLUMN IF NOT EXISTS locale varchar(16) NULL,
    ADD COLUMN IF NOT EXISTS timezone varchar(64) NULL,
    ADD COLUMN IF NOT EXISTS website varchar(512) NULL;

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'fk_iam_accounts_avatar_object'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT fk_iam_accounts_avatar_object
                FOREIGN KEY (avatar_object_id)
                    REFERENCES storage.objects (id)
                    ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'chk_iam_accounts_bio_not_blank'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT chk_iam_accounts_bio_not_blank
                CHECK (bio IS NULL OR btrim(bio) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'chk_iam_accounts_phone_not_blank'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT chk_iam_accounts_phone_not_blank
                CHECK (phone IS NULL OR btrim(phone) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'chk_iam_accounts_locale_not_blank'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT chk_iam_accounts_locale_not_blank
                CHECK (locale IS NULL OR btrim(locale) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'chk_iam_accounts_timezone_not_blank'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT chk_iam_accounts_timezone_not_blank
                CHECK (timezone IS NULL OR btrim(timezone) <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'iam.accounts'::regclass
          AND conname = 'chk_iam_accounts_website_not_blank'
    ) THEN
        ALTER TABLE iam.accounts
            ADD CONSTRAINT chk_iam_accounts_website_not_blank
                CHECK (website IS NULL OR btrim(website) <> '');
    END IF;
END
$$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS ix_iam_accounts_avatar_object_id
    ON iam.accounts (avatar_object_id)
    WHERE avatar_object_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS iam.ix_iam_accounts_avatar_object_id;

ALTER TABLE iam.accounts
    DROP CONSTRAINT IF EXISTS fk_iam_accounts_avatar_object,
    DROP CONSTRAINT IF EXISTS chk_iam_accounts_bio_not_blank,
    DROP CONSTRAINT IF EXISTS chk_iam_accounts_phone_not_blank,
    DROP CONSTRAINT IF EXISTS chk_iam_accounts_locale_not_blank,
    DROP CONSTRAINT IF EXISTS chk_iam_accounts_timezone_not_blank,
    DROP CONSTRAINT IF EXISTS chk_iam_accounts_website_not_blank,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS locale,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS website;
