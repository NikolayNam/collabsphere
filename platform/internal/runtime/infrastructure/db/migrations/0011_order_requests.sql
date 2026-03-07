-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_namespace
                       WHERE nspname = 'sales') THEN
            RAISE EXCEPTION 'schema "sales" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_namespace
                       WHERE nspname = 'iam') THEN
            RAISE EXCEPTION 'schema "iam" does not exist';
        END IF;

        IF to_regclass('sales.orders') IS NULL THEN
            RAISE EXCEPTION 'table "sales.orders" does not exist; run orders migration first';
        END IF;

        IF to_regclass('sales.requests') IS NULL THEN
            RAISE EXCEPTION 'table "sales.requests" does not exist; run requests migration first';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist; run accounts migration first';
        END IF;

        IF to_regclass('sales.order_requests') IS NOT NULL THEN
            RAISE EXCEPTION 'table "sales.order_requests" already exists; migration already applied or schema drift';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE sales.order_requests
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    organization_id uuid        NOT NULL,
    order_id        uuid        NOT NULL,
    request_id      uuid        NOT NULL,

    role            text        NULL,
    sort_order      integer     NOT NULL DEFAULT 0,

    created_at      timestamptz NOT NULL DEFAULT now(),
    created_by      uuid        NULL,
    deleted_at      timestamptz NULL,

    CONSTRAINT chk_sales_order_requests_sort_order_nonneg
        CHECK (sort_order >= 0),

    CONSTRAINT chk_sales_order_requests_deleted_at_valid
        CHECK (deleted_at IS NULL OR deleted_at >= created_at),

    CONSTRAINT fk_sales_order_requests_order
        FOREIGN KEY (organization_id, order_id)
            REFERENCES sales.orders (organization_id, id)
            ON DELETE CASCADE,

    CONSTRAINT fk_sales_order_requests_request
        FOREIGN KEY (organization_id, request_id)
            REFERENCES sales.requests (organization_id, id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_sales_order_requests_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL
);

CREATE UNIQUE INDEX ux_sales_order_requests_request_once
    ON sales.order_requests (request_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_sales_order_requests_order_request_pair
    ON sales.order_requests (order_id, request_id)
    WHERE deleted_at IS NULL;

CREATE INDEX ix_sales_order_requests_order
    ON sales.order_requests (order_id, sort_order, created_at);

CREATE INDEX ix_sales_order_requests_request
    ON sales.order_requests (request_id);

CREATE INDEX ix_sales_order_requests_org_order
    ON sales.order_requests (organization_id, order_id);

CREATE INDEX ix_sales_order_requests_org_request
    ON sales.order_requests (organization_id, request_id);

-- +goose Down

DROP INDEX sales.ix_sales_order_requests_org_request;
DROP INDEX sales.ix_sales_order_requests_org_order;
DROP INDEX sales.ix_sales_order_requests_request;
DROP INDEX sales.ix_sales_order_requests_order;
DROP INDEX sales.ux_sales_order_requests_order_request_pair;
DROP INDEX sales.ux_sales_order_requests_request_once;

DROP TABLE sales.order_requests;