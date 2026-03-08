-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
        RAISE EXCEPTION 'schema "org" does not exist';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'integration') THEN
        RAISE EXCEPTION 'schema "integration" does not exist';
    END IF;

    IF to_regclass('org.organization_legal_documents') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_documents" does not exist';
    END IF;

    IF to_regclass('org.organization_legal_document_analysis') IS NOT NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_document_analysis" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE org.organization_legal_document_analysis
(
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id            uuid         NOT NULL,
    organization_id        uuid         NOT NULL,
    status                 varchar(32)  NOT NULL DEFAULT 'pending',
    provider               varchar(64)  NOT NULL DEFAULT 'generic-http',
    extracted_text         text         NULL,
    summary                text         NULL,
    extracted_fields_json  jsonb        NOT NULL DEFAULT '{}'::jsonb,
    detected_document_type varchar(128) NULL,
    confidence_score       double precision NULL,
    requested_at           timestamptz  NOT NULL DEFAULT now(),
    started_at             timestamptz  NULL,
    completed_at           timestamptz  NULL,
    updated_at             timestamptz  NULL,
    last_error             text         NULL,
    CONSTRAINT uq_org_organization_legal_document_analysis_document
        UNIQUE (document_id),
    CONSTRAINT fk_org_organization_legal_document_analysis_document
        FOREIGN KEY (document_id)
            REFERENCES org.organization_legal_documents (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_org_organization_legal_document_analysis_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT chk_org_organization_legal_document_analysis_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT chk_org_organization_legal_document_analysis_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_org_organization_legal_document_analysis_summary_not_blank
        CHECK (summary IS NULL OR btrim(summary) <> ''),
    CONSTRAINT chk_org_organization_legal_document_analysis_detected_document_type_not_blank
        CHECK (detected_document_type IS NULL OR btrim(detected_document_type) <> ''),
    CONSTRAINT chk_org_organization_legal_document_analysis_last_error_not_blank
        CHECK (last_error IS NULL OR btrim(last_error) <> ''),
    CONSTRAINT chk_org_organization_legal_document_analysis_extracted_fields_object
        CHECK (jsonb_typeof(extracted_fields_json) = 'object'),
    CONSTRAINT chk_org_organization_legal_document_analysis_confidence_score_range
        CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 1)),
    CONSTRAINT chk_org_organization_legal_document_analysis_started_at_valid
        CHECK (started_at IS NULL OR started_at >= requested_at),
    CONSTRAINT chk_org_organization_legal_document_analysis_completed_at_valid
        CHECK (completed_at IS NULL OR completed_at >= requested_at),
    CONSTRAINT chk_org_organization_legal_document_analysis_updated_at_valid
        CHECK (updated_at IS NULL OR updated_at >= requested_at)
);

CREATE INDEX ix_org_organization_legal_document_analysis_status
    ON org.organization_legal_document_analysis (status, requested_at DESC);

-- +goose Down

DROP INDEX org.ix_org_organization_legal_document_analysis_status;
DROP TABLE org.organization_legal_document_analysis;
