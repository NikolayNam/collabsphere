-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_class c
                                JOIN pg_namespace n ON n.oid = c.relnamespace
                       WHERE n.nspname = 'collab'
                         AND c.relname = 'conferences'
                         AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.conferences" does not exist';
        END IF;
    END
$$;
-- +goose StatementEnd

UPDATE collab.conferences
SET provider = 'mediasoup'
WHERE provider <> 'mediasoup';

ALTER TABLE collab.conferences
    ALTER COLUMN provider SET DEFAULT 'mediasoup';

ALTER TABLE collab.conferences
    DROP CONSTRAINT IF EXISTS chk_collab_conferences_provider;

ALTER TABLE collab.conferences
    ADD CONSTRAINT chk_collab_conferences_provider
        CHECK (provider IN ('mediasoup'));

-- +goose Down

ALTER TABLE collab.conferences
    DROP CONSTRAINT IF EXISTS chk_collab_conferences_provider;

ALTER TABLE collab.conferences
    ADD CONSTRAINT chk_collab_conferences_provider
        CHECK (provider IN ('jitsi', 'mediasoup'));
