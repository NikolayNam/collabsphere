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
    ADD COLUMN return_to text NOT NULL DEFAULT '/',
    ADD COLUMN intent varchar(32) NOT NULL DEFAULT 'login';

ALTER TABLE auth.oauth_states
    ADD CONSTRAINT chk_auth_oauth_states_return_to_not_blank
        CHECK (btrim(return_to) <> ''),
    ADD CONSTRAINT chk_auth_oauth_states_intent_valid
        CHECK (intent IN ('login', 'signup'));

-- +goose Down

ALTER TABLE auth.oauth_states
    DROP CONSTRAINT chk_auth_oauth_states_intent_valid,
    DROP CONSTRAINT chk_auth_oauth_states_return_to_not_blank,
    DROP COLUMN intent,
    DROP COLUMN return_to;
