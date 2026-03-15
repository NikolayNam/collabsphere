-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'payments') THEN
            RAISE EXCEPTION 'schema "payments" does not exist';
        END IF;

        IF to_regclass('payments.payment_accounts') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_accounts" does not exist';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist';
        END IF;

        IF to_regclass('payments.payment_transactions') IS NOT NULL THEN
            RAISE EXCEPTION 'table "payments.payment_transactions" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE payments.payment_transactions
(
    id              uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id uuid         NOT NULL,
    type            varchar(32)  NOT NULL,
    source          varchar(32)  NOT NULL,
    occurred_at     timestamptz  NOT NULL,
    description     text         NULL,
    external_ref    varchar(512) NULL,
    created_by_id   uuid         NULL,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),

    CONSTRAINT fk_payments_payment_transactions_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_payments_payment_transactions_created_by
        FOREIGN KEY (created_by_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,

    CONSTRAINT chk_payments_payment_transactions_type
        CHECK (type IN ('income', 'expense', 'transfer', 'platform_fee')),

    CONSTRAINT chk_payments_payment_transactions_source
        CHECK (source IN ('manual', 'import', 'platform', 'api'))
);

CREATE INDEX ix_payments_payment_transactions_organization_id
    ON payments.payment_transactions (organization_id);

CREATE INDEX ix_payments_payment_transactions_occurred_at
    ON payments.payment_transactions (occurred_at DESC);

CREATE INDEX ix_payments_payment_transactions_type
    ON payments.payment_transactions (type);

-- +goose Down

DROP INDEX payments.ix_payments_payment_transactions_type;
DROP INDEX payments.ix_payments_payment_transactions_occurred_at;
DROP INDEX payments.ix_payments_payment_transactions_organization_id;
DROP TABLE payments.payment_transactions;
