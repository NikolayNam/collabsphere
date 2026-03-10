-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'iam') THEN
        RAISE EXCEPTION 'schema "iam" does not exist';
    END IF;

    IF to_regclass('iam.accounts') IS NULL THEN
        RAISE EXCEPTION 'table "iam.accounts" does not exist';
    END IF;

    IF to_regclass('iam.platform_audit_events') IS NOT NULL THEN
        RAISE EXCEPTION 'table "iam.platform_audit_events" already exists';
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE iam.platform_audit_events
(
    id               uuid PRIMARY KEY      DEFAULT gen_random_uuid(),
    actor_account_id uuid         NULL,
    actor_roles      jsonb        NOT NULL DEFAULT '[]'::jsonb,
    actor_bootstrap  boolean      NOT NULL DEFAULT false,
    action           varchar(128) NOT NULL,
    target_type      varchar(64)  NOT NULL,
    target_id        varchar(255) NULL,
    status           varchar(32)  NOT NULL,
    summary          text         NULL,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT fk_iam_platform_audit_events_actor_account
        FOREIGN KEY (actor_account_id)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL,
    CONSTRAINT chk_iam_platform_audit_events_action_not_blank
        CHECK (btrim(action) <> ''),
    CONSTRAINT chk_iam_platform_audit_events_target_type_not_blank
        CHECK (btrim(target_type) <> ''),
    CONSTRAINT chk_iam_platform_audit_events_status_allowed
        CHECK (status IN ('success', 'denied', 'failed')),
    CONSTRAINT chk_iam_platform_audit_events_actor_roles_array
        CHECK (jsonb_typeof(actor_roles) = 'array')
);

CREATE INDEX ix_iam_platform_audit_events_actor_account_id
    ON iam.platform_audit_events (actor_account_id);

CREATE INDEX ix_iam_platform_audit_events_action
    ON iam.platform_audit_events (action);

CREATE INDEX ix_iam_platform_audit_events_created_at
    ON iam.platform_audit_events (created_at DESC);

CREATE INDEX ix_iam_platform_audit_events_target
    ON iam.platform_audit_events (target_type, target_id);

-- +goose Down

DROP INDEX iam.ix_iam_platform_audit_events_target;
DROP INDEX iam.ix_iam_platform_audit_events_created_at;
DROP INDEX iam.ix_iam_platform_audit_events_action;
DROP INDEX iam.ix_iam_platform_audit_events_actor_account_id;
DROP TABLE iam.platform_audit_events;
