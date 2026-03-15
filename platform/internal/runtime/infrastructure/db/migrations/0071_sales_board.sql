-- +goose Up

CREATE TABLE IF NOT EXISTS sales.order_items
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid        NOT NULL,
    order_id        uuid        NOT NULL,
    category_id     uuid        NULL,
    product_name    text        NULL,
    quantity        numeric(14, 3) NULL,
    unit            varchar(32) NULL,
    note            text        NULL,
    sort_order      integer     NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NULL,
    deleted_at      timestamptz NULL,
    CONSTRAINT fk_sales_order_items_order
        FOREIGN KEY (organization_id, order_id)
            REFERENCES sales.orders (organization_id, id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_order_items_category
        FOREIGN KEY (category_id)
            REFERENCES catalog.product_categories (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_sales_order_items_target
        CHECK (category_id IS NOT NULL OR btrim(COALESCE(product_name, '')) <> '')
);

CREATE INDEX IF NOT EXISTS ix_sales_order_items_order
    ON sales.order_items (order_id, sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sales.order_offers
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        uuid        NOT NULL,
    organization_id uuid        NOT NULL,
    status          varchar(32) NOT NULL DEFAULT 'submitted',
    comment         text        NULL,
    created_by      uuid        NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NULL,
    deleted_at      timestamptz NULL,
    CONSTRAINT fk_sales_order_offers_order
        FOREIGN KEY (order_id)
            REFERENCES sales.orders (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_order_offers_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_order_offers_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS ix_sales_order_offers_order
    ON sales.order_offers (order_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sales.offer_items
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id        uuid        NOT NULL,
    organization_id uuid        NOT NULL,
    category_id     uuid        NULL,
    product_id      uuid        NULL,
    custom_title    text        NULL,
    quantity        numeric(14, 3) NULL,
    unit            varchar(32) NULL,
    price_amount    numeric(14, 2) NULL,
    currency_code   varchar(3)  NULL,
    note            text        NULL,
    sort_order      integer     NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NULL,
    deleted_at      timestamptz NULL,
    CONSTRAINT fk_sales_offer_items_offer
        FOREIGN KEY (offer_id)
            REFERENCES sales.order_offers (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_offer_items_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_offer_items_category
        FOREIGN KEY (category_id)
            REFERENCES catalog.product_categories (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_sales_offer_items_product
        FOREIGN KEY (product_id)
            REFERENCES catalog.products (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_sales_offer_items_target
        CHECK (
            category_id IS NOT NULL
            OR product_id IS NOT NULL
            OR btrim(COALESCE(custom_title, '')) <> ''
        )
);

CREATE INDEX IF NOT EXISTS ix_sales_offer_items_offer
    ON sales.offer_items (offer_id, sort_order, created_at)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sales.board_comments
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        uuid        NULL,
    offer_id        uuid        NULL,
    organization_id uuid        NOT NULL,
    account_id      uuid        NOT NULL,
    comment         text        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_sales_board_comments_order
        FOREIGN KEY (order_id)
            REFERENCES sales.orders (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_board_comments_offer
        FOREIGN KEY (offer_id)
            REFERENCES sales.order_offers (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_board_comments_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_sales_board_comments_account
        FOREIGN KEY (account_id)
            REFERENCES iam.accounts (id)
            ON DELETE CASCADE,
    CONSTRAINT chk_sales_board_comments_target
        CHECK ((order_id IS NOT NULL AND offer_id IS NULL) OR (order_id IS NULL AND offer_id IS NOT NULL)),
    CONSTRAINT chk_sales_board_comments_comment_not_blank
        CHECK (btrim(comment) <> '')
);

CREATE INDEX IF NOT EXISTS ix_sales_board_comments_order
    ON sales.board_comments (order_id, created_at)
    WHERE order_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_sales_board_comments_offer
    ON sales.board_comments (offer_id, created_at)
    WHERE offer_id IS NOT NULL;

-- +goose Down

DROP INDEX IF EXISTS sales.ix_sales_board_comments_offer;
DROP INDEX IF EXISTS sales.ix_sales_board_comments_order;
DROP TABLE IF EXISTS sales.board_comments;

DROP INDEX IF EXISTS sales.ix_sales_offer_items_offer;
DROP TABLE IF EXISTS sales.offer_items;

DROP INDEX IF EXISTS sales.ix_sales_order_offers_order;
DROP TABLE IF EXISTS sales.order_offers;

DROP INDEX IF EXISTS sales.ix_sales_order_items_order;
DROP TABLE IF EXISTS sales.order_items;
