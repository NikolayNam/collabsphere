-- +goose Up

CREATE TABLE IF NOT EXISTS kyc.levels
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope          varchar(32) NOT NULL,
    code           varchar(64) NOT NULL,
    name           varchar(128) NOT NULL,
    rank           integer NOT NULL DEFAULT 1,
    is_active      boolean NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chk_kyc_levels_scope
        CHECK (scope IN ('account', 'organization')),
    CONSTRAINT chk_kyc_levels_code_not_blank
        CHECK (btrim(code) <> ''),
    CONSTRAINT chk_kyc_levels_name_not_blank
        CHECK (btrim(name) <> ''),
    CONSTRAINT chk_kyc_levels_rank_positive
        CHECK (rank > 0),
    CONSTRAINT uq_kyc_levels_scope_code
        UNIQUE (scope, code)
);

CREATE TABLE IF NOT EXISTS kyc.level_requirements
(
    level_id       uuid NOT NULL,
    document_type  varchar(64) NOT NULL,
    min_count      integer NOT NULL DEFAULT 1,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT pk_kyc_level_requirements PRIMARY KEY (level_id, document_type),
    CONSTRAINT fk_kyc_level_requirements_level
        FOREIGN KEY (level_id) REFERENCES kyc.levels (id) ON DELETE CASCADE,
    CONSTRAINT chk_kyc_level_requirements_document_type_not_blank
        CHECK (btrim(document_type) <> ''),
    CONSTRAINT chk_kyc_level_requirements_min_count_positive
        CHECK (min_count > 0)
);

ALTER TABLE kyc.account_profiles
    ADD COLUMN IF NOT EXISTS kyc_level_code varchar(64) NULL,
    ADD COLUMN IF NOT EXISTS kyc_level_assigned_at timestamptz NULL;

ALTER TABLE kyc.organization_profiles
    ADD COLUMN IF NOT EXISTS kyc_level_code varchar(64) NULL,
    ADD COLUMN IF NOT EXISTS kyc_level_assigned_at timestamptz NULL;

-- +goose Down

ALTER TABLE kyc.organization_profiles
    DROP COLUMN IF EXISTS kyc_level_assigned_at,
    DROP COLUMN IF EXISTS kyc_level_code;

ALTER TABLE kyc.account_profiles
    DROP COLUMN IF EXISTS kyc_level_assigned_at,
    DROP COLUMN IF EXISTS kyc_level_code;

DROP TABLE IF EXISTS kyc.level_requirements;
DROP TABLE IF EXISTS kyc.levels;
