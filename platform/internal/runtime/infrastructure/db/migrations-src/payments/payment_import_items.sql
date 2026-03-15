-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('payments.payment_imports') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_imports" does not exist';
        END IF;

        IF to_regclass('payments.payment_accounts') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_accounts" does not exist';
        END IF;

        IF to_regclass('payments.payment_transactions') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_transactions" does not exist';
        END IF;

        IF to_regclass('payments.payment_import_items') IS NOT NULL THEN
            RAISE EXCEPTION 'table "payments.payment_import_items" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE payments.payment_import_items
(
    id                uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    import_id         uuid         NOT NULL,
    row_index         integer      NOT NULL,
    direction         varchar(32)  NOT NULL,
    amount_cents      bigint       NOT NULL,
    currency_code     varchar(3)   NOT NULL DEFAULT 'RUB',
    occurred_at       timestamptz  NOT NULL,
    counterparty      varchar(512) NULL,
    purpose           text         NULL,
    raw_data          jsonb        NOT NULL DEFAULT '{}'::jsonb,
    mapped_account_id  uuid         NULL,
    transaction_id    uuid         NULL,
    error_code        varchar(128) NULL,
    error_message     text         NULL,
    created_at        timestamptz  NOT NULL DEFAULT now(),
    updated_at        timestamptz  NULL,

    CONSTRAINT fk_payments_payment_import_items_import
        FOREIGN KEY (import_id)
            REFERENCES payments.payment_imports (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_payments_payment_import_items_mapped_account
        FOREIGN KEY (mapped_account_id)
            REFERENCES payments.payment_accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_payments_payment_import_items_direction
        CHECK (direction IN ('incoming', 'outgoing')),

    CONSTRAINT chk_payments_payment_import_items_amount_nonzero
        CHECK (amount_cents <> 0)
);

CREATE INDEX ix_payments_payment_import_items_import_id
    ON payments.payment_import_items (import_id);

CREATE INDEX ix_payments_payment_import_items_row_index
    ON payments.payment_import_items (import_id, row_index);

-- +goose Down

DROP INDEX payments.ix_payments_payment_import_items_row_index;
DROP INDEX payments.ix_payments_payment_import_items_import_id;
DROP TABLE payments.payment_import_items;
