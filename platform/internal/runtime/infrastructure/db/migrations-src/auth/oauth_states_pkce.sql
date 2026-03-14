-- +goose Up
-- +goose StatementBegin
DO
$$
    BEGIN
        IF to_regclass('auth.oauth_states') IS NULL THEN
            RAISE EXCEPTION 'table "auth.oauth_states" does not exist; run oauth_states migration first';
        END IF;
    END
$$;
-- +goose StatementEnd

ALTER TABLE auth.oauth_states
    ADD COLUMN code_verifier text NULL;

ALTER TABLE auth.oauth_states
    ADD CONSTRAINT chk_auth_oauth_states_code_verifier_not_blank
        CHECK (code_verifier IS NULL OR btrim(code_verifier) <> '');

-- +goose Down

ALTER TABLE auth.oauth_states
    DROP CONSTRAINT chk_auth_oauth_states_code_verifier_not_blank,
    DROP COLUMN code_verifier;
