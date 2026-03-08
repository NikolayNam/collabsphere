-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'sales') THEN
            RAISE EXCEPTION 'schema "sales" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'sales'
                     AND c.relname = 'requests'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "sales.requests" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE sales.requests
(
    id              uuid PRIMARY KEY        DEFAULT gen_random_uuid(),
    organization_id uuid           NOT NULL,
    title           varchar(255)   NOT NULL,
    description     text           NULL,
    status          varchar(64)    NOT NULL DEFAULT 'draft',
    budget_amount   numeric(14, 2) NULL,
    currency_code   varchar(3)     NULL,
    created_at      timestamptz    NOT NULL DEFAULT now(),
    updated_at      timestamptz    NOT NULL DEFAULT now(),
    deleted_at      timestamptz    NULL,

    CONSTRAINT fk_sales_requests_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_sales_requests_org_id
        UNIQUE (organization_id, id),

    CONSTRAINT chk_sales_requests_title_not_blank
        CHECK (btrim(title) <> ''),

    CONSTRAINT chk_sales_requests_budget_nonneg
        CHECK (budget_amount IS NULL OR budget_amount >= 0),

    CONSTRAINT chk_sales_requests_currency_code_len
        CHECK (currency_code IS NULL OR length(currency_code) = 3)
);

CREATE INDEX ix_sales_requests_organization_id
    ON sales.requests (organization_id);

CREATE INDEX ix_sales_requests_status
    ON sales.requests (status);

CREATE INDEX ix_sales_requests_created_at
    ON sales.requests (created_at);


-- +goose Down

DROP INDEX sales.ix_sales_requests_created_at;
DROP INDEX sales.ix_sales_requests_status;
DROP INDEX sales.ix_sales_requests_organization_id;
DROP TABLE sales.requests;