-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE org.organization_domains
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid         NOT NULL,
    hostname        varchar(253) NOT NULL,
    kind            varchar(32)  NOT NULL,
    is_primary      boolean      NOT NULL DEFAULT false,
    verified_at     timestamptz  NULL,
    disabled_at     timestamptz  NULL,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_org_organization_domains_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT chk_org_organization_domains_hostname_not_blank
        CHECK (btrim(hostname) <> ''),
    CONSTRAINT chk_org_organization_domains_hostname_lower
        CHECK (hostname = lower(hostname)),
    CONSTRAINT chk_org_organization_domains_kind
        CHECK (kind IN ('subdomain', 'custom_domain')),
    CONSTRAINT chk_org_organization_domains_disabled_after_created
        CHECK (disabled_at IS NULL OR disabled_at >= created_at)
);

CREATE UNIQUE INDEX uq_org_organization_domains_active_hostname
    ON org.organization_domains (hostname)
    WHERE disabled_at IS NULL;

CREATE UNIQUE INDEX uq_org_organization_domains_active_primary_per_org
    ON org.organization_domains (organization_id)
    WHERE disabled_at IS NULL AND is_primary = true;

CREATE INDEX ix_org_organization_domains_organization_id
    ON org.organization_domains (organization_id, is_primary DESC, hostname)
    WHERE disabled_at IS NULL;

CREATE INDEX ix_org_organization_domains_hostname_verified
    ON org.organization_domains (hostname)
    WHERE disabled_at IS NULL AND verified_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS org.ix_org_organization_domains_hostname_verified;
DROP INDEX IF EXISTS org.ix_org_organization_domains_organization_id;
DROP INDEX IF EXISTS org.uq_org_organization_domains_active_primary_per_org;
DROP INDEX IF EXISTS org.uq_org_organization_domains_active_hostname;
DROP TABLE IF EXISTS org.organization_domains;

