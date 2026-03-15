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

        IF to_regclass('storage.objects') IS NULL THEN
            RAISE EXCEPTION 'table "storage.objects" does not exist';
        END IF;

        IF to_regclass('iam.accounts') IS NULL THEN
            RAISE EXCEPTION 'table "iam.accounts" does not exist';
        END IF;

        IF to_regclass('payments.payment_imports') IS NOT NULL THEN
            RAISE EXCEPTION 'table "payments.payment_imports" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE payments.payment_imports
(
    id                    uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id       uuid         NOT NULL,
    source_object_id      uuid         NOT NULL,
    created_by_account_id uuid         NOT NULL,
    format_code           varchar(64)  NOT NULL,
    direction             varchar(32)  NULL,
    status                varchar(32)  NOT NULL,
    total_items           integer      NULL,
    applied_items         integer      NOT NULL DEFAULT 0,
    error_items           integer      NOT NULL DEFAULT 0,
    analysis_result       jsonb        NULL,
    started_at            timestamptz  NOT NULL DEFAULT now(),
    finished_at           timestamptz  NULL,
    created_at            timestamptz  NOT NULL DEFAULT now(),
    updated_at            timestamptz  NULL,

    CONSTRAINT fk_payments_payment_imports_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_payments_payment_imports_source_object
        FOREIGN KEY (organization_id, source_object_id)
            REFERENCES storage.objects (organization_id, id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_payments_payment_imports_created_by
        FOREIGN KEY (created_by_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE RESTRICT,

    CONSTRAINT chk_payments_payment_imports_format_code_not_blank
        CHECK (btrim(format_code) <> ''),

    CONSTRAINT chk_payments_payment_imports_direction
        CHECK (direction IS NULL OR direction IN ('incoming', 'outgoing', 'mixed')),

    CONSTRAINT chk_payments_payment_imports_status
        CHECK (status IN ('pending', 'parsing', 'parsed', 'mapping', 'applying', 'completed', 'failed', 'cancelled'))
);

CREATE INDEX ix_payments_payment_imports_organization_id
    ON payments.payment_imports (organization_id);

CREATE INDEX ix_payments_payment_imports_status
    ON payments.payment_imports (status);

CREATE INDEX ix_payments_payment_imports_created_at
    ON payments.payment_imports (created_at DESC);

-- +goose Down

DROP INDEX payments.ix_payments_payment_imports_created_at;
DROP INDEX payments.ix_payments_payment_imports_status;
DROP INDEX payments.ix_payments_payment_imports_organization_id;
DROP TABLE payments.payment_imports;
