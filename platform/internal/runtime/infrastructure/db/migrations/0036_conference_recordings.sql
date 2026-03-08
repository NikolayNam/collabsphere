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
                     AND c.relname = 'conference_recordings'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.conference_recordings" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE collab.conference_recordings
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conference_id  uuid        NOT NULL,
    object_id      uuid        NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    created_by     uuid        NULL,
    duration_sec   integer     NULL,
    mime_type      text        NULL,
    CONSTRAINT fk_collab_conference_recordings_conference
        FOREIGN KEY (conference_id)
            REFERENCES collab.conferences (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_conference_recordings_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_collab_conference_recordings_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_collab_conference_recordings_object
        UNIQUE (object_id),
    CONSTRAINT chk_collab_conference_recordings_duration_nonneg
        CHECK (duration_sec IS NULL OR duration_sec >= 0)
);

-- +goose Down

DROP TABLE collab.conference_recordings;
