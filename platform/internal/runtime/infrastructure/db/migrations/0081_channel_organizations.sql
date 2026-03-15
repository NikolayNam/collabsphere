-- +goose Up

-- +goose StatementBegin
DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'collab') THEN
            RAISE EXCEPTION 'schema "collab" does not exist';
        END IF;

        IF EXISTS (SELECT 1
                   FROM pg_class c
                            JOIN pg_namespace n ON n.oid = c.relnamespace
                   WHERE n.nspname = 'collab'
                     AND c.relname = 'channel_organizations'
                     AND c.relkind IN ('r', 'p')) THEN
            RAISE EXCEPTION 'table "collab.channel_organizations" already exists';
        END IF;
    END
$$;
-- +goose StatementEnd

-- Channel visibility: if a channel has rows here, only members of these organizations (within the group) can access.
-- If empty, all group members can access (current behavior).
CREATE TABLE collab.channel_organizations
(
    channel_id      uuid        NOT NULL,
    organization_id uuid        NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_channel_organizations_channel
        FOREIGN KEY (channel_id)
            REFERENCES collab.channels (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_channel_organizations_organization
        FOREIGN KEY (organization_id)
            REFERENCES org.organizations (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_channel_organizations_channel_org
        UNIQUE (channel_id, organization_id)
);

CREATE INDEX idx_channel_organizations_channel_id
    ON collab.channel_organizations (channel_id);

CREATE INDEX idx_channel_organizations_organization_id
    ON collab.channel_organizations (organization_id);

-- +goose Down

DROP TABLE collab.channel_organizations;
