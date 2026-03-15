-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'payments') THEN
            RAISE EXCEPTION 'schema "payments" does not exist';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist';
        END IF;

        IF to_regclass('payments.payment_accounts') IS NOT NULL THEN
            RAISE EXCEPTION 'table "payments.payment_accounts" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE payments.payment_accounts
(
    id              uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id uuid         NOT NULL,
    name            varchar(255) NOT NULL,
    kind            varchar(32)  NOT NULL,
    currency_code   varchar(3)   NOT NULL DEFAULT 'RUB',
    is_active       boolean      NOT NULL DEFAULT true,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    updated_at      timestamptz  NOT NULL DEFAULT now(),
    deleted_at      timestamptz  NULL,

    CONSTRAINT fk_payments_payment_accounts_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE RESTRICT,

    CONSTRAINT chk_payments_payment_accounts_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT chk_payments_payment_accounts_kind
        CHECK (kind IN ('cash', 'bank', 'platform'))
);

CREATE INDEX ix_payments_payment_accounts_organization_id
    ON payments.payment_accounts (organization_id);

CREATE INDEX ix_payments_payment_accounts_kind
    ON payments.payment_accounts (kind);

-- +goose Down

DROP INDEX payments.ix_payments_payment_accounts_kind;
DROP INDEX payments.ix_payments_payment_accounts_organization_id;
DROP TABLE payments.payment_accounts;
