-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'catalog') THEN
            RAISE EXCEPTION 'schema "catalog" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'sales') THEN
            RAISE EXCEPTION 'schema "sales" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'storage') THEN
            RAISE EXCEPTION 'schema "storage" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'integration') THEN
            RAISE EXCEPTION 'schema "integration" already exists';
        END IF;

        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'auth') THEN
            RAISE EXCEPTION 'schema "auth" already exists';
        END IF;

    END
$$;
-- +goose StatementEnd

CREATE SCHEMA iam;
CREATE SCHEMA org;
CREATE SCHEMA catalog;
CREATE SCHEMA sales;
CREATE SCHEMA storage;
CREATE SCHEMA integration;
CREATE SCHEMA auth;


-- +goose Down

-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'integration'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "integration" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'storage'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "storage" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'sales'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "sales" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'catalog'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "catalog" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'org'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "org" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'iam'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "iam" is not empty';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'auth'
                     AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')) THEN
            RAISE EXCEPTION 'schema "iam" is not empty';
        END IF;
    END
$$;
-- +goose StatementEnd

DROP SCHEMA integration;
DROP SCHEMA storage;
DROP SCHEMA sales;
DROP SCHEMA catalog;
DROP SCHEMA org;
DROP SCHEMA iam;
DROP SCHEMA auth;