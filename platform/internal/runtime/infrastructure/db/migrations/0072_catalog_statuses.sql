-- +goose Up

ALTER TABLE catalog.product_categories
    ADD COLUMN status varchar(24) NOT NULL DEFAULT 'draft';

ALTER TABLE catalog.product_categories
    ADD CONSTRAINT chk_catalog_product_categories_status
        CHECK (status IN ('draft', 'validating', 'verified', 'published', 'withdrawn', 'archived'));

CREATE INDEX ix_catalog_product_categories_organization_status
    ON catalog.product_categories (organization_id, status);

ALTER TABLE catalog.products
    ADD COLUMN status varchar(24) NOT NULL DEFAULT 'draft';

ALTER TABLE catalog.products
    ADD CONSTRAINT chk_catalog_products_status
        CHECK (status IN ('draft', 'validating', 'verified', 'published', 'withdrawn', 'archived'));

CREATE INDEX ix_catalog_products_organization_status
    ON catalog.products (organization_id, status);

ALTER TABLE org.cooperation_applications
    ADD COLUMN price_list_status varchar(24) NOT NULL DEFAULT 'draft';

ALTER TABLE org.cooperation_applications
    ADD CONSTRAINT chk_org_cooperation_applications_price_list_status
        CHECK (price_list_status IN ('draft', 'validating', 'verified', 'published', 'withdrawn', 'archived'));

CREATE INDEX ix_org_cooperation_applications_price_list_status
    ON org.cooperation_applications (price_list_status);

-- +goose Down

DROP INDEX org.ix_org_cooperation_applications_price_list_status;
ALTER TABLE org.cooperation_applications
    DROP CONSTRAINT chk_org_cooperation_applications_price_list_status;
ALTER TABLE org.cooperation_applications
    DROP COLUMN price_list_status;

DROP INDEX catalog.ix_catalog_products_organization_status;
ALTER TABLE catalog.products
    DROP CONSTRAINT chk_catalog_products_status;
ALTER TABLE catalog.products
    DROP COLUMN status;

DROP INDEX catalog.ix_catalog_product_categories_organization_status;
ALTER TABLE catalog.product_categories
    DROP CONSTRAINT chk_catalog_product_categories_status;
ALTER TABLE catalog.product_categories
    DROP COLUMN status;
