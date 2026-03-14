-- +goose Up
CREATE TABLE iam.organization_invitations (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES org.organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    inviter_account_id UUID NOT NULL REFERENCES iam.accounts(id) ON DELETE RESTRICT,
    accepted_by_account_id UUID REFERENCES iam.accounts(id) ON DELETE SET NULL,
    accepted_at TIMESTAMPTZ,
    revoked_by_account_id UUID REFERENCES iam.accounts(id) ON DELETE SET NULL,
    revoked_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX iam_organization_invitations_open_email_idx
    ON iam.organization_invitations (organization_id, email)
    WHERE accepted_at IS NULL AND revoked_at IS NULL;

CREATE INDEX iam_organization_invitations_org_created_idx
    ON iam.organization_invitations (organization_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS iam.iam_organization_invitations_org_created_idx;
DROP INDEX IF EXISTS iam.iam_organization_invitations_open_email_idx;
DROP TABLE IF EXISTS iam.organization_invitations;
