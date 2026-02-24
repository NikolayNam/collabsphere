-- +goose Up
-- Включаем uuid генератор
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Таблица users
CREATE TABLE IF NOT EXISTS users
(
    id            uuid PRIMARY KEY      DEFAULT gen_random_uuid(),

    email         varchar(254) NOT NULL,
    password_hash text         NOT NULL,

    first_name    text         NOT NULL,
    last_name     text         NOT NULL,
    phone         varchar(16),

    is_active     boolean      NOT NULL DEFAULT true,

    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now(),
    created_by    uuid         NULL,
    updated_by    uuid         NULL
);


-- Уникальность email (глобально).
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email_lower ON users (lower(email));

-- Индексы для аудита
CREATE INDEX IF NOT EXISTS ix_users_created_by ON users (created_by);
CREATE INDEX IF NOT EXISTS ix_users_updated_by ON users (updated_by);
CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at);


-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_email_not_blank') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_email_not_blank
                    CHECK (btrim(email) <> '');
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_email_trimmed') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_email_trimmed
                    CHECK (email = btrim(email));
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_password_hash_not_blank') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_password_hash_not_blank
                    CHECK (btrim(password_hash) <> '');
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_first_name_not_blank') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_first_name_not_blank
                    CHECK (btrim(first_name) <> '');
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_last_name_not_blank') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_last_name_not_blank
                    CHECK (btrim(last_name) <> '');
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'chk_users_phone_e164') THEN
            ALTER TABLE users
                ADD CONSTRAINT chk_users_phone_e164
                    CHECK (phone IS NULL OR phone ~ '^\+[1-9][0-9]{1,14}$');
        END IF;
    END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION users_set_updated_at()
    RETURNS trigger AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION users_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
DROP FUNCTION IF EXISTS users_set_updated_at();

DROP TABLE IF EXISTS users;