-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('payments.payment_transactions') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_transactions" does not exist';
        END IF;

        IF to_regclass('payments.payment_accounts') IS NULL THEN
            RAISE EXCEPTION 'table "payments.payment_accounts" does not exist';
        END IF;

        IF to_regclass('payments.payment_transaction_lines') IS NOT NULL THEN
            RAISE EXCEPTION 'table "payments.payment_transaction_lines" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE payments.payment_transaction_lines
(
    id                uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    transaction_id     uuid         NOT NULL,
    account_id        uuid         NOT NULL,
    side              varchar(16)  NOT NULL,
    amount_cents      bigint       NOT NULL,
    currency_code     varchar(3)   NOT NULL DEFAULT 'RUB',
    created_at        timestamptz  NOT NULL DEFAULT now(),

    CONSTRAINT fk_payments_payment_transaction_lines_transaction
        FOREIGN KEY (transaction_id)
            REFERENCES payments.payment_transactions (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_payments_payment_transaction_lines_account
        FOREIGN KEY (account_id)
            REFERENCES payments.payment_accounts (id)
            ON DELETE RESTRICT,

    CONSTRAINT chk_payments_payment_transaction_lines_side
        CHECK (side IN ('debit', 'credit')),

    CONSTRAINT chk_payments_payment_transaction_lines_amount_nonzero
        CHECK (amount_cents <> 0)
);

CREATE INDEX ix_payments_payment_transaction_lines_transaction_id
    ON payments.payment_transaction_lines (transaction_id);

CREATE INDEX ix_payments_payment_transaction_lines_account_id
    ON payments.payment_transaction_lines (account_id);

-- +goose Down

DROP INDEX payments.ix_payments_payment_transaction_lines_account_id;
DROP INDEX payments.ix_payments_payment_transaction_lines_transaction_id;
DROP TABLE payments.payment_transaction_lines;
