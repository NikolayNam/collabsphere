-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" already exists';
        END IF;

    END
$$;
-- +goose StatementEnd

CREATE SCHEMA IF NOT EXISTS collab;

-- +goose Down

DROP SCHEMA collab;
