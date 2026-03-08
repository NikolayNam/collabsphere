CREATE TABLE collab.conferences
(
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id            uuid        NOT NULL,
    kind                  text        NOT NULL,
    status                text        NOT NULL DEFAULT 'scheduled',
    provider              text        NOT NULL DEFAULT 'jitsi',
    title                 text        NOT NULL,
    jitsi_room_name       text        NOT NULL,
    scheduled_start_at    timestamptz NULL,
    started_at            timestamptz NULL,
    ended_at              timestamptz NULL,
    recording_enabled     boolean     NOT NULL DEFAULT false,
    recording_started_at  timestamptz NULL,
    recording_stopped_at  timestamptz NULL,
    transcription_status  text        NOT NULL DEFAULT 'pending',
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NULL,
    created_by            uuid        NULL,
    updated_by            uuid        NULL,
    CONSTRAINT fk_collab_conferences_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_conferences_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_collab_conferences_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT uq_collab_conferences_room_name
        UNIQUE (jitsi_room_name),
    CONSTRAINT chk_collab_conferences_kind
        CHECK (kind IN ('audio', 'video')),
    CONSTRAINT chk_collab_conferences_status
        CHECK (status IN ('scheduled', 'live', 'ended', 'cancelled')),
    CONSTRAINT chk_collab_conferences_provider
        CHECK (provider IN ('jitsi')),
    CONSTRAINT chk_collab_conferences_title_not_blank
        CHECK (btrim(title) <> ''),
    CONSTRAINT chk_collab_conferences_room_not_blank
        CHECK (btrim(jitsi_room_name) <> ''),
    CONSTRAINT chk_collab_conferences_transcription_status
        CHECK (transcription_status IN ('pending', 'processing', 'ready', 'failed', 'disabled')),
    CONSTRAINT chk_collab_conferences_started_after_schedule
        CHECK (started_at IS NULL OR scheduled_start_at IS NULL OR started_at >= scheduled_start_at),
    CONSTRAINT chk_collab_conferences_end_after_start
        CHECK (ended_at IS NULL OR started_at IS NULL OR ended_at >= started_at),
    CONSTRAINT chk_collab_conferences_recording_start_valid
        CHECK (recording_started_at IS NULL OR started_at IS NULL OR recording_started_at >= started_at),
    CONSTRAINT chk_collab_conferences_recording_stop_valid
        CHECK (recording_stopped_at IS NULL OR recording_started_at IS NULL OR recording_stopped_at >= recording_started_at),
    CONSTRAINT chk_collab_conferences_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_collab_conferences_channel_id
    ON collab.conferences (channel_id, created_at DESC);
