-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'realtime_event_buffer'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.realtime_event_buffer" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

-- Buffer for realtime events when Redis is unavailable. Drained to Redis when it recovers.
CREATE TABLE collab.realtime_event_buffer
(
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id uuid        NOT NULL,
    event_type text        NOT NULL,
    payload    jsonb       NOT NULL,
    created_at timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_realtime_event_buffer_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_realtime_event_buffer_channel_created
    ON collab.realtime_event_buffer (channel_id, created_at ASC);

-- +goose Down

DROP TABLE collab.realtime_event_buffer;
