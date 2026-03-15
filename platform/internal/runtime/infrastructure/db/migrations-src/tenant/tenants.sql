-- +goose Up
CREATE SCHEMA IF NOT EXISTS tenant;

CREATE TABLE IF NOT EXISTS tenant.tenants (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS tenant.tenant_account_members (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenant.tenants(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES iam.accounts(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_tenant_member_role CHECK (role IN ('owner', 'admin', 'member')),
    CONSTRAINT uq_tenant_member_active UNIQUE (tenant_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant_id ON tenant.tenant_account_members (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_account_id ON tenant.tenant_account_members (account_id);

CREATE TABLE IF NOT EXISTS tenant.tenant_organizations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenant.tenants(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES org.organizations(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT uq_tenant_organization UNIQUE (tenant_id, organization_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_orgs_tenant_id ON tenant.tenant_organizations (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_orgs_org_id ON tenant.tenant_organizations (organization_id);

-- +goose Down
DROP TABLE IF EXISTS tenant.tenant_organizations;
DROP TABLE IF EXISTS tenant.tenant_account_members;
DROP TABLE IF EXISTS tenant.tenants;
DROP SCHEMA IF EXISTS tenant;
