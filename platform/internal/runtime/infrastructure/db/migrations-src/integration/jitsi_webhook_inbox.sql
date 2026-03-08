-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'integration') THEN
            RAISE EXCEPTION 'schema "jitsi_webhook_inbox" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'integration'
                     AND c.relname = 'jitsi_webhook_inbox'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "integration.chanel_admins" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE integration.jitsi_webhook_inbox
(
    id                uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    provider_event_id text        NOT NULL,
    event_type        text        NOT NULL,
    payload_json      jsonb       NOT NULL,
    received_at       timestamptz NOT NULL DEFAULT now(),
    processed_at      timestamptz NULL,
    error_message     text        NULL,
    CONSTRAINT uq_integration_jitsi_webhook_inbox_provider_event
        UNIQUE (provider_event_id),
    CONSTRAINT chk_integration_jitsi_webhook_inbox_event_not_blank
        CHECK (btrim(provider_event_id) <> '' AND btrim(event_type) <> '')
);

-- +goose Down

DROP TABLE integration.jitsi_webhook_inbox;
