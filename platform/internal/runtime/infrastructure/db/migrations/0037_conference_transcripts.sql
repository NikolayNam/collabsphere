-- +goose Up

CREATE TABLE collab.conference_transcripts
(
    conference_id       uuid PRIMARY KEY,
    transcript_text     text        NOT NULL,
    segments_json       jsonb       NOT NULL DEFAULT '[]'::jsonb,
    language_code       text        NULL,
    source_recording_id uuid        NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NULL,
    CONSTRAINT fk_collab_conference_transcripts_conference
        FOREIGN KEY (conference_id)
            REFERENCES collab.conferences (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_conference_transcripts_recording
        FOREIGN KEY (source_recording_id)
            REFERENCES collab.conference_recordings (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_collab_conference_transcripts_text_not_blank
        CHECK (btrim(transcript_text) <> ''),
    CONSTRAINT chk_collab_conference_transcripts_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

-- +goose Down

DROP TABLE collab.conference_transcripts;
