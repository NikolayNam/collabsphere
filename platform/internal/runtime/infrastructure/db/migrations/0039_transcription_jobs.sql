-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'integration') THEN
            RAISE EXCEPTION 'schema "integration" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'integration'
                     AND c.relname = 'transcription_jobs'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "integration.transcription_jobs" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd


CREATE TABLE integration.transcription_jobs
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conference_id    uuid        NOT NULL,
    recording_id     uuid        NOT NULL,
    status           text        NOT NULL DEFAULT 'pending',
    provider         text        NOT NULL DEFAULT 'whisper',
    attempts         integer     NOT NULL DEFAULT 0,
    available_at     timestamptz NOT NULL DEFAULT now(),
    leased_until     timestamptz NULL,
    last_error       text        NULL,
    completed_at     timestamptz NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NULL,
    CONSTRAINT fk_integration_transcription_jobs_conference
        FOREIGN KEY (conference_id)
            REFERENCES collab.conferences (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_integration_transcription_jobs_recording
        FOREIGN KEY (recording_id)
            REFERENCES collab.conference_recordings (id)
            ON DELETE CASCADE,
    CONSTRAINT uq_integration_transcription_jobs_recording
        UNIQUE (recording_id),
    CONSTRAINT chk_integration_transcription_jobs_status
        CHECK (status IN ('pending', 'leased', 'processing', 'completed', 'failed')),
    CONSTRAINT chk_integration_transcription_jobs_attempts_nonneg
        CHECK (attempts >= 0),
    CONSTRAINT chk_integration_transcription_jobs_provider
        CHECK (provider IN ('whisper')),
    CONSTRAINT chk_integration_transcription_jobs_lease_valid
        CHECK (leased_until IS NULL OR leased_until >= available_at),
    CONSTRAINT chk_integration_transcription_jobs_completed_valid
        CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT chk_integration_transcription_jobs_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_integration_transcription_jobs_status_available
    ON integration.transcription_jobs (status, available_at);

-- +goose Down

DROP TABLE integration.transcription_jobs;
