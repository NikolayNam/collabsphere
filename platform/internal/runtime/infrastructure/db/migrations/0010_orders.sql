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
                     AND c.relname = 'orders'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "sales.orders" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE sales.orders
(
    id              uuid PRIMARY KEY        DEFAULT gen_random_uuid(),
    organization_id uuid           NOT NULL,
    number          varchar(128)   NOT NULL,
    title           varchar(255)   NOT NULL,
    description     text           NULL,
    status          varchar(64)    NOT NULL DEFAULT 'draft',
    budget_amount   numeric(14, 2) NULL,
    currency_code   varchar(3)     NULL,
    created_at      timestamptz    NOT NULL DEFAULT now(),
    updated_at      timestamptz    NOT NULL DEFAULT now(),
    deleted_at      timestamptz    NULL,

    CONSTRAINT fk_sales_orders_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_sales_orders_org_number
        UNIQUE (organization_id, number),

    CONSTRAINT uq_sales_orders_org_id
        UNIQUE (organization_id, id),

    CONSTRAINT chk_sales_orders_number_not_blank
        CHECK (btrim(number) <> ''),

    CONSTRAINT chk_sales_orders_title_not_blank
        CHECK (btrim(title) <> ''),

    CONSTRAINT chk_sales_orders_budget_nonneg
        CHECK (budget_amount IS NULL OR budget_amount >= 0),

    CONSTRAINT chk_sales_orders_currency_code_len
        CHECK (currency_code IS NULL OR length(currency_code) = 3)
);

CREATE INDEX ix_sales_orders_organization_id
    ON sales.orders (organization_id);

CREATE INDEX ix_sales_orders_status
    ON sales.orders (status);

CREATE INDEX ix_sales_orders_created_at
    ON sales.orders (created_at);


-- +goose Down

DROP INDEX sales.ix_sales_orders_created_at;
DROP INDEX sales.ix_sales_orders_status;
DROP INDEX sales.ix_sales_orders_organization_id;
DROP TABLE sales.orders;