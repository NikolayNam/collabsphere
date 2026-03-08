-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'integration') THEN
        RAISE EXCEPTION 'schema "integration" does not exist';
    END IF;

    IF to_regclass('org.organization_legal_documents') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_documents" does not exist';
    END IF;

    IF to_regclass('integration.organization_document_analysis_jobs') IS NOT NULL THEN
        RAISE EXCEPTION 'table "integration.organization_document_analysis_jobs" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE integration.organization_document_analysis_jobs
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     uuid         NOT NULL,
    status          varchar(32)  NOT NULL DEFAULT 'pending',
    provider        varchar(64)  NOT NULL DEFAULT 'generic-http',
    attempts        integer      NOT NULL DEFAULT 0,
    available_at    timestamptz  NOT NULL DEFAULT now(),
    leased_until    timestamptz  NULL,
    completed_at    timestamptz  NULL,
    last_error      text         NULL,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NULL,
    CONSTRAINT uq_integration_organization_document_analysis_jobs_document
        UNIQUE (document_id),
    CONSTRAINT fk_integration_organization_document_analysis_jobs_document
        FOREIGN KEY (document_id)
            REFERENCES org.organization_legal_documents (id)
            ON DELETE CASCADE,
    CONSTRAINT chk_integration_organization_document_analysis_jobs_status
        CHECK (status IN ('pending', 'leased', 'completed', 'failed')),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_provider_not_blank
        CHECK (btrim(provider) <> ''),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_attempts_nonneg
        CHECK (attempts >= 0),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_last_error_not_blank
        CHECK (last_error IS NULL OR btrim(last_error) <> ''),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_lease_valid
        CHECK (leased_until IS NULL OR leased_until >= available_at),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_completed_valid
        CHECK (completed_at IS NULL OR completed_at >= created_at),
    CONSTRAINT chk_integration_organization_document_analysis_jobs_updated_valid
        CHECK (updated_at IS NULL OR updated_at >= created_at)
);

CREATE INDEX idx_integration_organization_document_analysis_jobs_status_available
    ON integration.organization_document_analysis_jobs (status, available_at);

-- +goose Down

DROP INDEX integration.idx_integration_organization_document_analysis_jobs_status_available;
DROP TABLE integration.organization_document_analysis_jobs;
