-- +goose Up
CREATE TABLE iam.organization_access_requests (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES org.organizations(id) ON DELETE CASCADE,
    requester_account_id UUID NOT NULL REFERENCES iam.accounts(id) ON DELETE CASCADE,
    requested_role TEXT NOT NULL,
    message TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    reviewer_account_id UUID REFERENCES iam.accounts(id) ON DELETE SET NULL,
    review_note TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_organization_access_requests_status
        CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_organization_access_requests_requested_role
        CHECK (requested_role IN ('owner', 'admin', 'manager', 'member', 'viewer')),
    CONSTRAINT chk_organization_access_requests_reviewed_at_required
        CHECK ((status = 'pending' AND reviewed_at IS NULL) OR (status IN ('approved', 'rejected') AND reviewed_at IS NOT NULL))
);

CREATE UNIQUE INDEX iam_organization_access_requests_open_idx
    ON iam.organization_access_requests (organization_id, requester_account_id)
    WHERE status = 'pending';

CREATE INDEX iam_organization_access_requests_org_created_idx
    ON iam.organization_access_requests (organization_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS iam.iam_organization_access_requests_org_created_idx;
DROP INDEX IF EXISTS iam.iam_organization_access_requests_open_idx;
DROP TABLE IF EXISTS iam.organization_access_requests;
