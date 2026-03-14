-- +goose Up
CREATE TABLE iam.organization_access_audit_events (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES org.organizations(id) ON DELETE CASCADE,
    actor_subject_type TEXT NOT NULL,
    actor_subject_id UUID,
    actor_account_id UUID REFERENCES iam.accounts(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID,
    request_id TEXT,
    previous_state_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    next_state_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX iam_organization_access_audit_events_org_created_idx
    ON iam.organization_access_audit_events (organization_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS iam.iam_organization_access_audit_events_org_created_idx;
DROP TABLE IF EXISTS iam.organization_access_audit_events;
