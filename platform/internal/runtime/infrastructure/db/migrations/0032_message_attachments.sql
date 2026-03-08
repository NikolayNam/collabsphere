-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'org') THEN
            RAISE EXCEPTION 'schema "org" does not exist';
        END IF;

        IF to_regclass('org.organizations') IS NULL THEN
            RAISE EXCEPTION 'table "org.organizations" does not exist; run organizations migration first';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'message_attachments'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.message_attachments" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

CREATE TABLE collab.message_attachments
(
    message_id       uuid        NOT NULL,
    object_id        uuid        NOT NULL,
    organization_id  uuid        NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    created_by       uuid        NULL,
    PRIMARY KEY (message_id, object_id),
    CONSTRAINT fk_collab_message_attachments_message
        FOREIGN KEY (message_id)
            REFERENCES collab.messages (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_collab_message_attachments_object
        FOREIGN KEY (object_id)
            REFERENCES storage.objects (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_collab_message_attachments_org
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_collab_message_attachments_created_by
        FOREIGN KEY (created_by)
            REFERENCES iam.accounts (id)
            ON DELETE SET NULL
);

-- +goose Down

DROP TABLE collab.message_attachments;
